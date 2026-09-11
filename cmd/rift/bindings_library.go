package rift

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"rift/internal/adapters"
	"rift/internal/engine"
	"rift/internal/parsers"
	"rift/internal/runners"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// ScanMacForInstalledGames scans the standard /Applications, ~/Applications, and native Steam library folders.
// Any discovered games are automatically updated in RIFT.
func (a *App) ScanMacForInstalledGames() []adapters.RIFTUIProfile {
	a.logInfo("[Library] ===== SCAN MAC FOR INSTALLED GAMES =====")
	var found []adapters.RIFTUIProfile
	home, _ := os.UserHomeDir()

	applicationsDirs := []string{"/Applications", filepath.Join(home, "Applications")}

	// Get Epic game list from runner if logged in
	var epicGames []runners.EpicGameEntry
	if runner, err := a.getEpicRunner(); err == nil {
		epicGames, _ = runner.ListGames()
		a.logInfo("[Library] Found %d Epic game entries for scanning", len(epicGames))
	}

	for _, appDir := range applicationsDirs {
		entries, err := os.ReadDir(appDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() && strings.HasSuffix(entry.Name(), ".app") {
				appNameClean := strings.TrimSuffix(entry.Name(), ".app")

				for _, game := range epicGames {
					// STRICT MATCHING: Only match if the app name exactly matches the Epic AppName or Title.
					// This prevents false positives like 'Arc.app' matching 'Arcade Paradise' or 'Opera.app' matching 'Operation Tango'.
					if strings.EqualFold(game.AppTitle, appNameClean) || strings.EqualFold(game.AppName, appNameClean) {
						appPath := filepath.Join(appDir, entry.Name())
						a.ImportLocalGame(game.AppName, "Epic", game.AppName, appPath)

						dbBackend := queryGameBackendFromDB(game.AppName)
						found = append(found, adapters.AdaptEpicGame(game.AppName, game.AppTitle, true, filepath.Join(a.UserAuthDir(), "epic"), dbBackend))
					}
				}
			}
		}
	}

	// Scan Master Prefix Steam library
	masterSteamapps := filepath.Join(a.StoresDir(), "steam", "prefix", "drive_c", "Program Files (x86)", "Steam", "steamapps")
	if entries, err := os.ReadDir(masterSteamapps); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), "appmanifest_") && strings.HasSuffix(entry.Name(), ".acf") {
				acfPath := filepath.Join(masterSteamapps, entry.Name())
				appID, title, _, isInstalled, err := parsers.ParseAppManifest(acfPath)
				if err == nil && isInstalled && title != "" {
					dbBackend := queryGameBackendFromDB(appID)
					found = append(found, adapters.AdaptSteamGame(appID, title, true, true, dbBackend))
				}
			}
		}
	}

	return found
}

func (a *App) UninstallGame(gameID string) string {
	gameDir := a.ResolveGameDir(gameID)
	a.logInfo("[Uninstall] Removing entire game capsule: %s", gameDir)
	err := os.RemoveAll(gameDir)
	if err != nil {
		a.logError("[Uninstall] Failed to remove game directory: %v", err)
		return fmt.Sprintf("Uninstall failed: %v", err)
	}

	// Also remove from ImportedGames so it doesn't respawn as a ghost
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".rift", "config.json")
	if data, err := os.ReadFile(configPath); err == nil {
		var cfg AppConfig
		if json.Unmarshal(data, &cfg) == nil && cfg.ImportedGames != nil {
			if _, exists := cfg.ImportedGames[gameID]; exists {
				delete(cfg.ImportedGames, gameID)
				if newData, err := json.MarshalIndent(cfg, "", "  "); err == nil {
					os.WriteFile(configPath, newData, 0644)
					a.logInfo("[Uninstall] Successfully removed imported game from config.json")
				}
			}
		}
	}

	return "Game uninstalled successfully."
}

