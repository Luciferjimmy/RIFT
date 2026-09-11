package rift

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"rift/internal/engine"
	"rift/internal/optimizer"
	"rift/internal/parsers"
	"rift/internal/sysmon"
	"rift/internal/types"
	"runtime"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// launchMutex prevents multiple rapid Play clicks from launching the same game twice.
var launchMutex sync.Mutex

// handleNativeAcquire installs the native macOS version of the game.
func (a *App) handleNativeAcquire(gameID, platform, appID string) {
	if platform == "Epic" {
		runner, err := a.getEpicRunner()
		if err != nil {
			return
		}
		destDir := filepath.Join(a.ResolveGameDir(gameID), "game_files")
		runner.InstallMac(appID, destDir)

		foundApp := false
		filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && info.IsDir() && strings.HasSuffix(info.Name(), ".app") {
				foundApp = true
				return filepath.SkipDir
			}
			return nil
		})

		if foundApp {
			a.emitEvent("game_installed", gameID)
		} else {
			a.logInfo("[AcquireGame] Epic marked Mac native but provided Windows files. Falling back to Wine.\n")
			go a.acquirePipeline(gameID, platform, appID, true)
		}
	} else if platform == "Steam" {
		// Sanitize appID
		for _, c := range appID {
			if c < '0' || c > '9' {
				a.logError("[Launch] Invalid Steam appID: %s", appID)
				return
			}
		}
		if err := exec.Command("open", "steam://install/"+appID).Run(); err != nil {
			a.logError("[Launch] Failed to trigger steam://install: %v", err)
		}
		return
	}
}

// PlayGame launches a game if ready, or starts download if not yet downloaded.
func (a *App) PlayGame(gameID string, platform string, appID string) string {
	a.logInfo("[Launch] ===== PLAY GAME: gameID=%s platform=%s appID=%s =====", gameID, platform, appID)

	if !a.IsLoggedIn() {
		a.logError("[Security] PlayGame blocked: User is not authenticated")
		a.emitEvent("auth_required", map[string]interface{}{
			"message": "You must be signed in to RIFT to launch games.",
		})
		return "AUTHENTICATION_REQUIRED"
	}

	// Spam protection: only one launch at a time
	if !launchMutex.TryLock() {
		a.logWarn("[Launch] Launch already in progress, ignoring duplicate click for %s", gameID)
		return "Launch already in progress."
	}
	defer launchMutex.Unlock()

	// Check if post-download setup completed (writes this marker for ALL platforms)
	gameDir := a.ResolveGameDir(gameID)
	setupMarker := filepath.Join(gameDir, ".rift_setup_complete")
	prefixDir := filepath.Join(gameDir, "prefix")
	driveC := filepath.Join(prefixDir, "drive_c")

	if _, err := os.Stat(setupMarker); err != nil {
		a.logInfo("[Launch] No setup marker found, triggering download for %s", gameID)
		return a.DownloadGame(gameID)
	}

	// Self-healing: if user deleted prefix/drive_c in Finder, auto-rebuild Wine bottle
	if _, err := os.Stat(driveC); os.IsNotExist(err) {
		a.logWarn("[Self-Healing] Game prefix/drive_c missing for %s. Auto-rebuilding Wine capsule...", gameID)
		a.emitEvent("launch_status", map[string]interface{}{
			"gameId": gameID,
			"status": "Rebuilding Wine Capsule...",
		})
		installDir := filepath.Join(gameDir, "game_files")
		if strings.HasPrefix(gameID, "epic-") {
			_ = a.SetupEpicTranslated(gameID, appID, installDir)
		} else {
			_ = a.SetupSteamTranslated(gameID, appID, installDir)
		}
	}

	return a.ExecuteTranslation(gameID, platform, appID)
}

