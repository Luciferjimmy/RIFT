package rift

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"rift/internal/engine"
	"rift/internal/parsers"
	"rift/internal/runners"
	"strings"
	"sync"
	"time"
)

var (
	activeDownloadCancels = make(map[string]context.CancelFunc)
	activeDownloadMutex   sync.Mutex
)

// ActiveDownloadRecord represents an ongoing or interrupted download session.
type ActiveDownloadRecord struct {
	GameID    string  `json:"gameId"`
	Platform  string  `json:"platform"`
	AppID     string  `json:"appID"`
	GameName  string  `json:"gameName"`
	Percent   float64 `json:"percent"`
	Speed     float64 `json:"speed"`
	Stage     string  `json:"stage"`
	Status    string  `json:"status"` // "downloading", "download_interrupted", "completed", "failed"
	StartedAt int64   `json:"startedAt"`
}

func (a *App) activeDownloadsFilePath() string {
	return filepath.Join(a.RiftDir(), "active_downloads.json")
}

func (a *App) saveActiveDownloadRecord(rec ActiveDownloadRecord) {
	activeDownloadMutex.Lock()
	defer activeDownloadMutex.Unlock()

	records := make(map[string]ActiveDownloadRecord)
	path := a.activeDownloadsFilePath()
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &records)
	}
	records[rec.GameID] = rec
	if b, err := json.MarshalIndent(records, "", "  "); err == nil {
		_ = os.WriteFile(path, b, 0644)
	}
}

func (a *App) removeActiveDownloadRecord(gameID string) {
	activeDownloadMutex.Lock()
	defer activeDownloadMutex.Unlock()

	records := make(map[string]ActiveDownloadRecord)
	path := a.activeDownloadsFilePath()
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &records)
	}
	delete(records, gameID)
	if b, err := json.MarshalIndent(records, "", "  "); err == nil {
		_ = os.WriteFile(path, b, 0644)
	}
}

// HasActiveDownloads checks if any game download is currently active in memory.
func (a *App) HasActiveDownloads() bool {
	activeDownloadMutex.Lock()
	defer activeDownloadMutex.Unlock()
	return len(activeDownloadCancels) > 0
}

// StopActiveDownloadsOnExit cancels all running download worker goroutines gracefully
// on app close. It leaves active_downloads.json untouched so progress can be resumed later.
func (a *App) StopActiveDownloadsOnExit() {
	activeDownloadMutex.Lock()
	defer activeDownloadMutex.Unlock()
	for gameID, cancel := range activeDownloadCancels {
		if cancel != nil {
			cancel()
		}
		delete(activeDownloadCancels, gameID)
	}
}

// GetActiveDownloads returns a list of all active or interrupted downloads.
func (a *App) GetActiveDownloads() []ActiveDownloadRecord {
	activeDownloadMutex.Lock()
	defer activeDownloadMutex.Unlock()

	records := make(map[string]ActiveDownloadRecord)
	path := a.activeDownloadsFilePath()
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &records)
	}

	var list []ActiveDownloadRecord
	for gameID, rec := range records {
		if _, isActive := activeDownloadCancels[gameID]; !isActive {
			rec.Status = "download_interrupted"
		} else {
			rec.Status = "downloading"
		}
		list = append(list, rec)
	}
	return list
}

// resolveSteamUsername finds the Steam username from the most reliable source available.
// Priority: 1) store.json (RIFT's own auth state), 2) loginusers.vdf (SteamCMD's cached credentials),
// 3) Wine prefix Steam login, 4) empty string (no valid user found).
func (a *App) resolveSteamUsername() string {
	// 1. Try store.json (RIFT's own auth record)
	storeJSON := filepath.Join(a.StoresDir(), "steam", "store.json")
	if data, err := os.ReadFile(storeJSON); err == nil {
		var conf StoreConfig
		if json.Unmarshal(data, &conf) == nil && conf.Username != "" {
			return conf.Username
		}
	}

	// 2. Try SteamCMD's loginusers.vdf (most reliable for cached credential login)
	loginUsersVdf := filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Steam", "config", "loginusers.vdf")
	if data, err := os.ReadFile(loginUsersVdf); err == nil {
		content := string(data)
		// Parse the VDF to find AccountName
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "\"AccountName\"") {
				parts := strings.SplitN(trimmed, "\"", 5)
				if len(parts) >= 4 {
					username := parts[3]
					if username != "" {
						a.logInfo("[Download] Resolved Steam username from loginusers.vdf: %s", username)
						return username
					}
				}
			}
		}
	}

	// 3. Try Wine prefix Steam login parser
	if username, err := parsers.ParseSteamLoginUser(filepath.Join(a.StoresDir(), "steam", "prefix", "drive_c", "Program Files (x86)", "Steam")); err == nil && username != "" {
		a.logInfo("[Download] Resolved Steam username from Wine prefix: %s", username)
		return username
	}

	return ""
}