// cleanSharedInstalledEntry removes a game entry from the shared legendary installed.json
// so the per-game .legendary/ dir is the single source of truth.
func (a *App) cleanSharedInstalledEntry(appID string) {
	for _, path := range []string{
		filepath.Join(a.UserAuthDir(), "epic", "installed.json"),
		filepath.Join(os.Getenv("HOME"), ".config", "legendary", "installed.json"),
		filepath.Join(os.Getenv("HOME"), ".config", "nile", "installed.json"),
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var installed map[string]interface{}
		if err := json.Unmarshal(data, &installed); err != nil {
			continue
		}
		if _, exists := installed[appID]; !exists {
			continue
		}
		delete(installed, appID)
		cleaned, _ := json.MarshalIndent(installed, "", "  ")
		os.WriteFile(path, cleaned, 0644)
		a.logInfo("[Clean] Removed %s from shared %s\n", appID, path)
	}
}

// LocateAndImportGame triggers folder picker and configures imported preinstalled game files.
func (a *App) LocateAndImportGame(gameID string, platform string, appID string) string {
	a.logInfo("[Library] ===== LOCATE AND IMPORT: gameID=%s platform=%s appID=%s =====", gameID, platform, appID)
	path, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Game Installation Directory",
	})
	if err != nil || path == "" {
		a.logInfo("[Library] LocateAndImport: cancelled by user")
		return "Cancelled directory selection"
	}
	a.logInfo("[Library] User selected path: %s", path)

	return a.ImportLocalGame(gameID, platform, appID, path)
}

// ImportGameFromFolder is called by the Vacant Exhibit Slot.
// It opens a folder picker and registers any custom native Mac app or Windows game folder.
func (a *App) ImportGameFromFolder() (adapters.RIFTUIProfile, error) {
	a.logInfo("[Library] ===== IMPORT GAME FROM FOLDER =====")
	path, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Game Folder to Import",
	})
	if err != nil || path == "" {
		a.logInfo("[Library] ImportGameFromFolder: cancelled")
		return adapters.RIFTUIProfile{}, fmt.Errorf("cancelled selection")
	}
	a.logInfo("[Library] ImportGameFromFolder: selected path=%s", path)

	title := filepath.Base(path)
	if strings.HasSuffix(strings.ToLower(title), ".app") {
		title = title[:len(title)-4]
	}

	hasher := md5.New()
	hasher.Write([]byte(path))
	gameID := "imported-" + hex.EncodeToString(hasher.Sum(nil))[:8]

	res := a.ImportLocalGame(gameID, "Custom", gameID, path)
	if strings.Contains(res, "Error") {
		return adapters.RIFTUIProfile{}, fmt.Errorf("%s", res)
	}

	isMac := false
	if strings.HasSuffix(path, ".app") || strings.HasSuffix(path, ".app/") {
		isMac = true
	} else {
		filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err == nil && info.IsDir() && strings.HasSuffix(info.Name(), ".app") {
				isMac = true
				return filepath.SkipDir
			}
			return nil
		})
	}

	home, _ := os.UserHomeDir()

	backend := "Awaiting Translation Profile"
	configPath := filepath.Join(home, ".rift", "games", gameID, "rift_config.json")
	if data, err := os.ReadFile(configPath); err == nil {
		var cfg struct {
			Engine string `json:"engine"`
		}
		if json.Unmarshal(data, &cfg) == nil {
			if cfg.Engine == "dxvk" {
				backend = "Wine + DXVK (DX9/10/11)"
			} else {
				backend = "Wine + D3DMetal (DX11/12)"
			}
		}
	}

	if isMac {
		backend = "macOS Native"
	}

	profile := adapters.RIFTUIProfile{
		ID:          gameID,
		Name:        title,
		Cover:       "https://images.unsplash.com/photo-1542751371-adc38448a05e?w=600",
		HeroCover:   "https://images.unsplash.com/photo-1542751371-adc38448a05e?w=1200",
		Platform:    "Custom",
		AppID:       gameID,
		Status:      "ready",
		Playtime:    "—",
		LastPlayed:  "Just imported",
		Backend:     backend,
		Description: "Imported game folder located at: " + path,
		Tags:        []string{"Imported"},
		IsInstalled: true,
		IsMacNative: isMac,
	}

	return profile, nil
}