// ExecuteTranslation is the main router that decides whether to launch natively or through translation.
func (a *App) ExecuteTranslation(gameID string, platform string, appID string) string {
	a.logInfo("[Launch] ===== EXECUTE ROUTER: gameID=%s platform=%s appID=%s =====", gameID, platform, appID)
	
	// Emit event to trigger the frontend Loading Overlay
	a.emitEvent("launch_started", map[string]interface{}{
		"gameId": gameID,
		"status": "Initializing Environment...",
	})

	// Trigger 'first_blood' achievement (First launch) silently in background
	go a.UnlockAchievement("first_blood")
	a.TrackUserEvent("game_launched", gameID, appID, platform, nil)

	isolatedCapsule := a.ResolveGameDir(gameID)
	setupMarker := filepath.Join(isolatedCapsule, ".rift_setup_complete")
	if _, err := os.Stat(setupMarker); os.IsNotExist(err) {
		a.logError("[Launch] Launch blocked for %s: .rift_setup_complete missing. Game download is incomplete.", gameID)
		a.emitEvent("launch_failed", map[string]interface{}{
			"gameId": gameID,
			"error":  "Game download is incomplete. Please finish downloading before launching.",
		})
		return "Error: Game download is incomplete. Please resume or finish downloading first."
	}
	gameDir := filepath.Join(isolatedCapsule, "game_files")
	runtimePath := filepath.Join(isolatedCapsule, "game_runtime.json")
	if data, err := os.ReadFile(runtimePath); err == nil {
		var rconf map[string]interface{}
		if err := json.Unmarshal(data, &rconf); err == nil {
			if gdir, ok := rconf["gameDir"].(string); ok && gdir != "" {
				gameDir = gdir
			}
		}
	}

	isMacNative := a.IsGameMacNative(gameID, platform, appID, gameDir)
	if isMacNative {
		if strings.EqualFold(platform, "epic") {
			return a.LaunchEpicNative(gameID, appID, gameDir)
		}
		if strings.EqualFold(platform, "steam") {
			return a.LaunchSteamNative(gameID, appID, gameDir)
		}
	}

	// Translation Route
	rt := &types.RuntimeConfig{EnvVars: make(map[string]string), DLLOverrides: make(map[string]string)}
	hasCuratedConfig := false
	if data, err := os.ReadFile(a.runtimeConfigPath(gameID)); err == nil {
		if err := json.Unmarshal(data, rt); err == nil {
			var rmap map[string]interface{}
			_ = json.Unmarshal(data, &rmap)
			if curated, ok := rmap["ai_curated"].(bool); ok && curated {
				hasCuratedConfig = true
			} else if rt.ManualOverride || rt.Engine != "" {
				hasCuratedConfig = true
			} else if len(rt.Winetricks) > 0 || len(rt.LaunchArgs) > 0 || len(rt.DLLOverrides) > 0 {
				hasCuratedConfig = true
			}
		}
	}

	if !hasCuratedConfig {
		a.logInfo("[Launch] Capsule for %s is not AI-curated yet. Triggering Supabase AI Curation...", gameID)
		a.runAICuration(gameID, appID, filepath.Join(isolatedCapsule, "prefix"))
		if data, err := os.ReadFile(a.runtimeConfigPath(gameID)); err == nil {
			json.Unmarshal(data, rt)
		}
		if rt.Engine == "" {
			rt.Engine = "d3dmetal"
		}
	}

	// Trigger translation-specific achievements
	if strings.EqualFold(rt.Engine, "d3dmetal") || strings.EqualFold(rt.Engine, "gptk") {
		go a.UnlockAchievement("dx12_vanguard")
	}

	configPath := filepath.Join(a.RiftDir(), "config.json")
	var cfg AppConfig
	if data, err := os.ReadFile(configPath); err == nil {
		json.Unmarshal(data, &cfg)
	}
	override, hasOverride := cfg.GameOverrides[gameID]

	if rt.EnvVars == nil {
		rt.EnvVars = make(map[string]string)
	}

	if hasOverride && override.Engine != "" && override.Engine != "auto" {
		rt.Engine = override.Engine
	}
	hudSetting := "auto"
	if hasOverride && override.HUD != "" && override.HUD != "auto" {
		hudSetting = override.HUD
	}
	if hudSetting == "force-on" || (hudSetting == "auto" && cfg.GlobalHUD) {
		rt.EnvVars["MTL_HUD_ENABLED"] = "1"
	} else if hudSetting == "force-off" || (hudSetting == "auto" && !cfg.GlobalHUD) {
		rt.EnvVars["MTL_HUD_ENABLED"] = "0"
	}

	retinaSetting := "auto"
	if hasOverride && override.Retina != "" && override.Retina != "auto" {
		retinaSetting = override.Retina
	}
	if retinaSetting == "force-on" || (retinaSetting == "auto" && cfg.GlobalRetina) {
		rt.EnvVars["WINE_RETINA_MODE"] = "Y"
		rt.EnvVars["WINEMAC_RETINA_MODE"] = "Y"
		rt.EnvVars["WINE_HIGH_DPI"] = "1"
	} else if retinaSetting == "force-off" || (retinaSetting == "auto" && !cfg.GlobalRetina) {
		rt.EnvVars["WINE_RETINA_MODE"] = "N"
		rt.EnvVars["WINEMAC_RETINA_MODE"] = "N"
		rt.EnvVars["WINE_HIGH_DPI"] = "0"
	}

	if cfg.GlobalEsync {
		rt.EnvVars["WINEESYNC"] = "1"
	}
	if hasOverride && override.Args != "" {
		customArgs := strings.Fields(override.Args)
		rt.LaunchArgs = append(rt.LaunchArgs, customArgs...)
	}

	if strings.EqualFold(platform, "epic") {
		return a.LaunchEpicTranslated(gameID, appID, gameDir, rt)
	}
	return a.LaunchSteamTranslated(gameID, appID, gameDir, rt)
}