// AddGameToLibrary adds a Steam game by appID into RIFT's library so the user can download it.
func (a *App) AddGameToLibrary(appID, name string) map[string]interface{} {
	if appID == "" || appID == "0" {
		return map[string]interface{}{"success": false, "error": "Invalid appID"}
	}
	gameID := fmt.Sprintf("steam-%s", appID)
	if name == "" {
		name = fmt.Sprintf("Steam App %s", appID)
	}

	gameCapsuleDir := filepath.Join(a.GamesDir(), gameID)
	os.MkdirAll(gameCapsuleDir, 0755)

	runtimePath := filepath.Join(gameCapsuleDir, "game_runtime.json")
	if _, err := os.Stat(runtimePath); os.IsNotExist(err) {
		conf := map[string]interface{}{
			"id":         gameID,
			"appID":      appID,
			"gameName":   name,
			"platform":   "steam",
			"installed":  false,
			"engine":     "gptk",
			"executable": "",
		}
		b, _ := json.MarshalIndent(conf, "", "  ")
		os.WriteFile(runtimePath, b, 0644)
	}

	a.logInfo("[Library] Added %s (%s) to library", name, appID)
	a.SyncLibrary()
	a.emitEvent("library_changed", nil)

	return map[string]interface{}{"success": true, "gameId": gameID}
}

// GetEpicDownloadPacks returns any selective download / DLC packages available for an Epic title.
func (a *App) GetEpicDownloadPacks(appID string) ([]runners.SelectivePack, error) {
	runner, err := a.getEpicRunner()
	if err != nil {
		return nil, err
	}
	epicAuthDir := filepath.Join(a.UserAuthDir(), "epic")
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	return runner.GetSelectivePacks(ctx, appID, epicAuthDir)
}

// DownloadEpicWithTags initiates an Epic download with specific install tags selected by the user.
func (a *App) DownloadEpicWithTags(gameID, appID string, installTags []string) string {
	if appID == "" {
		appID = strings.TrimPrefix(gameID, "epic-")
	}
	go a.downloadEpicGameWithTagsAsync(gameID, appID, installTags)
	return "started"
}

// AcquireGame is called by the UI when the user clicks DOWNLOAD on a game card.
func (a *App) AcquireGame(gameID, platform, appID string) string {
	if !a.IsLoggedIn() {
		a.logError("[Security] AcquireGame blocked: User is not authenticated")
		a.emitEvent("auth_required", map[string]interface{}{
			"message": "You must be signed in to RIFT to acquire and download games.",
		})
		return "not_authenticated"
	}

	lowerPlatform := strings.ToLower(platform)
	if lowerPlatform == "epic" || strings.HasPrefix(gameID, "epic-") {
		epicAuthPath := filepath.Join(a.UserAuthDir(), "epic", "user.json")
		if _, err := os.Stat(epicAuthPath); err != nil {
			a.logError("[Download] Epic authentication required to download %s", gameID)
			a.emitEvent("show_epic_login", nil)
			return "not_authenticated"
		}

		if appID == "" {
			appID = strings.TrimPrefix(gameID, "epic-")
		}

		// Check for optional selective download packages (e.g. HD Textures, DLCs)
		go func() {
			packs, _ := a.GetEpicDownloadPacks(appID)
			if len(packs) > 0 {
				gameName := queryGameNameFromDB(appID)
				a.logInfo("[Download] Epic game %s has %d optional download packs. Prompting user...", appID, len(packs))
				a.emitEvent("epic_sdl_prompt_required", map[string]interface{}{
					"gameId":   gameID,
					"appID":    appID,
					"gameName": gameName,
					"packs":    packs,
				})
			} else {
				a.downloadEpicGameWithTagsAsync(gameID, appID, nil)
			}
		}()
		return "started"
	}

	if !a.IsStoreReady("steam") {
		a.logError("[Download] Steam authentication required to download %s", gameID)
		a.emitEvent("show_steam_login", nil)
		return "not_authenticated"
	}

	steamcmdDir := a.SteamCMDDir()

	// Automatically redownload SteamCMD if it was cleared/deleted
	if err := engine.EnsureSteamCMD(steamcmdDir); err != nil {
		a.logError("[Download] Failed to setup SteamCMD: %v", err)
		a.emitEvent("download_failed", map[string]interface{}{
			"gameId": gameID,
			"error":  "Failed to initialize Steam downloader.",
		})
		return "error"
	}

	// Resolve the Steam username from store.json, then loginusers.vdf, then Wine prefix
	username := a.resolveSteamUsername()
	if username == "" {
		a.logError("[Download] Could not determine Steam username for download")
		a.emitEvent("show_steam_login", nil)
		return "not_authenticated"
	}

	a.logInfo("[Download] Using Steam account '%s' to download appID %s", username, appID)
	go a.downloadSteamGameAsync(gameID, appID, false, username)
	return "started"
}

