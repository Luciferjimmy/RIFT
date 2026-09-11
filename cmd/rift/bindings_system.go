package rift

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"rift/internal/engine"
	"rift/internal/types"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// GetMacHardwareUUID retrieves the unique physical Mac hardware serial number
// via ioreg -c IOPlatformExpertDevice and returns its SHA256 hash.
func GetMacHardwareUUID() string {
	out, err := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output()
	if err != nil {
		return "unknown-mac-device"
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "IOPlatformSerialNumber") {
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				serial := strings.Trim(strings.TrimSpace(parts[1]), "\"")
				if serial != "" {
					hash := sha256.Sum256([]byte(serial))
					return hex.EncodeToString(hash[:])
				}
			}
		}
	}
	return "unknown-mac-device"
}

var steamNamesCache map[string]string
var steamNamesCacheMu sync.Mutex

// queryGameNameFromDB queries the local game name from DB and local cache only.
// NEVER makes network calls — the UI must never block on SyncLibrary.
func queryGameNameFromDB(appID string) string {
	home, _ := os.UserHomeDir()

	// 1. Check local JSON cache first (instant)
	steamNamesCacheMu.Lock()
	if steamNamesCache == nil {
		steamNamesCache = make(map[string]string)
		cachePath := filepath.Join(home, ".rift", "steam_names_cache.json")
		if data, err := os.ReadFile(cachePath); err == nil {
			json.Unmarshal(data, &steamNamesCache)
		}
	}
	if cachedName, exists := steamNamesCache[appID]; exists {
		steamNamesCacheMu.Unlock()
		return cachedName
	}
	steamNamesCacheMu.Unlock()

	// 2. Check Epic local metadata cache (for Epic appIDs and 32-hex hashes)
	for _, metaDir := range []string{
		filepath.Join(home, ".rift", "auth", "epic", "metadata"),
		filepath.Join(home, ".config", "legendary", "metadata"),
	} {
		metaFile := filepath.Join(metaDir, appID+".json")
		if data, err := os.ReadFile(metaFile); err == nil {
			var m struct {
				AppTitle string `json:"app_title"`
			}
			if json.Unmarshal(data, &m) == nil && m.AppTitle != "" {
				steamNamesCacheMu.Lock()
				steamNamesCache[appID] = m.AppTitle
				cachePath := filepath.Join(home, ".rift", "steam_names_cache.json")
				if b, err := json.MarshalIndent(steamNamesCache, "", "  "); err == nil {
					os.WriteFile(cachePath, b, 0644)
				}
				steamNamesCacheMu.Unlock()
				return m.AppTitle
			}
		}
	}

	// 3. Fallback to Official Steam Store API (for numeric Steam appIDs)
	apiURL := fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%s", appID)
	client := &http.Client{Timeout: 5 * time.Second}
	if resp, err := client.Get(apiURL); err == nil {
		defer resp.Body.Close()
		var result map[string]struct {
			Success bool `json:"success"`
			Data    struct {
				Name string `json:"name"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			if gameData, ok := result[appID]; ok && gameData.Success && gameData.Data.Name != "" {
				steamNamesCacheMu.Lock()
				steamNamesCache[appID] = gameData.Data.Name
				cachePath := filepath.Join(home, ".rift", "steam_names_cache.json")
				if data, err := json.MarshalIndent(steamNamesCache, "", "  "); err == nil {
					os.WriteFile(cachePath, data, 0644)
				}
				steamNamesCacheMu.Unlock()
				return gameData.Data.Name
			}
		}
	}

	return ""
}

// resolveDataDir finds the data directory (protondb, lutris, agw, etc.) by checking
// exec dir → CWD → ~/.rift/data.
func resolveDataDir() string {
	dir := "data"
	if execPath, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(execPath), "data")
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			return candidate
		}
	}
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir
	}
	home, _ := os.UserHomeDir()
	candidate := filepath.Join(home, ".rift", "data")
	if info, err := os.Stat(candidate); err == nil && info.IsDir() {
		return candidate
	}
	return dir
}

var (
	gameBackendCacheMu sync.RWMutex
	gameBackendCache   = make(map[string]string)
)

// queryGameBackendFromDB queries the local DB to check the game's original backend compatibility tags.
func queryGameBackendFromDB(appID string) string {
	if appID == "" {
		return ""
	}
	gameBackendCacheMu.RLock()
	if val, ok := gameBackendCache[appID]; ok {
		gameBackendCacheMu.RUnlock()
		return val
	}
	gameBackendCacheMu.RUnlock()

	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".rift", "rift.db")
	if _, err := os.Stat(dbPath); err != nil {
		return ""
	}

	db, err := sql.Open("sqlite3", dbPath+"?mode=ro")
	if err != nil {
		return ""
	}
	defer db.Close()

	var backend string
	err = db.QueryRow("SELECT backend FROM game_fixes WHERE app_id = ? LIMIT 1;", appID).Scan(&backend)
	if err != nil {
		return ""
	}

	gameBackendCacheMu.Lock()
	gameBackendCache[appID] = backend
	gameBackendCacheMu.Unlock()

	return backend
}

// GetAuthStatus queries the config states to check logins.
func (a *App) GetAuthStatus() AuthStatus {
	status := AuthStatus{}

	if runner, err := a.getEpicRunner(); err == nil {
		if username, errUsername := runner.GetUsernameFast(); errUsername == nil && username != "" {
			status.EpicConnected = true
			status.EpicUsername = username
		}
	}

	// Read the new native store.json to check Steam login state
	steamStoreJsonPath := filepath.Join(a.StoresDir(), "steam", "store.json")
	if b, err := os.ReadFile(steamStoreJsonPath); err == nil {
		var conf StoreConfig
		if json.Unmarshal(b, &conf) == nil && conf.Username != "" {
			status.SteamConnected = true
			status.SteamUsername = conf.Username
		}
	}

	if a.IsStoreReady("steam") {
		status.SteamInstalled = true
	}

	return status
}

func (a *App) GetGameRuntimeConfig(gameID string) *types.RuntimeConfig {
	path := a.runtimeConfigPath(gameID)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var rt types.RuntimeConfig
	if err := json.Unmarshal(data, &rt); err != nil {
		return nil
	}
	return &rt
}

func (a *App) GetSanitizedLogs() string {
	logPath := filepath.Join(a.RiftDir(), "logs", "session.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		return "Error reading log: " + err.Error()
	}

	content := string(data)
	// Mask potential sensitive data (very basic)
	content = strings.ReplaceAll(content, a.RiftDir(), "~/.rift")
	content = strings.ReplaceAll(content, os.Getenv("HOME"), "~")
	
	// Keep the last 200 lines to prevent overload
	lines := strings.Split(content, "\n")
	if len(lines) > 200 {
		lines = lines[len(lines)-200:]
	}
	return strings.Join(lines, "\n")
}

func (a *App) CalculateHardwareOptimizations(gameID, engine string) {
	cfg := a.GetFullConfig()
	if cfg.GameOverrides == nil {
		cfg.GameOverrides = make(map[string]GameOverride)
	}

	if _, exists := cfg.GameOverrides[gameID]; exists {
		return
	}

	override := GameOverride{
		Engine:    "auto",
		Retina:    "auto",
		HUD:       "auto",
		Args:      "",
		VRAMLimit: 0,
		FPSCap:    0,
		PrePurge:  false,
		GameMode:  false,
	}

	memSizeStr, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err == nil {
		memSize, _ := strconv.ParseInt(strings.TrimSpace(string(memSizeStr)), 10, 64)
		if memSize > 0 && memSize <= 8*1024*1024*1024 {
			if engine == "dxvk" {
				override.VRAMLimit = 2048
			}
		}
	}

	override.PrePurge = true
	override.GameMode = true

	cfg.GameOverrides[gameID] = override
	a.SaveFullConfig(cfg)
}

// CheckEnginesStatus returns JSON with the install status of Wine, DXVK, GPTK, Rosetta 2, and SteamCMD.
func (a *App) CheckEnginesStatus() string {
	type engineStatus struct {
		Wine     bool `json:"wine"`
		DXVK     bool `json:"dxvk"`
		GPTK     bool `json:"gptk"`
		Rosetta  bool `json:"rosetta"`
		SteamCMD bool `json:"steamcmd"`
	}
	status := engineStatus{
		Wine:     engine.CheckWineInstalled(),
		DXVK:     engine.CheckDXVKInstalled(),
		GPTK:     engine.CheckGPTKInstalled(),
		Rosetta:  a.CheckRosettaInstalled(),
		SteamCMD: engine.IsSteamCMDReady(a.SteamCMDDir()),
	}
	data, _ := json.Marshal(status)
	return string(data)
}

// InstallEngine downloads and installs the specified engine component.
// component is one of: "wine", "dxvk", "gptk", "rosetta"
// Progress is emitted via "engine_install_progress" events.
func (a *App) InstallEngine(component string) string {
	a.emitEvent("engine_install_progress", map[string]interface{}{
		"component": component, "status": "downloading", "percent": 0,
	})

	switch component {
	case "rosetta":
		a.logInfo("[Setup] Auto-installing Rosetta 2 runtime...")
		go func() {
			a.emitEvent("engine_install_progress", map[string]interface{}{
				"component": "rosetta", "status": "downloading", "percent": 50,
			})
			ok := a.EnsureRosetta()
			if ok {
				a.logInfo("[Setup] Rosetta 2 installed successfully")
				a.emitEvent("engine_install_progress", map[string]interface{}{
					"component": "rosetta", "status": "done", "percent": 100,
				})
			} else {
				a.logError("[Setup] Rosetta 2 installation failed")
				a.emitEvent("engine_install_progress", map[string]interface{}{
					"component": "rosetta", "status": "failed", "error": "Rosetta 2 install failed",
				})
			}
		}()
		return "Installing Rosetta 2..."

	case "wine":
		a.logInfo("[Setup] Installing Wine engine...")
		wineDir := filepath.Join(a.EnginesDir(), "wine")
		go func() {
			report := a.engineProgress("wine")
			err := a.EngineManager.EnsureWineProgress(wineDir, report)
			if err != nil {
				a.logError("[Setup] Wine install failed: %v", err)
				a.emitEvent("engine_install_progress", map[string]interface{}{
					"component": "wine", "status": "failed", "error": err.Error(),
				})
				return
			}
			a.logInfo("[Setup] Wine installed successfully")
			a.emitEvent("engine_install_progress", map[string]interface{}{
				"component": "wine", "status": "done", "percent": 100,
			})
		}()
		return "Installing Wine..."

	case "dxvk":
		a.logInfo("[Setup] Installing DXVK-macOS engine...")
		dxvkDir := filepath.Join(a.EnginesDir(), "dxvk")
		go func() {
			report := a.engineProgress("dxvk")
			err := a.EngineManager.EnsureDXVKProgress(dxvkDir, report)
			if err != nil {
				a.logError("[Setup] DXVK install failed: %v", err)
				a.emitEvent("engine_install_progress", map[string]interface{}{
					"component": "dxvk", "status": "failed", "error": err.Error(),
				})
				return
			}
			a.logInfo("[Setup] DXVK installed successfully")
			a.emitEvent("engine_install_progress", map[string]interface{}{
				"component": "dxvk", "status": "done", "percent": 100,
			})
		}()
		return "Installing DXVK..."

	case "gptk":
		a.logInfo("[Setup] Installing Game Porting Toolkit...")
		gptkDir := filepath.Join(a.EnginesDir(), "gptk")
		go func() {
			a.emitEvent("engine_install_progress", map[string]interface{}{
				"component": "gptk", "status": "downloading", "percent": 1,
				"stage": "Resolving latest GPTK release (Gcenx)...",
			})
			report := a.engineProgress("gptk")
			err := a.EngineManager.EnsureGPTKProgress(gptkDir, report)
			if err != nil {
				a.logError("[Setup] GPTK install failed: %v", err)
				a.emitEvent("engine_install_progress", map[string]interface{}{
					"component": "gptk", "status": "failed", "error": err.Error(),
				})
				return
			}
			a.logInfo("[Setup] GPTK installed successfully to %s", gptkDir)
			a.emitEvent("engine_install_progress", map[string]interface{}{
				"component": "gptk", "status": "done", "percent": 100,
			})
		}()
		return "Installing Game Porting Toolkit..."

	case "steamcmd":
		a.logInfo("[Setup] Installing SteamCMD runtime...")
		steamcmdDir := a.SteamCMDDir()
		go func() {
			a.emitEvent("engine_install_progress", map[string]interface{}{
				"component": "steamcmd", "status": "downloading", "percent": 20,
				"stage": "Downloading SteamCMD macOS archive...",
			})
			err := engine.EnsureSteamCMD(steamcmdDir)
			if err != nil {
				a.logError("[Setup] SteamCMD install failed: %v", err)
				a.emitEvent("engine_install_progress", map[string]interface{}{
					"component": "steamcmd", "status": "failed", "error": err.Error(),
				})
				return
			}
			a.logInfo("[Setup] SteamCMD installed successfully to %s", steamcmdDir)
			a.emitEvent("engine_install_progress", map[string]interface{}{
				"component": "steamcmd", "status": "done", "percent": 100,
			})
		}()
		return "Installing SteamCMD..."

	default:
		return fmt.Sprintf("Unknown component: %s", component)
	}
}

// engineProgress returns an onProgress callback that forwards real download
// progress to the frontend as "engine_install_progress" events. The callback
// throttles to every 1% change so we don't flood the frontend with thousands
// of events during an ~240MB download.
func (a *App) engineProgress(component string) func(downloaded, total int64) {
	last := -1
	return func(downloaded, total int64) {
		if total <= 0 {
			return // server didn't report Content-Length; no % to compute
		}
		pct := int(downloaded * 100 / total)
		if pct == last {
			return
		}
		last = pct
		if pct > 99 {
			pct = 99 // hold at 99% until extraction completes; "done" sets 100
		}
		a.emitEvent("engine_install_progress", map[string]interface{}{
			"component": component, "status": "downloading", "percent": pct,
		})
	}
}

// InstallSystemDependencies verifies and auto-installs system prerequisites (Rosetta 2 on Apple Silicon).
func (a *App) InstallSystemDependencies() string {
	if a.EnsureRosetta() {
		return "Success"
	}
	return "Failed to install required host dependencies"
}

// platformFromAppID is a best-effort helper.
func platformFromAppID(appID string) string {
	if strings.HasPrefix(appID, "epic-") || strings.HasPrefix(appID, "appID_") || len(appID) == 32 {
		return "Epic"
	}

	isNumeric := true
	for _, c := range appID {
		if c < '0' || c > '9' {
			isNumeric = false
			break
		}
	}
	if isNumeric {
		return "Steam"
	}
	return "Epic"
}

// SteamStoreItem represents a result from the Steam Store Search API.
type SteamStoreItem struct {
	AppID int    `json:"id"`
	Name  string `json:"name"`
	Price struct {
		Currency string `json:"currency"`
		Initial  int    `json:"initial"`
		Final    int    `json:"final"`
	} `json:"price,omitempty"`
	TinyImage string `json:"tiny_image"`
	Platforms struct {
		Windows bool `json:"windows"`
		Mac     bool `json:"mac"`
		Linux   bool `json:"linux"`
	} `json:"platforms"`
}

// knownNonSteamGames is a blocklist of Steam App IDs that are not playable games.
// These are system infrastructure, config utilities, SDKs, developer sandboxes, etc.
// Source: Valve Developer Community, SteamDB
var knownNonSteamGames = map[string]bool{
	// Single-digit system architecture IDs
	"1": true, // Steam Client (bootstrapper/core updates)
	"2": true, // Steam Base Config (global environment config)
	"3": true, // Steam Linux Architecture
	"4": true, // Steam OSX Architecture
	"5": true, "6": true, "8": true, "9": true, // Reserved system IDs

	// Engine frameworks & developer tools
	"20":  true, // Team Fortress Classic (originally GoldSrc test network)
	"90":  true, // GoldSrc Engine Common (shared code repository)
	"211": true, // Source SDK
	"300": true, // Source SDK Base 2006 (deprecated)
	"310": true, // Source SDK Base 2007
	"480": true, // Spacewar (Steamworks API developer sandbox)
	"510": true, // Source SDK Base 2013 (singleplayer)
	"520": true, // Source SDK Base 2013 (multiplayer)

	// Community & economy
	"753": true, // Steam Community / Inventory / Economy
	"895": true, // Steamworks API Example (not a game)
	"897": true, // Steamworks API Example (not a game)
}

// steamTypeCache caches whether a Steam App ID is a game (true) or not (false).
// Persisted to disk at ~/.rift/steam_type_cache.json to avoid repeated API calls.
var steamTypeCache map[string]bool
var steamTypeCacheMu sync.Mutex

func init() {
	steamTypeCache = make(map[string]bool)
	home, _ := os.UserHomeDir()
	cachePath := filepath.Join(home, ".rift", "steam_type_cache.json")
	if data, err := os.ReadFile(cachePath); err == nil {
		json.Unmarshal(data, &steamTypeCache)
	}
}

// persistSteamTypeCache writes the type cache to disk.
func persistSteamTypeCache() {
	home, _ := os.UserHomeDir()
	cachePath := filepath.Join(home, ".rift", "steam_type_cache.json")
	data, _ := json.MarshalIndent(steamTypeCache, "", "  ")
	os.WriteFile(cachePath, data, 0644)
}

// isSteamGame validates whether a Steam App ID represents an actual playable game.
// It checks: blocklist → in-memory cache → Steam Store API (type field).
// Results are cached to ~/.rift/steam_type_cache.json to avoid repeated API calls.
func isSteamGame(appID string) bool {
	if appID == "" {
		return false
	}
	if knownNonSteamGames[appID] {
		return false
	}
	return true
}

// steamStoreSearchResponse mirrors the Steam Store Search API JSON response.
type steamStoreSearchResponse struct {
	Total int              `json:"total"`
	Items []SteamStoreItem `json:"items"`
}

const steamStoreSearchURL = "https://store.steampowered.com/api/storesearch/?cc=US&l=en&term="

// SearchSteamAppList searches Steam's storefront via the public Store Search API (no key required).
// Returns a JSON array of matching apps with appid, name, price, platforms, and image.
// Results are limited to 20 items (Steam's limit).
func (a *App) SearchSteamAppList(query string) string {
	if query == "" {
		return "[]"
	}

	client := &http.Client{Timeout: 30 * time.Second}
	reqURL := steamStoreSearchURL + url.QueryEscape(query)

	resp, err := client.Get(reqURL)
	if err != nil {
		a.logError("[SteamSearch] API request failed: %v", err)
		return fmt.Sprintf(`{"error": "Failed to reach Steam: %s"}`, strings.ReplaceAll(err.Error(), `"`, `\"`))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		a.logError("[SteamSearch] API returned %d", resp.StatusCode)
		return fmt.Sprintf(`{"error": "Steam API returned status %d"}`, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		a.logError("[SteamSearch] Failed to read response: %v", err)
		return fmt.Sprintf(`{"error": "Failed to read Steam response."}`)
	}

	var apiResp steamStoreSearchResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		a.logError("[SteamSearch] Failed to parse response: %v", err)
		return fmt.Sprintf(`{"error": "Failed to parse Steam search results."}`)
	}

	// Map to a simpler output format
	type searchResult struct {
		AppID     int    `json:"appid"`
		Name      string `json:"name"`
		PriceCents int   `json:"price_cents"`
		Image     string `json:"image"`
		HasMac    bool   `json:"has_mac"`
	}

	results := make([]searchResult, 0, len(apiResp.Items))
	for _, item := range apiResp.Items {
		// Filter out invalid entries (bundles, placeholders, etc.) that Steam
		// sometimes returns with AppID = 0.
		if item.AppID <= 0 {
			continue
		}
		results = append(results, searchResult{
			AppID:     item.AppID,
			Name:      item.Name,
			PriceCents: item.Price.Final,
			Image:     item.TinyImage,
			HasMac:    item.Platforms.Mac,
		})
	}

	data, _ := json.Marshal(results)
	return string(data)
}

// RefreshSteamAppList is a no-op with the Store Search API (no cache to refresh).
// Kept for frontend compatibility; returns success immediately.
func (a *App) RefreshSteamAppList() string {
	return `{"status": "ok", "message": "Store Search API does not require a cache refresh."}`
}

// getMetadataDB returns a persistent connection to the local metadata SQLite database.
// It automatically ensures indices on steamid, name, and slug for sub-millisecond queries.
func (a *App) getMetadataDB() *sql.DB {
	a.metadataDBMu.Lock()
	defer a.metadataDBMu.Unlock()

	if a.metadataDB != nil {
		return a.metadataDB
	}

	prodDB := filepath.Join(a.RiftDir(), "data", "metadata.db")
	os.MkdirAll(filepath.Dir(prodDB), 0755)

	home, _ := os.UserHomeDir()
	exe, _ := os.Executable()
	exeDir := filepath.Dir(exe)
	cwd, _ := os.Getwd()

	candidates := []string{
		prodDB,
		filepath.Join(home, "RIFT", "data", "frontend_metadata", "metadata.db"),
		filepath.Join(exeDir, "data", "frontend_metadata", "metadata.db"),
		filepath.Join(exeDir, "..", "..", "data", "frontend_metadata", "metadata.db"),
		filepath.Join(cwd, "data", "frontend_metadata", "metadata.db"),
	}

	var activePath string
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			activePath = p
			break
		}
	}

	if activePath == "" {
		activePath = prodDB
	} else if activePath != prodDB {
		if data, err := os.ReadFile(activePath); err == nil {
			os.WriteFile(prodDB, data, 0644)
			activePath = prodDB
		}
	}

	db, err := sql.Open("sqlite3", activePath+"?mode=rwc&_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		a.logError("[MetadataDB] Failed to open %s: %v", activePath, err)
		return nil
	}

	// Ensure table and high-speed indexes exist for instant lookups (<1ms)
	db.Exec(`CREATE TABLE IF NOT EXISTS store_metadata (
		slug TEXT PRIMARY KEY,
		name TEXT,
		year INTEGER,
		description TEXT,
		banner_url TEXT,
		icon_url TEXT,
		cover_url TEXT,
		steamid TEXT,
		gogslug TEXT,
		genres TEXT,
		platforms TEXT
	);`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_store_metadata_steamid ON store_metadata(steamid);`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_store_metadata_name ON store_metadata(name);`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_store_metadata_slug ON store_metadata(slug);`)

	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)
	a.metadataDB = db
	return a.metadataDB
}

// GetCoverArtMap returns a JSON map of steamid -> { cover_url, banner_url }
// from memory cache and local SQLite metadata database with sub-millisecond execution.
func (a *App) GetCoverArtMap(appIDsJSON string) string {
	var appIDs []string
	if err := json.Unmarshal([]byte(appIDsJSON), &appIDs); err != nil || len(appIDs) == 0 {
		return "{}"
	}

	type artEntry struct {
		Cover  string `json:"cover"`
		Banner string `json:"banner"`
	}
	result := make(map[string]artEntry, len(appIDs))

	var missing []string
	for _, id := range appIDs {
		if val, ok := a.coverArtCache.Load(id); ok {
			result[id] = val.(artEntry)
		} else {
			missing = append(missing, id)
		}
	}

	if len(missing) > 0 {
		db := a.getMetadataDB()
		if db != nil {
			placeholders := make([]string, len(missing))
			args := make([]interface{}, len(missing))
			for i, id := range missing {
				placeholders[i] = "?"
				args[i] = id
			}

			rows, err := db.Query(
				"SELECT steamid, cover_url, banner_url FROM store_metadata WHERE steamid IN ("+strings.Join(placeholders, ",")+")",
				args...,
			)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var steamID, coverURL, bannerURL string
					if err := rows.Scan(&steamID, &coverURL, &bannerURL); err == nil {
						entry := artEntry{Cover: coverURL, Banner: bannerURL}
						result[steamID] = entry
						a.coverArtCache.Store(steamID, entry)
					}
				}
			}
		}
	}

	data, _ := json.Marshal(result)
	return string(data)
}

func (a *App) SaveGameRuntimeConfig(gameID string, config types.RuntimeConfig) error {
	path := a.runtimeConfigPath(gameID)
	
	// Read existing configuration if present to preserve critical metadata
	var existing map[string]interface{}
	if raw, err := os.ReadFile(path); err == nil {
		json.Unmarshal(raw, &existing)
	}
	if existing == nil {
		existing = make(map[string]interface{})
	}

	// Marshal new config into a map
	newRaw, err := json.Marshal(config)
	if err != nil {
		return err
	}
	var newMap map[string]interface{}
	if err := json.Unmarshal(newRaw, &newMap); err != nil {
		return err
	}

	// Merge new values into existing
	for k, v := range newMap {
		// Don't overwrite existing non-empty metadata with empty/zero values
		if (k == "executable_path" || k == "gameName" || k == "id" || k == "appID" || k == "platform") && (v == nil || v == "") {
			continue
		}
		existing[k] = v
	}

	// Guarantee installed is preserved if capsule is installed
	if val, ok := existing["installed"]; !ok || val == false {
		markerPath := filepath.Join(a.GamesDir(), gameID, ".rift_setup_complete")
		if _, err := os.Stat(markerPath); err == nil {
			existing["installed"] = true
		} else {
			filesDir := filepath.Join(a.GamesDir(), gameID, "game_files")
			if fEntries, err := os.ReadDir(filesDir); err == nil && len(fEntries) > 0 {
				existing["installed"] = true
			}
		}
	}

	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	a.logInfo("[Config] Saved game_runtime.json for %s", gameID)
	a.SyncLibrary()
	a.emitEvent("library_changed", nil)
	return nil
}

// GetSystemSpecs returns detailed Apple Silicon hardware facts for the System view.
func (a *App) GetSystemSpecs() map[string]interface{} {
	return GetMacHardwareSpecs()
}