// hasExeFiles checks if a directory contains any non-trivial .exe files
// (used to detect legacy Steam root installs).
func hasExeFiles(dir string) bool {
	found := false
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || found {
			return filepath.SkipAll
		}
		if !info.IsDir() && strings.EqualFold(filepath.Ext(info.Name()), ".exe") && info.Size() > 100*1024 {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

// pickGameExecutable finds the best game exe in a directory.
// Skips known installers/helpers and prefers the largest exe (usually the game).
func pickGameExecutable(gameDir string) string {
	skipExes := map[string]bool{
		"eosbootstrapper.exe":      true,
		"eosoverlayrenderer32.exe": true,
		"eosoverlayrenderer64.exe": true,
		"unins000.exe":             true,
		"uninstall.exe":            true,
		"dxsetup.exe":              true,
		"vc_redist.exe":            true,
		"vcredist.exe":             true,
	}
	var best string
	var bestSize int64
	filepath.Walk(gameDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable paths
		}
		if info.IsDir() || !strings.EqualFold(filepath.Ext(info.Name()), ".exe") {
			return nil
		}
		name := strings.ToLower(info.Name())
		if skipExes[name] || strings.HasPrefix(name, "dotnet") || strings.HasPrefix(name, "vc_redist") {
			return nil
		}
		// Prefer non-helper executables > 100KB
		if info.Size() > 100*1024 && info.Size() > bestSize {
			best = path
			bestSize = info.Size()
		}
		return nil
	})
	return best
}

func (a *App) ConfigureGamePrefix(gameID string) string {
	isolatedCapsule := a.ResolveGameDir(gameID)
	winePrefix := filepath.Join(isolatedCapsule, "prefix")
	runtimePath := filepath.Join(isolatedCapsule, "game_runtime.json")
	if data, err := os.ReadFile(runtimePath); err == nil {
		var rconf map[string]interface{}
		if err := json.Unmarshal(data, &rconf); err == nil {
			if pdir, ok := rconf["prefixDir"].(string); ok && pdir != "" {
				winePrefix = pdir
			}
		}
	}
	return winePrefix
}