func (a *App) downloadEpicGameAsync(gameID, appID string) {
	a.downloadEpicGameWithTagsAsync(gameID, appID, nil)
}

func (a *App) downloadEpicGameWithTagsAsync(gameID, appID string, installTags []string) {
	runner, err := a.getEpicRunner()
	if err != nil {
		a.logError("[Download] Epic runner unavailable: %v", err)
		a.emitEvent("download_failed", map[string]interface{}{
			"gameId":    gameID,
			"error":     "Epic download runner is not available on system.",
			"errorType": "generic",
		})
		return
	}

	gameCapsuleDir := a.ResolveGameDir(gameID)
	installDir := filepath.Join(gameCapsuleDir, "game_files")
	os.MkdirAll(installDir, 0755)
	epicAuthDir := filepath.Join(a.UserAuthDir(), "epic")

	ctx, cancel := context.WithCancel(context.Background())

	activeDownloadMutex.Lock()
	activeDownloadCancels[gameID] = cancel
	activeDownloadMutex.Unlock()

	defer func() {
		activeDownloadMutex.Lock()
		delete(activeDownloadCancels, gameID)
		activeDownloadMutex.Unlock()
	}()

	a.logInfo("[Download] Starting Epic download for %s (AppName: %s, Tags: %v)...", gameID, appID, installTags)
	gameName := queryGameNameFromDB(appID)

	a.saveActiveDownloadRecord(ActiveDownloadRecord{
		GameID:    gameID,
		Platform:  "epic",
		AppID:     appID,
		GameName:  gameName,
		Status:    "downloading",
		StartedAt: time.Now().Unix(),
	})

	a.emitEvent("download_started", map[string]interface{}{
		"gameId":   gameID,
		"appID":    appID,
		"gameName": gameName,
	})
	a.TrackUserEvent("download_started", gameID, appID, "epic", map[string]interface{}{"game_name": gameName})

	maxRetries := 3
	var lastErr error

	progressCallback := func(percent, speed, downloaded, total float64) {
		stageStr := fmt.Sprintf("%.1f / %.1f MB", downloaded, total)
		a.emitEvent("download_progress", map[string]interface{}{
			"gameId":     gameID,
			"percent":    percent,
			"speed":      speed,
			"downloaded": downloaded,
			"total":      total,
			"stage":      stageStr,
		})
		a.saveActiveDownloadRecord(ActiveDownloadRecord{
			GameID:    gameID,
			Platform:  "epic",
			AppID:     appID,
			GameName:  gameName,
			Percent:   percent,
			Speed:     speed,
			Stage:     stageStr,
			Status:    "downloading",
			StartedAt: time.Now().Unix(),
		})
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if ctx.Err() != nil {
			a.emitEvent("download_cancelled", gameID)
			return
		}

		err = runner.DownloadEpicApp(ctx, appID, installDir, epicAuthDir, false, installTags, progressCallback)
		if err == nil {
			// Success!
			a.logInfo("[Download] Epic download completed successfully for %s!", gameID)
			a.TrackUserEvent("download_completed", gameID, appID, "epic", map[string]interface{}{"game_name": gameName})
			a.postDownloadSetup(gameID, appID, "epic", installDir, false)
			return
		}

		lastErr = err
		a.logWarn("[Download] Attempt %d/%d failed for %s: %v", attempt, maxRetries, gameID, err)

		if legErr, ok := err.(*runners.LegendaryError); ok {
			// If cancelled, stop immediately
			if legErr.Type == runners.LegendaryErrCancelled {
				a.emitEvent("download_cancelled", gameID)
				return
			}

			// If auth expired, don't retry, logout and notify user
			if legErr.Type == runners.LegendaryErrAuthExpired {
				a.logError("[Download] Epic session expired! Auto-logging out.")
				a.LogoutEpic()
				a.emitEvent("download_failed", map[string]interface{}{
					"gameId":       gameID,
					"error":        legErr.Message,
					"errorType":    "auth_expired",
					"requireLogin": true,
				})
				return
			}

			// If disk space, don't retry
			if legErr.Type == runners.LegendaryErrDiskSpace {
				a.emitEvent("download_failed", map[string]interface{}{
					"gameId":    gameID,
					"error":     legErr.Message,
					"errorType": "disk_full",
				})
				return
			}

			// If network error and we have retries left, wait and retry
			if legErr.Type == runners.LegendaryErrNetwork && attempt < maxRetries {
				backoff := time.Duration(attempt*2) * time.Second
				a.logInfo("[Download] Network glitch detected. Retrying in %v (attempt %d/%d)...", backoff, attempt+1, maxRetries)
				a.emitEvent("download_progress", map[string]interface{}{
					"gameId":  gameID,
					"percent": 0,
					"speed":   0,
					"stage":   fmt.Sprintf("Connection drop. Retrying in %ds (Attempt %d/%d)...", int(backoff.Seconds()), attempt+1, maxRetries),
				})
				select {
				case <-time.After(backoff):
					continue
				case <-ctx.Done():
					a.emitEvent("download_cancelled", gameID)
					return
				}
			}
		} else {
			// Unclassified error
			if attempt < maxRetries {
				time.Sleep(2 * time.Second)
				continue
			}
		}
	}

	// If all retries exhausted, emit download_failed with classified error
	if legErr, ok := lastErr.(*runners.LegendaryError); ok {
		errorType := "generic"
		if legErr.Type == runners.LegendaryErrNetwork {
			errorType = "network"
		} else if legErr.Type == runners.LegendaryErrDiskSpace {
			errorType = "disk_full"
		}
		a.emitEvent("download_failed", map[string]interface{}{
			"gameId":    gameID,
			"error":     legErr.Message,
			"errorType": errorType,
		})
	} else {
		a.emitEvent("download_failed", map[string]interface{}{
			"gameId":    gameID,
			"error":     lastErr.Error(),
			"errorType": "generic",
		})
	}
}