// ImportLocalGame symlinks game files to save space, and configures the environment correctly.
func (a *App) ImportLocalGame(gameID string, platform string, appID string, localPath string) string {
	if _, err := os.Stat(localPath); err != nil {
		return "Error: selected path does not exist"
	}

	isMacNative := false
	if strings.HasSuffix(localPath, ".app") || strings.HasSuffix(localPath, ".app/") {
		isMacNative = true
	} else {
		filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
			if err == nil && info.IsDir() && strings.HasSuffix(info.Name(), ".app") {
				isMacNative = true
				localPath = path
				return filepath.SkipDir
			}
			return nil
		})
	}

	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".rift", "config.json")
	var cfg AppConfig
	cfg.ImportedGames = make(map[string]string)
	if data, err := os.ReadFile(configPath); err == nil {
		json.Unmarshal(data, &cfg)
	}
	if cfg.ImportedGames == nil {
		cfg.ImportedGames = make(map[string]string)
	}

	if isMacNative {

		cfg.ImportedGames[gameID] = localPath
		newData, _ := json.MarshalIndent(cfg, "", "  ")
		os.WriteFile(configPath, newData, 0644)

		a.logInfo("[Import] Successfully imported native macOS game: %s -> %s\n", gameID, localPath)
		return "Imported native macOS game successfully."
	}

	destDir := filepath.Join(a.ResolveGameDir(gameID), "game_files")

	os.MkdirAll(destDir, 0755)

	entries, err := os.ReadDir(localPath)
	if err != nil {
		return fmt.Sprintf("Error reading source folder: %v", err)
	}
	for _, entry := range entries {
		srcPath := filepath.Join(localPath, entry.Name())
		destPath := filepath.Join(destDir, entry.Name())
		os.Remove(destPath)
		os.Symlink(srcPath, destPath)
	}

	go func() {
		sharedWineDir := filepath.Join(a.EnginesDir(), "wine")
		sharedDXVKDir := filepath.Join(a.EnginesDir(), "dxvk")

		if err := a.EngineManager.EnsureWine(sharedWineDir); err != nil {
			return
		}

		winePrefix := filepath.Join(a.ResolveGameDir(gameID), "prefix_c")

		dbBackend := queryGameBackendFromDB(appID)
		isDX9or10 := strings.Contains(strings.ToLower(dbBackend), "dxvk") ||
			strings.Contains(strings.ToLower(dbBackend), "dx9") ||
			strings.Contains(strings.ToLower(dbBackend), "d9vk") ||
			strings.Contains(strings.ToLower(dbBackend), "opengl")

		if isDX9or10 {
			if err := a.EngineManager.EnsureDXVK(sharedDXVKDir); err == nil {
				a.EngineManager.InstallDXVKToPrefix(winePrefix, sharedDXVKDir)
			}
		}

		if platform == "Steam" {
			// With Master Prefix, importing native mac steam games is not supported or needed,
			// since games are downloaded directly inside the master prefix via the Steam client.
			// However, if we need to manually register a game, we would just create an acf file in the master prefix.
			// But for now, we just skip it to prevent messing up the master prefix.
			a.logInfo("Skipping ImportLocalGame for Steam (Master Prefix handles its own library)")
		}
	}()

	return "Imported Windows game into isolated container successfully."
}