func (a *App) EnsureEssentialRuntimes(wineBinary, winePrefix string) {
	marker := filepath.Join(winePrefix, ".master_redist_seeded")
	if _, err := os.Stat(marker); err == nil {
		return // Already completely pre-seeded
	}

	system32 := filepath.Join(winePrefix, "drive_c", "windows", "system32")
	syswow64 := filepath.Join(winePrefix, "drive_c", "windows", "syswow64")

	os.MkdirAll(system32, 0755)
	os.MkdirAll(syswow64, 0755)

	a.logInfo("[Setup] Pre-seeding master runtime redists (DirectX, VC++ 2005-2022, XInput, XAudio2, Media Foundation) into prefix %s...", winePrefix)

	gptkX64 := filepath.Join(a.EnginesDir(), "gptk", "lib", "wine", "x86_64-windows")
	gptkX32 := filepath.Join(a.EnginesDir(), "gptk", "lib", "wine", "i386-windows")
	wineX64 := filepath.Join(a.EnginesDir(), "wine", "lib", "wine", "x86_64-windows")
	wineX32 := filepath.Join(a.EnginesDir(), "wine", "lib", "wine", "i386-windows")

	copyFile := func(src, dst string) error {
		source, err := os.Open(src)
		if err != nil {
			return err
		}
		defer source.Close()
		destination, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer destination.Close()
		_, err = io.Copy(destination, source)
		return err
	}

	patterns := []string{
		"vcruntime*.dll",
		"msvcp*.dll",
		"msvcr*.dll",
		"vcomp*.dll",
		"ucrtbase.dll",
		"d3dcompiler_*.dll",
		"d3dx9_*.dll",
		"d3dx10_*.dll",
		"d3dx11_*.dll",
		"d3dxof.dll",
		"xinput*.dll",
		"xaudio2_*.dll",
		"x3daudio1_*.dll",
		"xapofx1_*.dll",
		"wmvcore.dll",
		"quartz.dll",
		"wintypes.dll",
	}

	copyPatterns := func(srcDir, dstDir string) {
		for _, pattern := range patterns {
			matches, _ := filepath.Glob(filepath.Join(srcDir, pattern))
			for _, match := range matches {
				target := filepath.Join(dstDir, filepath.Base(match))
				if _, err := os.Stat(target); os.IsNotExist(err) {
					copyFile(match, target)
				}
			}
		}
	}

	// Copy 64-bit DLLs into system32
	copyPatterns(gptkX64, system32)
	copyPatterns(wineX64, system32)

	// Copy 32-bit DLLs into syswow64
	copyPatterns(gptkX32, syswow64)
	copyPatterns(wineX32, syswow64)

	os.WriteFile(marker, []byte("ok"), 0644)
	a.logInfo("[Setup] Master redist pre-seeding complete.")
}