// DownloadGame re-triggers download for an existing game in library.
func (a *App) DownloadGame(gameID string) string {
	if strings.HasPrefix(gameID, "steam-") {
		appID := strings.TrimPrefix(gameID, "steam-")
		return a.AcquireGame(gameID, "steam", appID)
	}
	appID := strings.TrimPrefix(gameID, "epic-")
	return a.AcquireGame(gameID, "epic", appID)
}

// CancelDownload cancels an active download for a game.
func (a *App) CancelDownload(gameID string) string {
	activeDownloadMutex.Lock()
	cancel, exists := activeDownloadCancels[gameID]
	if exists {
		cancel()
		delete(activeDownloadCancels, gameID)
	}
	activeDownloadMutex.Unlock()
	a.removeActiveDownloadRecord(gameID)

	// Clean up any incomplete capsule directory left by the cancelled download
	gameCapsuleDir := filepath.Join(a.GamesDir(), gameID)
	runtimePath := filepath.Join(gameCapsuleDir, "game_runtime.json")
	if _, statErr := os.Stat(runtimePath); os.IsNotExist(statErr) {
		a.logInfo("[Cleanup] Removing incomplete capsule on download cancel: %s", gameID)
		_ = os.RemoveAll(gameCapsuleDir)
	}

	a.logInfo("[Download] Cancelled download for %s", gameID)
	a.emitEvent("download_cancelled", gameID)
	return "Cancelled"
}

