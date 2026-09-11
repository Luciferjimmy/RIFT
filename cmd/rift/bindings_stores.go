package rift

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"rift/internal/engine"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type StoreConfig struct {
	Ready     bool   `json:"ready"`
	SetupDate string `json:"setupDate"`
	Username  string `json:"username"`
}

// IsStoreReady checks if the user is authenticated with Steam.
func (a *App) IsStoreReady(platform string) bool {
	if platform != "steam" {
		return true
	}
	
	storeJSON := filepath.Join(a.StoresDir(), "steam", "store.json")
	data, err := os.ReadFile(storeJSON)
	if err != nil {
		return false
	}
	var conf StoreConfig
	if err := json.Unmarshal(data, &conf); err != nil {
		return false
	}
	return conf.Ready
}

// StartSteamAuth handles login via native SteamCMD.
func (a *App) StartSteamAuth(username, password, guardCode string) map[string]interface{} {
	a.logInfo("[Auth] Starting Steam authentication for %s...", username)
	a.emitEvent("steam_loading", true)
	defer a.emitEvent("steam_loading", false)

	steamcmdDir := a.SteamCMDDir()
	if err := engine.EnsureSteamCMD(steamcmdDir); err != nil {
		a.logError("[Auth] Failed to prepare SteamCMD: %v", err)
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to prepare SteamCMD: %v", err),
		}
	}

	var res *engine.AuthResult
	var err error

	if guardCode != "" {
		res, err = engine.AuthenticateWith2FA(steamcmdDir, username, password, guardCode)
	} else {
		res, err = engine.Authenticate(steamcmdDir, username, password)
	}

	if err != nil {
		a.logError("[Auth] Steam authentication error: %v", err)
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	if res.NeedsSteamGuard {
		a.logInfo("[Auth] Steam Guard code required for %s", username)
		errMsg := "Steam Guard code required"
		if res.ErrorMessage != "" {
			errMsg = res.ErrorMessage
		}
		return map[string]interface{}{
			"success":    false,
			"needsGuard": true,
			"error":      errMsg,
		}
	}

	if !res.Success {
		a.logError("[Auth] Steam login failed for %s: %s", username, res.ErrorMessage)
		return map[string]interface{}{
			"success": false,
			"error":   res.ErrorMessage,
		}
	}

	// Save store.json
	steamStoreDir := filepath.Join(a.StoresDir(), "steam")
	os.MkdirAll(steamStoreDir, 0755)
	conf := StoreConfig{
		Ready:     true,
		SetupDate: fmt.Sprintf("%v", os.Getenv("USER")),
		Username:  username,
	}
	b, _ := json.Marshal(conf)
	os.WriteFile(filepath.Join(steamStoreDir, "store.json"), b, 0644)

	a.logInfo("[Auth] Steam login successful for %s!", username)
	a.TrackUserEvent("store_connected", "", "", "steam", map[string]interface{}{"username": username})
	a.emitEvent("steam_logged_in", username)
	a.SyncLibrary()

	return map[string]interface{}{
		"success":  true,
		"username": username,
	}
}

// LogoutSteam removes saved Steam credentials, clears cached SteamCMD tokens, and resets store.json.
func (a *App) LogoutSteam() string {
	a.logInfo("[Auth] Logging out of Steam...")

	// 1. Remove RIFT's own store.json
	steamStoreDir := filepath.Join(a.StoresDir(), "steam")
	os.Remove(filepath.Join(steamStoreDir, "store.json"))

	// 2. Clear SteamCMD cached credentials so next login is forced fresh
	home, _ := os.UserHomeDir()
	steamConfigDir := filepath.Join(home, "Library", "Application Support", "Steam", "config")
	// Remove config.vdf (contains login tokens) and loginusers.vdf
	os.Remove(filepath.Join(steamConfigDir, "config.vdf"))
	os.Remove(filepath.Join(steamConfigDir, "loginusers.vdf"))

	// 3. Clear sentry files (Steam Guard device tokens)
	steamBaseDir := filepath.Join(home, "Library", "Application Support", "Steam")
	if entries, err := os.ReadDir(steamBaseDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && len(entry.Name()) > 4 && entry.Name()[:4] == "ssfn" {
				os.Remove(filepath.Join(steamBaseDir, entry.Name()))
			}
		}
	}

	a.logInfo("[Auth] Steam credentials cleared completely.")
	a.TrackUserEvent("store_disconnected", "", "", "steam", nil)
	a.emitEvent("steam_logged_out", nil)
	a.SyncLibrary()
	return "Logged out"
}

// OpenSteamBrowser opens store.steampowered.com in the default macOS system browser.
func (a *App) OpenSteamBrowser() string {
	a.logInfo("[Store] Opening Steam web store in default browser...")
	runtime.BrowserOpenURL(a.ctx, "https://store.steampowered.com")
	return "Opened"
}

// OpenSteamGame opens a specific game's store page in the system browser.
func (a *App) OpenSteamGame(appID string) string {
	if appID == "" || appID == "0" {
		return "Invalid App ID"
	}
	a.logInfo("[Store] Opening Steam store page for app %s...", appID)
	runtime.BrowserOpenURL(a.ctx, fmt.Sprintf("https://store.steampowered.com/app/%s", appID))
	return "Opened"
}

// KillSteamBrowser placeholder (no Wine browser running anymore).
func (a *App) KillSteamBrowser() string {
	return "No background Steam process"
}

// SetupEpicStore placeholder.
func (a *App) SetupEpicStore() string {
	return "Not implemented"
}

// OpenEpicStore placeholder.
func (a *App) OpenEpicStore() string {
	return "Not implemented"
}