func (a *App) runAICuration(gameID, appID, winePrefix string) {
	a.logInfo("[Supabase AI] Fetching curated configuration for %s...", gameID)

	// 1. Build SystemProfile using optimizer package
	profiler := optimizer.NewSystemProfiler()
	optProfile := profiler.Profile()
	
	// Map to types.SystemProfile
	sysProfile := types.SystemProfile{
		Chip:           "Apple Silicon", // Reasonable fallback
		RAMGB:          optProfile.RAMGB,
		MacOSVersion:   optProfile.MacOSVersion,
		GPUMetalFamily: "Apple9", // Assuming an M-series chip
		HasGPTK:        optProfile.Engines.GPTK,
		HasDXVK:        optProfile.Engines.DXVK,
		HasMoltenVK:    true, // Assume true for modern setups
		HardwareUUID:   "",
		AppleSilicon:   optProfile.AppleSilicon,
		MetalGPUFamily: optProfile.MetalGPUFamily,
		Rosetta2:       optProfile.Rosetta2,
		HasWine:        optProfile.Engines.Wine,
	}

	runtimePath := a.runtimeConfigPath(gameID)
	var existingConfig types.RuntimeConfig
	var rawConfig map[string]interface{}

	if data, err := os.ReadFile(runtimePath); err == nil {
		json.Unmarshal(data, &existingConfig)
		json.Unmarshal(data, &rawConfig)
	}

	if existingConfig.ManualOverride || (rawConfig != nil && rawConfig["ai_curated"] == true) {
		a.logInfo("[Supabase AI] Manual Override or existing AI curation is enabled for %s. Skipping AI curation.", gameID)
		return
	}

	rt := types.RuntimeConfig{
		Engine:         existingConfig.Engine,
		ExecutablePath: existingConfig.ExecutablePath,
		EnvVars:        make(map[string]string),
		DLLOverrides:   make(map[string]string),
		Winetricks:     []string{},
		LaunchArgs:     []string{},
	}

	if existingConfig.EnvVars != nil {
		for k, v := range existingConfig.EnvVars {
			rt.EnvVars[k] = v
		}
	}
	if existingConfig.DLLOverrides != nil {
		for k, v := range existingConfig.DLLOverrides {
			rt.DLLOverrides[k] = v
		}
	}
	if len(existingConfig.Winetricks) > 0 {
		rt.Winetricks = append(rt.Winetricks, existingConfig.Winetricks...)
	}
	if len(existingConfig.LaunchArgs) > 0 {
		rt.LaunchArgs = append(rt.LaunchArgs, existingConfig.LaunchArgs...)
	}

	// Try to deduce platform: if appID is all digits, steam, otherwise epic
	platform := "epic"
	isSteam := true
	for _, c := range appID {
		if c < '0' || c > '9' {
			isSteam = false
			break
		}
	}
	if isSteam && appID != "" {
		platform = "steam"
	}

	if a.Supabase != nil && a.Supabase.IsLoggedIn() {
		realGameName := queryGameNameFromDB(appID)
		if realGameName == "" {
			realGameName = queryGameNameFromDB(gameID)
		}
		if realGameName == "" {
			realGameName = gameID
			if strings.EqualFold(platform, "epic") {
				gameFilesPath := filepath.Join(a.ResolveGameDir(gameID), "game_files")
				if entries, err := os.ReadDir(gameFilesPath); err == nil {
					for _, entry := range entries {
						if entry.IsDir() && entry.Name() != ".egstore" && entry.Name() != "Engine" && !strings.HasPrefix(entry.Name(), ".") {
							realGameName = entry.Name()
							break
						}
					}
				}
			}
		}

		cloudConf, err := a.Supabase.GetGameConfig(appID, realGameName, platform, "", sysProfile)
		if err != nil {
			a.logError("[Supabase AI] Failed to fetch game config: %v", err)
		} else if cloudConf != nil {
			a.logInfo("[Supabase AI] Successfully fetched curated config from AI provider: %s (%s)", cloudConf.AIProvider, cloudConf.AIModel)
			
			if cloudConf.Engine != "" {
				rt.Engine = cloudConf.Engine
			}
			if cloudConf.ExecutablePath != "" {
				rt.ExecutablePath = cloudConf.ExecutablePath
			}
			for k, v := range cloudConf.EnvVars {
				rt.EnvVars[k] = v
			}
			for k, v := range cloudConf.DLLOverrides {
				rt.DLLOverrides[k] = v
			}
			if len(cloudConf.Winetricks) > 0 {
				rt.Winetricks = append(rt.Winetricks, cloudConf.Winetricks...)
			}
			if len(cloudConf.LaunchArgs) > 0 {
				rt.LaunchArgs = append(rt.LaunchArgs, cloudConf.LaunchArgs...)
			}
		}
	} else {
		a.logInfo("[Supabase AI] Supabase client unavailable or not logged in. Falling back to local defaults.")
	}

	// Deduplicate Winetricks and LaunchArgs
	dedupSlice := func(items []string) []string {
		seen := make(map[string]bool)
		var result []string
		for _, item := range items {
			trimmed := strings.TrimSpace(item)
			if trimmed != "" && !seen[trimmed] {
				seen[trimmed] = true
				result = append(result, trimmed)
			}
		}
		return result
	}
	rt.Winetricks = dedupSlice(rt.Winetricks)
	rt.LaunchArgs = dedupSlice(rt.LaunchArgs)

	// Default fallback engine if still empty
	if rt.Engine == "" {
		rt.Engine = "d3dmetal" // best for DX11/12 on macOS
	}
	
	// Apply hardware hints
	engine.ApplyHardwareHints(&rt, sysProfile)

	if rawConfig == nil {
		rawConfig = make(map[string]interface{})
	}
	if appID == "11020" || strings.Contains(strings.ToLower(rt.ExecutablePath), "tmforever") {
		if strings.Contains(strings.ToLower(rt.ExecutablePath), "launcher") || rt.ExecutablePath == "" {
			rt.ExecutablePath = "TmForever.exe"
		}
	}
	rawConfig["engine"] = rt.Engine
	rawConfig["executable_path"] = rt.ExecutablePath
	rawConfig["env_vars"] = rt.EnvVars
	rawConfig["dll_overrides"] = rt.DLLOverrides
	rawConfig["winetricks"] = rt.Winetricks
	rawConfig["launch_args"] = rt.LaunchArgs
	rawConfig["ai_curated"] = true

	data, marshalErr := json.MarshalIndent(rawConfig, "", "  ")
	if marshalErr == nil {
		os.MkdirAll(filepath.Dir(runtimePath), 0755)
		if writeErr := os.WriteFile(runtimePath, data, 0644); writeErr == nil {
			a.logInfo("[Supabase AI] Successfully wrote game_runtime.json for %s", gameID)
		} else {
			a.logError("[Supabase AI] Failed to write game_runtime.json: %v", writeErr)
		}
	} else {
		a.logError("[Supabase AI] Failed to serialize game_runtime.json: %v", marshalErr)
	}
}