func (a *App) downloadSteamGameAsync(gameID, appID string, isMacNative bool, username string) {
	steamcmdDir := a.SteamCMDDir()
	gameCapsuleDir := filepath.Join(a.GamesDir(), gameID)

	var installDir string
	// Store game files directly in the capsule, outside the Wine prefix.
	// This protects the game files from prefix wipes, significantly shortens paths (avoiding Windows MAX_PATH limits),
	// and standardizes the structure for both Mac Native and Translated titles.
	installDir = filepath.Join(gameCapsuleDir, "game_files")
	os.MkdirAll(installDir, 0755)

	ctx, cancel := context.WithCancel(context.Background())

	activeDownloadMutex.Lock()
	activeDownloadCancels[gameID] = cancel
	activeDownloadMutex.Unlock()

	defer func() {
		activeDownloadMutex.Lock()
		delete(activeDownloadCancels, gameID)
		activeDownloadMutex.Unlock()
	}()

	a.logInfo("[Download] Starting SteamCMD download for %s (appID: %s)...", gameID, appID)
	gameName := queryGameNameFromDB(appID)

	a.saveActiveDownloadRecord(ActiveDownloadRecord{
		GameID:    gameID,
		Platform:  "steam",
		AppID:     appID,
		GameName:  gameName,
		Status:    "downloading",
		StartedAt: time.Now().Unix(),
	})

	a.emitEvent("download_started", map[string]interface{}{
		"gameId":   gameID,
		"appID":    appID,
		"gameName": gameName,
	})
	a.TrackUserEvent("download_started", gameID, appID, "steam", map[string]interface{}{"game_name": gameName})

	// Start tailing progress logs
	progressCh := engine.TailProgress(steamcmdDir, ctx)
	go func() {
		for p := range progressCh {
			stageStr := p.Stage
			if stageStr == "" {
				stageStr = fmt.Sprintf("%.1f / %.1f MB", float64(p.BytesDone)/(1024*1024), float64(p.BytesTotal)/(1024*1024))
			}
			a.emitEvent("download_progress", map[string]interface{}{
				"gameId":     gameID,
				"percent":    p.Percent,
				"speed":      p.SpeedMBps,
				"downloaded": float64(p.BytesDone) / (1024 * 1024),
				"total":      float64(p.BytesTotal) / (1024 * 1024),
				"stage":      stageStr,
			})
			a.saveActiveDownloadRecord(ActiveDownloadRecord{
				GameID:    gameID,
				Platform:  "steam",
				AppID:     appID,
				GameName:  gameName,
				Percent:   p.Percent,
				Speed:     p.SpeedMBps,
				Stage:     stageStr,
				Status:    "downloading",
				StartedAt: time.Now().Unix(),
			})
		}
	}()

	// Force Windows depots unless mac-native
	forceWindows := !isMacNative
	err := engine.DownloadSteamApp(steamcmdDir, installDir, username, appID, forceWindows, ctx)

	if err != nil {
		a.logError("[Download] Download failed for %s: %v", gameID, err)

		// Check if it's a classified SteamCMD error for smart handling
		if steamErr, ok := err.(*engine.SteamCMDError); ok {
			switch steamErr.Type {
			case engine.SteamErrLoginExpired:
				// AUTO-LOGOUT: Force the user to re-login
				a.logError("[Download] Steam session expired! Auto-logging out user.")
				a.LogoutSteam()
				a.emitEvent("download_failed", map[string]interface{}{
					"gameId":       gameID,
					"error":        steamErr.Message,
					"errorType":    "auth_expired",
					"requireLogin": true,
				})
				a.emitEvent("steam_session_expired", map[string]interface{}{
					"message": "Your Steam session has expired. You have been logged out. Please log in again.",
				})
				return

			case engine.SteamErrNoSubscription:
				a.emitEvent("download_failed", map[string]interface{}{
					"gameId":    gameID,
					"error":     steamErr.Message,
					"errorType": "not_owned",
				})
				return

			case engine.SteamErrDiskSpace:
				a.emitEvent("download_failed", map[string]interface{}{
					"gameId":    gameID,
					"error":     steamErr.Message,
					"errorType": "disk_full",
				})
				return

			case engine.SteamErrNetwork:
				a.emitEvent("download_failed", map[string]interface{}{
					"gameId":    gameID,
					"error":     steamErr.Message,
					"errorType": "network",
				})
				return

			case engine.SteamErrRateLimit:
				a.emitEvent("download_failed", map[string]interface{}{
					"gameId":    gameID,
					"error":     steamErr.Message,
					"errorType": "rate_limit",
				})
				return

			case engine.SteamErrCancelled:
				runtimePath := filepath.Join(gameCapsuleDir, "game_runtime.json")
				if _, statErr := os.Stat(runtimePath); os.IsNotExist(statErr) {
					_ = os.RemoveAll(gameCapsuleDir)
				}
				a.emitEvent("download_cancelled", gameID)
				return
			}
		}

		// Generic/unclassified error fallback
		a.emitEvent("download_failed", map[string]interface{}{
			"gameId":    gameID,
			"error":     err.Error(),
			"errorType": "generic",
		})
		return
	}

	a.logInfo("[Download] SteamCMD download completed for %s!", gameID)
	a.TrackUserEvent("download_completed", gameID, appID, "steam", map[string]interface{}{"game_name": gameName})

	// SteamCMD creates steamapps/common/<Game Name> inside the force_install_dir.
	// We need to resolve the actual game directory so it becomes the correct working directory.
	commonDir := filepath.Join(installDir, "steamapps", "common")
	if entries, err := os.ReadDir(commonDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				installDir = filepath.Join(commonDir, e.Name())
				break
			}
		}
	}

	a.postDownloadSetup(gameID, appID, "steam", installDir, isMacNative)
}

