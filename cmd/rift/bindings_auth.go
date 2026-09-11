package rift

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"rift/internal/auth"
	"rift/internal/runners"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// getEpicRunner helper gets system or local legendary runner.
func (a *App) getEpicRunner() (*runners.LegendaryRunner, error) {
	a.logInfo("[Auth] Getting Epic runner...")
	sharedLegendaryDir := filepath.Join(a.EnginesDir(), "legendary")
	legendaryPath, err := a.EngineManager.EnsureLegendary(sharedLegendaryDir)
	if err != nil {
		a.logError("[Auth] Failed to ensure legendary: %v", err)
		return nil, fmt.Errorf("failed to ensure legendary engine: %w", err)
	}
	a.logInfo("[Auth] Epic runner ready: %s", legendaryPath)
	return runners.NewLegendaryRunner(legendaryPath, filepath.Join(a.UserAuthDir(), "epic")), nil
}

// isSteamGameMacNative checks cache and store API asynchronously to see if Steam game has native Mac support.
func (a *App) isSteamGameMacNative(appID string) bool {
	if appID == "" {
		return false
	}

	steamPlatformCache.RLock()
	isMac, exists := steamPlatformCache.data[appID]
	steamPlatformCache.RUnlock()
	if exists {
		return isMac
	}

	cachePath := filepath.Join(a.RiftDir(), "steam_platforms_cache.json")
	diskCache := make(map[string]bool)
	if data, err := os.ReadFile(cachePath); err == nil {
		if err := json.Unmarshal(data, &diskCache); err == nil {
			if isMac, exists := diskCache[appID]; exists {
				steamPlatformCache.Lock()
				steamPlatformCache.data[appID] = isMac
				steamPlatformCache.Unlock()
				return isMac
			}
		}
	}

	// Non-blocking: fetch in background so main thread never chokes
	go a.fetchMacPlatformStatusAsync(appID)
	return false
}

func (a *App) fetchMacPlatformStatusAsync(appID string) {
	cachePath := filepath.Join(a.RiftDir(), "steam_platforms_cache.json")
	diskCache := make(map[string]bool)
	if data, err := os.ReadFile(cachePath); err == nil {
		json.Unmarshal(data, &diskCache)
	}

	url := fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%s&filters=platforms", appID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "RIFT-Client/1.0 (macOS)")
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		var raw map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&raw); err == nil {
			if appData, ok := raw[appID].(map[string]interface{}); ok {
				if success, ok := appData["success"].(bool); ok && success {
					if data, ok := appData["data"].(map[string]interface{}); ok {
						if platforms, ok := data["platforms"].(map[string]interface{}); ok {
							if macVal, ok := platforms["mac"].(bool); ok {
								steamPlatformCache.Lock()
								steamPlatformCache.data[appID] = macVal
								steamPlatformCache.Unlock()

								diskCache[appID] = macVal
								newData, _ := json.MarshalIndent(diskCache, "", "  ")
								os.WriteFile(cachePath, newData, 0644)
							}
						}
					}
				}
			}
		}
	}
}

func (a *App) SupabaseSignUp(email, name, password string) string {
	a.logInfo("[Auth] Supabase sign-up for email=%s name=%s", email, name)
	err := a.Supabase.SignUp(email, name, password)
	if err != nil {
		a.logError("[Auth] Supabase sign-up failed: %v", err)
		return fmt.Sprintf(`{"error": "%s"}`, strings.ReplaceAll(err.Error(), `"`, `\"`))
	}
	a.logInfo("[Auth] Supabase sign-up success: user=%s", a.Supabase.UserID)
	a.EnsureBaseDirectories()
	go a.SyncCloudPlaytimeOnLogin()
	a.emitEvent("auth_status_changed")
	a.emitEvent("library_changed")
	return fmt.Sprintf(`{"email": "%s", "userId": "%s", "name": "%s"}`, a.Supabase.UserEmail, a.Supabase.UserID, a.Supabase.UserName)
}

func (a *App) SupabaseSignIn(email, password string) string {
	a.logInfo("[Auth] Supabase sign-in for email=%s", email)
	err := a.Supabase.SignIn(email, password)
	if err != nil {
		a.logError("[Auth] Supabase sign-in failed: %v", err)
		return fmt.Sprintf(`{"error": "%s"}`, strings.ReplaceAll(err.Error(), `"`, `\"`))
	}
	a.logInfo("[Auth] Supabase sign-in success: user=%s", a.Supabase.UserID)
	a.EnsureBaseDirectories()
	go a.SyncCloudPlaytimeOnLogin()
	go a.FlushPendingTelemetry()
	a.emitEvent("auth_status_changed")
	a.emitEvent("library_changed")
	return fmt.Sprintf(`{"email": "%s", "userId": "%s", "name": "%s"}`, a.Supabase.UserEmail, a.Supabase.UserID, a.Supabase.UserName)
}

func (a *App) SupabaseSignOut() string {
	a.logInfo("[Auth] Supabase sign-out")
	a.Supabase.SignOut()
	a.logInfo("[Auth] Supabase sign-out complete")
	a.EnsureBaseDirectories()
	a.emitEvent("auth_status_changed")
	a.emitEvent("library_changed")
	return "{}"
}