func (a *App) applyWinetricksDependencies(wineBinary, winePrefix string, deps []string) error {
	var safeDeps []string
	for _, dep := range deps {
		lower := strings.ToLower(dep)
		// Only block 'directx9' — the full DirectX 9 runtime installer that overwrites d3d9.dll.
		// d3dx9 (shader/math helper library) is SAFE — it installs d3dx9_24..d3dx9_43 helper DLLs,
		// NOT the core d3d9.dll runtime.
		if lower == "directx9" {
			a.logInfo("[Winetricks] Filtering out destructive dependency: %s", dep)
			continue
		}
		safeDeps = append(safeDeps, dep)
	}

	if len(safeDeps) == 0 {
		return nil
	}
	a.logInfo("[Winetricks] Applying dependencies: %v", safeDeps)

	// Execute winetricks with Wine prefix — prefer bundled winetricks if present
	winetricksPath := filepath.Join(a.EnginesDir(), "wine", "bin", "winetricks")
	if _, err := os.Stat(winetricksPath); os.IsNotExist(err) {
		for _, p := range []string{"/opt/homebrew/bin/winetricks", "/usr/local/bin/winetricks", "/usr/bin/winetricks"} {
			if _, err := os.Stat(p); err == nil {
				winetricksPath = p
				break
			}
		}
	}

	if _, err := os.Stat(winetricksPath); os.IsNotExist(err) {
		a.logWarn("[Winetricks] winetricks not found at any standard path; skipping dependency installation")
		return nil
	}

	args := append([]string{"-q"}, safeDeps...)
	cmd := exec.Command(winetricksPath, args...)
	cmd.Env = append(engine.SafeEnviron(),
		"WINEPREFIX="+winePrefix,
		"WINE="+wineBinary,
		"WINEDEBUG=-all",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		a.logWarn("[Winetricks] Failed to apply dependencies %v: %v", safeDeps, err)
		return err
	}

	a.logInfo("[Winetricks] Successfully applied dependencies: %v", safeDeps)
	return nil
}