func (a *App) postDownloadSetup(gameID, appID, platform, installDir string, isMacNative bool) {
	isMacNative = a.IsGameMacNative(gameID, platform, appID, installDir)

	a.ExecuteSetupRouter(gameID, platform, appID, installDir, isMacNative)

	// Write setup marker so launch doesn't loop
	setupMarker := filepath.Join(a.ResolveGameDir(gameID), ".rift_setup_complete")
	os.WriteFile(setupMarker, []byte("1"), 0644)
	a.removeActiveDownloadRecord(gameID)
	a.emitEvent("download_completed", map[string]interface{}{
		"gameId": gameID,
	})
	a.emitEvent("game_installed", gameID)
	a.SyncLibrary()
}

// SelectGameExecutable allows user to manually pick an executable if auto-detection was ambiguous.
func (a *App) SelectGameExecutable(gameID, relExePath string) string {
	runtimePath := filepath.Join(a.GamesDir(), gameID, "game_runtime.json")
	data, err := os.ReadFile(runtimePath)
	if err != nil {
		return "Game config not found"
	}
	var conf map[string]interface{}
	json.Unmarshal(data, &conf)
	conf["executable"] = relExePath
	conf["installed"] = true

	b, _ := json.MarshalIndent(conf, "", "  ")
	os.WriteFile(runtimePath, b, 0644)

	a.logInfo("[Download] User manually set executable for %s to %s", gameID, relExePath)
	a.emitEvent("game_installed", gameID)
	a.SyncLibrary()
	return "Updated"
}

// RemoveFromLibrary removes an uninstalled game capsule from RIFT's library.
func (a *App) RemoveFromLibrary(gameID string) map[string]interface{} {
	gameDir := a.ResolveGameDir(gameID)
	a.logInfo("[Library] Removing game capsule from library: %s", gameDir)

	// If this is an Epic game, we must properly uninstall it via Legendary first
	// so that Legendary's installed.json doesn't permanently think the game is installed.
	// If we just delete the folder, future downloads will finish instantly with "Download size is 0".
	if !strings.HasPrefix(gameID, "steam-") {
		appID := strings.TrimPrefix(gameID, "epic-")
		if runner, err := a.getEpicRunner(); err == nil {
			epicAuthDir := filepath.Join(a.UserAuthDir(), "epic")
			cmd := exec.Command(runner.BinaryPath, "uninstall", appID, "-y")
			cmd.Env = append(engine.SafeEnviron(), "LEGENDARY_CONFIG_PATH="+epicAuthDir)
			cmd.Run()
		}
	}

	os.RemoveAll(gameDir)
	a.SyncLibrary()
	a.emitEvent("library_changed", nil)
	return map[string]interface{}{"success": true}
}