func isAllDigits(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func is32Hex(s string) bool {
	if len(s) != 32 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func detectCapsulePlatformAndAppID(gameID string) (platform string, cleanAppID string) {
	if strings.HasPrefix(gameID, "steam-") {
		return "Steam", strings.TrimPrefix(gameID, "steam-")
	}
	if strings.HasPrefix(gameID, "epic-") {
		return "Epic", strings.TrimPrefix(gameID, "epic-")
	}
	if isAllDigits(gameID) {
		return "Steam", gameID
	}
	if is32Hex(gameID) {
		return "Epic", gameID
	}
	return "Epic", gameID
}

// SyncLibrary fetches the owned games and maps them to the UI Contract.
func (a *App) SyncLibrary() []adapters.RIFTUIProfile {
	a.logInfo("[Library] ===== SYNC LIBRARY START =====")
	if !a.IsLoggedIn() {
		a.logWarn("[Security] SyncLibrary blocked: User is not authenticated")
		return []adapters.RIFTUIProfile{}
	}
	var library []adapters.RIFTUIProfile
	processedGames := make(map[string]bool)

	var cfg AppConfig
	configPath := filepath.Join(a.RiftDir(), "config.json")
	if data, err := os.ReadFile(configPath); err == nil {
		json.Unmarshal(data, &cfg)
	}

	// 1. Scan RIFT Local Game Capsules (~/.rift/games/*) — INSTANT, FAST & RELIABLE
	if entries, err := os.ReadDir(a.GamesDir()); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			gameID := entry.Name()
			gameCapsuleDir := filepath.Join(a.GamesDir(), gameID)

			// Clean up abandoned scratch/empty directories (e.g. aborted downloads)
			filesDir := filepath.Join(gameCapsuleDir, "game_files")
			fEntries, fErr := os.ReadDir(filesDir)
			capsuleEntries, _ := os.ReadDir(gameCapsuleDir)
			if len(capsuleEntries) == 0 || (len(capsuleEntries) == 1 && capsuleEntries[0].Name() == "game_files" && (fErr != nil || len(fEntries) == 0)) {
				a.logInfo("[Cleanup] Pruning abandoned empty capsule directory: %s", gameID)
				_ = os.RemoveAll(gameCapsuleDir)
				continue
			}

			platform, cleanAppID := detectCapsulePlatformAndAppID(gameID)
			cleanName := strings.ReplaceAll(cleanAppID, "-", " ")

			runtimePath := filepath.Join(gameCapsuleDir, "game_runtime.json")
			var rconf map[string]interface{}
			data, err := os.ReadFile(runtimePath)
			needsReconstruction := err != nil || json.Unmarshal(data, &rconf) != nil || rconf == nil

			if !needsReconstruction {
				markerPath := filepath.Join(gameCapsuleDir, ".rift_setup_complete")
				_, setupDone := os.Stat(markerPath)
				isInstalled := setupDone == nil

				if !isInstalled {
					// Game was added to library (e.g. from Store) or download is incomplete
					gameName, _ := rconf["gameName"].(string)
					if gameName == "" {
						gameName = cleanName
					}
					dbBackend := queryGameBackendFromDB(cleanAppID)
					status := "not_downloaded"
					if len(fEntries) > 0 {
						status = "download_interrupted"
					}
					activeDownloadMutex.Lock()
					if _, isDownloading := activeDownloadCancels[gameID]; isDownloading {
						status = "downloading"
					}
					activeDownloadMutex.Unlock()

					if strings.EqualFold(platform, "steam") {
						profile := adapters.AdaptSteamGame(cleanAppID, gameName, false, false, dbBackend)
						profile.Status = status
						library = append(library, profile)
						processedGames[gameID] = true
					} else {
						profile := adapters.AdaptEpicGame(cleanAppID, gameName, false, filepath.Join(a.UserAuthDir(), "epic"), dbBackend)
						profile.Status = status
						library = append(library, profile)
						processedGames[gameID] = true
					}
					continue
				}

				// Verify existing executable is valid and not a utility/extractor binary
				exeVal, _ := rconf["executable"].(string)
				if exeVal == "" {
					exeVal, _ = rconf["executable_path"].(string)
				}
				if exeVal == "" || engine.IsNonGameBinary(exeVal) {
					needsReconstruction = true
				}
			}

			if needsReconstruction {
				// Self-healing: Check if game_files exists inside capsule and has a genuine game binary
				if fErr == nil && len(fEntries) > 0 {
					resolvedExe, _, resErr := engine.ResolveGameExecutable(gameCapsuleDir, cleanAppID, cleanName, a.SteamCMDDir())
					if resErr != nil || resolvedExe == "" || engine.IsNonGameBinary(resolvedExe) {
						a.logWarn("[Self-Healing] Skipping incomplete or utility-only capsule %s: %v", gameID, resErr)
						_ = os.Remove(runtimePath)

						activeDownloadMutex.Lock()
						_, isDownloading := activeDownloadCancels[gameID]
						activeDownloadMutex.Unlock()
						if !isDownloading {
							a.logInfo("[Cleanup] Pruning abandoned incomplete/failed capsule: %s", gameID)
							_ = os.RemoveAll(gameCapsuleDir)
						}
						continue
					}
					a.logWarn("[Self-Healing] game_runtime.json repaired for %s with executable %s", gameID, resolvedExe)
					rconf = map[string]interface{}{
						"gameId":     gameID,
						"appID":      cleanAppID,
						"gameName":   cleanName,
						"platform":   platform,
						"executable": resolvedExe,
						"installed":  true,
						"status":     "ready",
						"engine":     "d3dmetal",
						"backend":    "Wine-Staging 11.10",
					}
					if recData, mErr := json.MarshalIndent(rconf, "", "  "); mErr == nil {
						_ = atomicWriteFile(runtimePath, recData, 0644)
					}
				} else {
					continue
				}
			}

			// Validate that capsule actually has a valid, non-utility executable
			exePath, _ := rconf["executable"].(string)
			if exePath == "" {
				exePath, _ = rconf["executable_path"].(string)
			}
			if exePath == "" || engine.IsNonGameBinary(exePath) {
				a.logWarn("[Library] Skipping capsule %s: executable %q is invalid or a utility binary", gameID, exePath)
				_ = os.Remove(runtimePath)
				continue
			}

			appID, _ := rconf["appID"].(string)
			if appID == "" {
				appID = cleanAppID
			}
			gameName, _ := rconf["gameName"].(string)
			savedPlatform, _ := rconf["platform"].(string)
			if savedPlatform != "" {
				// Fix any capsule where 32-hex ID was mistakenly saved as Steam
				if (savedPlatform == "Steam" || savedPlatform == "steam") && platform == "Epic" {
					savedPlatform = "Epic"
					rconf["platform"] = "Epic"
				}
				platform = savedPlatform
			}
			markerPath := filepath.Join(a.GamesDir(), gameID, ".rift_setup_complete")
			_, setupDone := os.Stat(markerPath)
			isInstalled := setupDone == nil

			if !isInstalled {
				status := "not_downloaded"
				if len(fEntries) > 0 {
					status = "download_interrupted"
				}
				activeDownloadMutex.Lock()
				if _, isDownloading := activeDownloadCancels[gameID]; isDownloading {
					status = "downloading"
				}
				activeDownloadMutex.Unlock()

				dbBackend := queryGameBackendFromDB(appID)
				if platform == "steam" || platform == "Steam" {
					realAppID := appID
					if realAppID == "" {
						realAppID = strings.TrimPrefix(gameID, "steam-")
					}
					profile := adapters.AdaptSteamGame(realAppID, gameName, false, false, dbBackend)
					profile.Status = status
					library = append(library, profile)
					processedGames[gameID] = true
					processedGames["steam-"+realAppID] = true
				} else {
					profile := adapters.AdaptEpicGame(gameID, gameName, false, filepath.Join(a.UserAuthDir(), "epic"), dbBackend)
					profile.Status = status
					library = append(library, profile)
					processedGames[gameID] = true
					processedGames["epic-"+gameID] = true
				}
				continue
			}

			// Name recovery for raw app IDs
			if gameName == "" || gameName == gameID {
				if platform == "steam" || platform == "Steam" {
					realAppID := appID
					if realAppID == "" {
						realAppID = strings.TrimPrefix(gameID, "steam-")
					}
					if dbName := queryGameNameFromDB(realAppID); dbName != "" {
						gameName = dbName
					}
				} else {
					// Check Epic metadata cache
					for _, metaDir := range []string{
						filepath.Join(a.UserAuthDir(), "epic", "metadata"),
						filepath.Join(os.Getenv("HOME"), ".config", "legendary", "metadata"),
					} {
						metaFile := filepath.Join(metaDir, gameID+".json")
						if mData, err := os.ReadFile(metaFile); err == nil {
							var m struct {
								AppTitle string `json:"app_title"`
							}
							if json.Unmarshal(mData, &m) == nil && m.AppTitle != "" {
								gameName = m.AppTitle
								break
							}
						}
					}
				}
			}

			if gameName == "" {
				gameName = gameID
			}

			// Persist repaired installed and gameName back to game_runtime.json
			wasInstalled, _ := rconf["installed"].(bool)
			if isInstalled && (!wasInstalled || rconf["gameName"] != gameName) {
				rconf["installed"] = true
				rconf["gameName"] = gameName
				if b, err := json.MarshalIndent(rconf, "", "  "); err == nil {
					os.WriteFile(runtimePath, b, 0644)
				}
			}

			if platform == "steam" || platform == "Steam" {
				realAppID := appID
				if realAppID == "" {
					realAppID = strings.TrimPrefix(gameID, "steam-")
				}
				dbBackend := queryGameBackendFromDB(realAppID)
				library = append(library, adapters.AdaptSteamGame(realAppID, gameName, isInstalled, false, dbBackend))
				processedGames[gameID] = true
				processedGames["steam-"+realAppID] = true
			} else {
				dbBackend := queryGameBackendFromDB(gameID)
				library = append(library, adapters.AdaptEpicGame(gameID, gameName, isInstalled, filepath.Join(a.UserAuthDir(), "epic"), dbBackend))
				processedGames[gameID] = true
				processedGames["epic-"+gameID] = true
			}
		}
	}
	a.logInfo("[Library] Synced %d games from local RIFT capsules", len(library))

	// 2. Scan Epic Games ONLY IF Epic user is logged in (avoid running CLI binary checks needlessly)
	epicAuthFile := filepath.Join(a.UserAuthDir(), "epic", "user.json")
	if _, err := os.Stat(epicAuthFile); err == nil {
		a.logInfo("[Library] Epic login detected, checking Epic games...")
		if runner, err := a.getEpicRunner(); err == nil {
			if _, errUsername := runner.GetUsernameFast(); errUsername == nil {
				if games, errList := runner.ListGames(); errList == nil {
					home, _ := os.UserHomeDir()
					epicPaths := []string{
						filepath.Join(a.UserAuthDir(), "epic", "installed.json"),
						filepath.Join(home, ".config", "legendary", "installed.json"),
						filepath.Join(home, ".config", "nile", "installed.json"),
					}
					epicInstalledMap := parsers.ParseLegendaryInstalled(epicPaths)

					for _, game := range games {
						if processedGames[game.AppName] {
							// Update name if it was loaded from local capsules without a proper name
							for i := range library {
								if strings.EqualFold(library[i].ID, game.AppName) && (library[i].Name == "" || strings.EqualFold(library[i].Name, game.AppName)) {
									library[i].Name = game.AppTitle
									// Persist the discovered name back to game_runtime.json
									runtimePath := filepath.Join(a.GamesDir(), library[i].ID, "game_runtime.json")
									if data, err := os.ReadFile(runtimePath); err == nil {
										var rconf map[string]interface{}
										if json.Unmarshal(data, &rconf) == nil {
											rconf["gameName"] = game.AppTitle
											if b, err := json.MarshalIndent(rconf, "", "  "); err == nil {
												os.WriteFile(runtimePath, b, 0644)
											}
										}
									}
								}
							}
							continue
						}
						isInstalled := false
						if _, exists := epicInstalledMap[game.AppName]; exists {
							isInstalled = true
						}
						dbBackend := queryGameBackendFromDB(game.AppName)
						library = append(library, adapters.AdaptEpicGame(game.AppName, game.AppTitle, isInstalled, filepath.Join(a.UserAuthDir(), "epic"), dbBackend))
						processedGames[game.AppName] = true
					}
				} else {
					a.logWarn("[Library] Failed to list Epic games: %v", errList)
				}
			}
		}
	}

	// 3. Scan Steam localconfig.vdf ONLY IF Steam user is logged in
	steamAuthDir := a.steamGlobalAuthDir()
	home, _ := os.UserHomeDir()
	steamPrefixDir := filepath.Join(a.UserAuthDir(), "steam_prefix", "drive_c", "Program Files (x86)", "Steam")
	macNativeSteamDir := filepath.Join(home, "Library", "Application Support", "Steam")
	ownedSteamGames := make(map[string]bool)

	if steam32ID := a.getSteam32ID(); steam32ID != "" {
		if ownedAppIDs, err := parsers.ParseOwnedSteamGames(steamAuthDir, steam32ID); err == nil {
			for _, appID := range ownedAppIDs {
				ownedSteamGames[appID] = true
			}
			a.logInfo("[Library] Synced %d owned Steam games from localconfig.vdf", len(ownedAppIDs))
		}
	}

	vdfPath1 := filepath.Join(steamPrefixDir, "steamapps", "libraryfolders.vdf")
	for _, appID := range parsers.ParseInstalledAppIDs(vdfPath1) {
		ownedSteamGames[appID] = true
	}
	vdfPath2 := filepath.Join(macNativeSteamDir, "steamapps", "libraryfolders.vdf")
	for _, appID := range parsers.ParseInstalledAppIDs(vdfPath2) {
		ownedSteamGames[appID] = true
	}

	for appID := range ownedSteamGames {
		// Skip if already processed from local capsules
		if processedGames["steam-"+appID] {
			continue
		}
		// Skip non-game apps (system tools, SDKs, configs, etc.)
		if !isSteamGame(appID) {
			continue
		}

		isMacNative := a.isSteamGameMacNative(appID)
		isInstalled := false

		if isMacNative {
			nativeSteamapps := filepath.Join(macNativeSteamDir, "steamapps")
			if parsers.CheckIfGameInstalled(nativeSteamapps, appID) {
				isInstalled = true
			}
		}

		title := queryGameNameFromDB(appID)
		if title == "" {
			// No local name available — skip this game for now,
			// it will appear once its name is cached
			continue
		}

		dbBackend := queryGameBackendFromDB(appID)
		library = append(library, adapters.AdaptSteamGame(appID, title, isInstalled, isMacNative, dbBackend))
		processedGames["steam-"+appID] = true
	}

	if cfg.ImportedGames != nil {
		for gameID, path := range cfg.ImportedGames {

			if strings.HasPrefix(gameID, "steam-") || len(gameID) == 32 {
				continue
			}

			title := filepath.Base(path)
			if strings.HasSuffix(strings.ToLower(title), ".app") {
				title = title[:len(title)-4]
			}

			isMac := strings.HasSuffix(path, ".app") || strings.HasSuffix(path, ".app/")
			backend := "Awaiting Translation Profile"
			configPath := filepath.Join(a.RiftDir(), "games", gameID, "rift_config.json")
			if data, err := os.ReadFile(configPath); err == nil {
				var cfg struct {
					Engine string `json:"engine"`
				}
				if json.Unmarshal(data, &cfg) == nil {
					if cfg.Engine == "dxvk" {
						backend = "Wine + DXVK (DX9/10/11)"
					} else {
						backend = "Wine + D3DMetal (DX11/12)"
					}
				}
			}

			if isMac {
				backend = "macOS Native"
			}

			// Read last_played.txt if it exists
			var lastPlayedTs int64
			if lpBytes, err := os.ReadFile(filepath.Join(a.ResolveGameDir(gameID), "last_played.txt")); err == nil {
				fmt.Sscanf(string(lpBytes), "%d", &lastPlayedTs)
			}

			library = append(library, adapters.RIFTUIProfile{
				ID:          gameID,
				Name:        title,
				Cover:       "https://images.unsplash.com/photo-1542751371-adc38448a05e?w=600",
				HeroCover:   "https://images.unsplash.com/photo-1542751371-adc38448a05e?w=1200",
				Platform:    "Custom",
				AppID:       gameID,
				Status:      "ready",
				Playtime:    "—",
				LastPlayed:  "Imported",
				Backend:     backend,
				Description: "Imported game folder located at: " + path,
				Tags:        []string{"Imported"},
				IsInstalled: true,
				IsMacNative: isMac,
			})
		}
	}

	a.logInfo("[Library] ===== SYNC LIBRARY END: %d total games =====", len(library))
	return library
}

// getSteam32ID scans the userdata directory for a numeric Steam account folder.
// Typically the path is: <globalAuthDir>/userdata/<steam32ID>/
// Returns the first numeric directory name found, or empty string if none.
func (a *App) getSteam32ID() string {
	steamAuthDir := a.steamGlobalAuthDir()
	userdataDir := filepath.Join(steamAuthDir, "userdata")
	entries, err := os.ReadDir(userdataDir)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			isNumeric := true
			for _, c := range name {
				if c < '0' || c > '9' {
					isNumeric = false
					break
				}
			}
			if isNumeric && len(name) > 0 {
				return name
			}
		}
	}
	return ""
}