func (a *App) InjectEngineDLLs(prefixPath string, engineType string) {
	a.logInfo("[Launch] Injecting %s DLLs into prefix %s", engineType, prefixPath)
	
	system32 := filepath.Join(prefixPath, "drive_c", "windows", "system32")
	syswow64 := filepath.Join(prefixPath, "drive_c", "windows", "syswow64")

	os.MkdirAll(system32, 0755)
	os.MkdirAll(syswow64, 0755)

	copyFile := func(srcFile, destFile string) {
		source, err := os.Open(srcFile)
		if err != nil {
			return
		}
		defer source.Close()
		
		destination, err := os.Create(destFile)
		if err != nil {
			return
		}
		defer destination.Close()
		
		io.Copy(destination, source)
	}

	engineType = strings.ToLower(engineType)
	
	// Inject DXVK DLLs
	if engineType == "dxvk" {
		dxvkX64 := filepath.Join(a.EnginesDir(), "dxvk", "x64")
		dxvkX32 := filepath.Join(a.EnginesDir(), "dxvk", "x32")
		
		copyFile(filepath.Join(dxvkX64, "dxgi.dll"), filepath.Join(system32, "dxgi.dll"))
		copyFile(filepath.Join(dxvkX64, "d3d9.dll"), filepath.Join(system32, "d3d9.dll"))
		copyFile(filepath.Join(dxvkX64, "d3d10core.dll"), filepath.Join(system32, "d3d10core.dll"))
		copyFile(filepath.Join(dxvkX64, "d3d11.dll"), filepath.Join(system32, "d3d11.dll"))
		
		copyFile(filepath.Join(dxvkX32, "dxgi.dll"), filepath.Join(syswow64, "dxgi.dll"))
		copyFile(filepath.Join(dxvkX32, "d3d9.dll"), filepath.Join(syswow64, "d3d9.dll"))
		copyFile(filepath.Join(dxvkX32, "d3d10core.dll"), filepath.Join(syswow64, "d3d10core.dll"))
		copyFile(filepath.Join(dxvkX32, "d3d11.dll"), filepath.Join(syswow64, "d3d11.dll"))
	}

	// Inject D3DMetal if selected
	if engineType == "d3dmetal" || engineType == "gptk" {
		gptkX64 := filepath.Join(a.EnginesDir(), "gptk", "lib", "wine", "x86_64-windows")
		gptkX32 := filepath.Join(a.EnginesDir(), "gptk", "lib", "wine", "i386-windows")
		
		copyFile(filepath.Join(gptkX64, "dxgi.dll"), filepath.Join(system32, "dxgi.dll"))
		copyFile(filepath.Join(gptkX64, "d3d11.dll"), filepath.Join(system32, "d3d11.dll"))
		copyFile(filepath.Join(gptkX64, "d3d12.dll"), filepath.Join(system32, "d3d12.dll"))

		copyFile(filepath.Join(gptkX32, "dxgi.dll"), filepath.Join(syswow64, "dxgi.dll"))
		copyFile(filepath.Join(gptkX32, "d3d11.dll"), filepath.Join(syswow64, "d3d11.dll"))
		copyFile(filepath.Join(gptkX32, "d3d12.dll"), filepath.Join(syswow64, "d3d12.dll"))
	}

	// Inject dgVoodoo2 for D3D9 fallback support on macOS (MoltenVK does not support upstream D3D9 Vulkan extensions)
	if runtime.GOOS == "darwin" {
		dgVoodooX64 := filepath.Join(a.EnginesDir(), "dgvoodoo", "x64")
		dgVoodooX32 := filepath.Join(a.EnginesDir(), "dgvoodoo", "x32")
		
		copyFile(filepath.Join(dgVoodooX64, "d3d9.dll"), filepath.Join(system32, "d3d9.dll"))
		copyFile(filepath.Join(dgVoodooX32, "d3d9.dll"), filepath.Join(syswow64, "d3d9.dll"))
	}

	// Always inject helper DLLs (d3dcompiler, d3dx9, d3dx10, d3dx11) from GPTK since Wine needs them
	if engineType == "dxvk" || engineType == "d3dmetal" || engineType == "gptk" {
		gptkX64 := filepath.Join(a.EnginesDir(), "gptk", "lib", "wine", "x86_64-windows")
		gptkX32 := filepath.Join(a.EnginesDir(), "gptk", "lib", "wine", "i386-windows")
		
		helpers := []string{"d3dcompiler_*.dll", "d3dx9_*.dll", "d3dx10_*.dll", "d3dx11_*.dll"}
		for _, pattern := range helpers {
			matchesX64, _ := filepath.Glob(filepath.Join(gptkX64, pattern))
			for _, match := range matchesX64 {
				copyFile(match, filepath.Join(system32, filepath.Base(match)))
			}
			matchesX32, _ := filepath.Glob(filepath.Join(gptkX32, pattern))
			for _, match := range matchesX32 {
				copyFile(match, filepath.Join(syswow64, filepath.Base(match)))
			}
		}
	}
}