func (a *App) SupabaseGetUser() string {
	if !a.Supabase.IsLoggedIn() {
		a.logInfo("[Auth] SupabaseGetUser: not logged in")
		return "{}"
	}
	a.logInfo("[Auth] SupabaseGetUser: logged in as %s", a.Supabase.UserEmail)
	return fmt.Sprintf(`{"email": "%s", "userId": "%s", "name": "%s"}`, a.Supabase.UserEmail, a.Supabase.UserID, a.Supabase.UserName)
}

func (a *App) IsLoggedIn() bool {
	return a.Supabase.IsLoggedIn()
}

func (a *App) SupabaseResetPasswordForEmail(email string) string {
	a.logInfo("[Auth] Requesting password reset OTP for email=%s", email)
	err := a.Supabase.ResetPasswordForEmail(email)
	if err != nil {
		a.logError("[Auth] Password reset request failed: %v", err)
		return fmt.Sprintf(`{"error": "%s"}`, strings.ReplaceAll(err.Error(), `"`, `\"`))
	}
	a.logInfo("[Auth] Password reset OTP sent successfully to %s", email)
	return `{"success": true}`
}

func (a *App) SupabaseVerifyResetCodeAndSetPassword(email, code, newPassword string) string {
	a.logInfo("[Auth] Verifying password reset OTP and setting new password for email=%s", email)
	err := a.Supabase.VerifyResetCodeAndSetPassword(email, code, newPassword)
	if err != nil {
		a.logError("[Auth] Verify reset code failed: %v", err)
		return fmt.Sprintf(`{"error": "%s"}`, strings.ReplaceAll(err.Error(), `"`, `\"`))
	}
	a.logInfo("[Auth] Password reset and authentication success: user=%s", a.Supabase.UserID)
	return fmt.Sprintf(`{"success": true, "email": "%s", "userId": "%s", "name": "%s"}`, a.Supabase.UserEmail, a.Supabase.UserID, a.Supabase.UserName)
}

func (a *App) SupabaseUpdatePassword(newPassword string) string {
	a.logInfo("[Auth] Updating password for active user %s", a.Supabase.UserEmail)
	err := a.Supabase.UpdatePassword(newPassword)
	if err != nil {
		a.logError("[Auth] Update password failed: %v", err)
		return fmt.Sprintf(`{"error": "%s"}`, strings.ReplaceAll(err.Error(), `"`, `\"`))
	}
	a.logInfo("[Auth] Password updated successfully for %s", a.Supabase.UserEmail)
	return `{"success": true}`
}

// loadRiftConfig reads ~/.rift/rift.json for runtime paths (gptk_lib, dxvk_path etc.)
func (a *App) loadRiftConfig() map[string]string {
	data, err := os.ReadFile(filepath.Join(a.RiftDir(), "rift.json"))
	if err != nil {
		return nil
	}
	var cfg struct {
		Runtimes map[string]string `json:"runtimes"`
	}
	json.Unmarshal(data, &cfg)
	return cfg.Runtimes
}

// StartEpicAuth handles the seamless Epic Games login flow.
func (a *App) StartEpicAuth() string {
	a.logInfo("Launching isolated Epic Auth webview popup...")

	runner, err := a.getEpicRunner()
	if err != nil {
		a.logInfo("Failed to setup Epic runner: %v\n", err)
		return "Failed"
	}

	token, err := auth.SpawnIsolatedEpicAuthPopup()
	if err != nil {
		a.logInfo("Epic login popup failed: %v. Attempting default config import...\n", err)

		if errImport := runner.ImportExistingAuth(); errImport == nil {
			if username, errUser := runner.GetUsernameFast(); errUser == nil {
				return username
			}
		}
		return "Failed"
	}

	_, err = runner.AuthWithCode(token)
	if err != nil {
		return "Failed"
	}

	username, err := runner.GetUsernameFast()
	if err != nil {
		username, err = runner.GetUsername()
		if err != nil {
			username = "Account Connected"
		}
	}

	a.TrackUserEvent("store_connected", "", "", "epic", map[string]interface{}{"username": username})
	return username
}

// steamGlobalAuthDir returns the unified user auth directory for all Steam components.
// Both SteamCMD and the full Steam client share this same directory for config/userdata wormholes.
func (a *App) steamGlobalAuthDir() string {
	return filepath.Join(a.UserAuthDir(), "steam")
}

// LogoutEpic purges all Epic session data
func (a *App) LogoutEpic() bool {
	os.RemoveAll(filepath.Join(a.UserAuthDir(), "epic"))
	a.TrackUserEvent("store_disconnected", "", "", "epic", nil)
	a.emitEvent("library_changed")
	return true
}

// OpenSteamChat opens the locked-down native Steam browser to steamcommunity.com/chat.
func (a *App) OpenSteamChat() string {
	a.logInfo("[Steam Browser] Opening Steam Chat...")
	if err := auth.SpawnSteamBrowser("https://steamcommunity.com/chat"); err != nil {
		a.logError("[Steam Browser] Failed: %v", err)
		return "Failed"
	}
	return "Opened"
}