func (a *App) IsGameMacNative(gameID, platform, appID, gameDir string) bool {
	configPath := filepath.Join(a.RiftDir(), "config.json")
	var cfg AppConfig
	if data, err := os.ReadFile(configPath); err == nil {
		json.Unmarshal(data, &cfg)
	}

	if cfg.ImportedGames != nil {
		if importedPath, exists := cfg.ImportedGames[gameID]; exists && importedPath != "" {
			return true
		}
	}

	if strings.EqualFold(platform, "epic") {
		homeDir, _ := os.UserHomeDir()
		manifestPaths := []string{
			filepath.Join(homeDir, ".config", "legendary", "installed.json"),
			filepath.Join(homeDir, ".rift", "auth", "epic", "installed.json"),
			filepath.Join(gameDir, "..", "installed.json"),
			filepath.Join(gameDir, "installed.json"),
		}
		epicGames := parsers.ParseLegendaryInstalled(manifestPaths)
		for appKey, g := range epicGames {
			match := (appKey == appID || g.AppName == appID || strings.EqualFold(g.Title, gameID))
			if !match && g.InstallPath != "" && gameDir != "" {
				rel, err := filepath.Rel(filepath.Clean(g.InstallPath), filepath.Clean(gameDir))
				if err == nil && (rel == "." || rel == "") {
					match = true
				}
			}
			if match {
				if g.Executable != "" && !strings.HasSuffix(strings.ToLower(g.Executable), ".exe") {
					return true
				}
			}
		}
	} else if strings.EqualFold(platform, "steam") {
		hasAppBundle := false
		filepath.Walk(gameDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && info.IsDir() && strings.HasSuffix(info.Name(), ".app") {
				hasAppBundle = true
				return filepath.SkipDir
			}
			return nil
		})
		return hasAppBundle
	}
	return false
}

func (a *App) acquirePipeline(gameID, platform, appID string, forceWindows bool) {
	a.logInfo("[Epic] Epic game downloading is currently offline.")
}

// ForceQuitGame attempts to cleanly kill the specific game's process or wine prefix.
func (a *App) ForceQuitGame(gameID string) string {
	a.logInfo("[Launch] ForceQuitGame requested for %s", gameID)
	
	prefixDir := filepath.Join(a.ResolveGameDir(gameID), "prefix")
	if _, err := os.Stat(prefixDir); err == nil {
		wineBinary := filepath.Join(a.EnginesDir(), "wine", "bin", "wineserver")
		if _, err := os.Stat(wineBinary); os.IsNotExist(err) {
			wineBinary = filepath.Join(a.EnginesDir(), "gptk", "bin", "wineserver")
		}
		
		cmd := exec.Command(wineBinary, "-k")
		cmd.Env = append(engine.SafeEnviron(), "WINEPREFIX="+prefixDir)
		cmd.Run()
	}

	a.emitEvent("game_process_exited", gameID)
	return "Force quit command sent."
}

// ForceQuitAll unconditionally kills all wine and translation processes globally.
func (a *App) ForceQuitAll() string {
	a.logInfo("[Launch] ForceQuitAll requested")
	exec.Command("killall", "-9", "wine64-preloader", "wine64", "wine-preloader", "wine", "wineserver").Run()
	
	// Also clear active games
	sysmon.ClearActiveGameAndSave(filepath.Join(a.RiftDir(), "logs"))
	return "Global force kill executed."
}
