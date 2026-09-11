package rift

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"rift/internal/engine"
	"rift/internal/sysmon"
	"rift/internal/types"
)

// LaunchSteamTranslated launches Windows-native Steam games via Wine/GPTK/DXVK.
func (a *App) LaunchSteamTranslated(gameID, appID, gameDir string, rt *types.RuntimeConfig) string {
	a.logInfo("[Launch] ===== STEAM TRANSLATED LAUNCH: gameID=%s appID=%s =====", gameID, appID)
	launchStartTime := time.Now()

	home, _ := os.UserHomeDir()

	a.emitEvent("launch_status", map[string]interface{}{
		"gameId": gameID,
		"status": "Configuring Sandboxed Wine Capsule...",
	})
	winePrefix := a.ConfigureGamePrefix(gameID)
	
	// Pre-requisites and setup
	a.emitEvent("launch_status", map[string]interface{}{
		"gameId": gameID,
		"status": "Verifying Essential VC++ & Wine Runtimes...",
	})
	a.EnsureEssentialRuntimes(a.EnginesDir(), winePrefix)

	runtimePath := filepath.Join(a.ResolveGameDir(gameID), "game_runtime.json")
	shouldCurate := true
	if data, err := os.ReadFile(runtimePath); err == nil {
		var rconf map[string]interface{}
		if err := json.Unmarshal(data, &rconf); err == nil {
			if curated, ok := rconf["ai_curated"].(bool); ok && curated {
				shouldCurate = false
			}
			if manual, ok := rconf["manual_override"].(bool); ok && manual {
				shouldCurate = false
			}
		}
	}
	if shouldCurate {
		a.runAICuration(gameID, appID, winePrefix)
	}

	// Workaround for wrap_oal.dll: Creative's legacy OpenAL wrapper uses VirtualAlloc in a manner
	// that triggers fatal Wine memory assertions (alloc_pages_vprot) on Apple Silicon.
	wrapOal := filepath.Join(gameDir, "wrap_oal.dll")
	if _, err := os.Stat(wrapOal); err == nil {
		_ = os.Rename(wrapOal, wrapOal+".disabled")
	}

	var exePath string
	targetExeName := rt.ExecutablePath

	// TrackMania Nations Forever: Launcher executable crashes or pops setup dialogs.
	// Always launch TmForever.exe directly with /useexedir.
	if appID == "11020" || strings.Contains(strings.ToLower(targetExeName), "tmforever") {
		if strings.Contains(strings.ToLower(targetExeName), "launcher") || targetExeName == "" {
			targetExeName = "TmForever.exe"
		}
	}

	if targetExeName == "" {
		if data, err := os.ReadFile(runtimePath); err == nil {
			var rconf map[string]interface{}
			if err := json.Unmarshal(data, &rconf); err == nil {
				if relExe, ok := rconf["executable"].(string); ok && relExe != "" {
					targetExeName = relExe
				}
			}
		}
	}

	if appID == "11020" || strings.Contains(strings.ToLower(targetExeName), "tmforever") {
		if strings.Contains(strings.ToLower(targetExeName), "launcher") {
			targetExeName = "TmForever.exe"
		}
	}

	if targetExeName != "" {
		if engine.IsNonGameBinary(targetExeName) {
			a.logWarn("[Launch] Target executable %q is a known utility/installer binary, re-resolving...", targetExeName)
			targetExeName = ""
		} else {
			if filepath.IsAbs(targetExeName) {
				exePath = targetExeName
			} else {
				exePath = filepath.Join(gameDir, targetExeName)
				if _, err := os.Stat(exePath); os.IsNotExist(err) {
					baseCandidate := filepath.Join(gameDir, filepath.Base(targetExeName))
					if _, err := os.Stat(baseCandidate); err == nil {
						exePath = baseCandidate
					} else {
						prefixCandidate := filepath.Join(winePrefix, targetExeName)
						if _, err := os.Stat(prefixCandidate); err == nil {
							exePath = prefixCandidate
						}
					}
				}
			}
			if _, err := os.Stat(exePath); os.IsNotExist(err) {
				exePath = ""
			}
		}
	}

	if exePath == "" {
		a.logInfo("[Launch] Resolving executable via manifest-driven pipeline for %s...", gameID)
		a.emitEvent("launch_status", map[string]interface{}{
			"gameId": gameID,
			"status": "Resolving best game executable...",
		})
		bestExe, candidates, err := engine.ResolveGameExecutable(gameDir, appID, gameID, a.SteamCMDDir())
		if err == nil && bestExe != "" {
			exePath = filepath.Join(gameDir, bestExe)
			
			// Persist resolved executable
			var runtimeMap map[string]interface{}
			if rdata, rerr := os.ReadFile(runtimePath); rerr == nil {
				json.Unmarshal(rdata, &runtimeMap)
			}
			if runtimeMap == nil {
				runtimeMap = make(map[string]interface{})
			}
			runtimeMap["id"] = gameID
			runtimeMap["appID"] = appID
			runtimeMap["platform"] = "steam"
			runtimeMap["installed"] = true
			runtimeMap["executable"] = bestExe
			runtimeMap["executable_path"] = bestExe
			if b, err := json.MarshalIndent(runtimeMap, "", "  "); err == nil {
				os.WriteFile(runtimePath, b, 0644)
			}
		} else if len(candidates) > 0 {
			a.emitEvent("exe_pick_required", map[string]interface{}{
				"gameId":     gameID,
				"candidates": candidates,
			})
			return "Please select game executable in the dialog."
		}
	}

	if exePath == "" {
		return fmt.Sprintf("Error: Executable was not found in the game folder: %s. Is the game installed?", gameDir)
	} else {
		a.logInfo("[Launch] Found executable at %s", exePath)
	}

	// Check if the executable is 32-bit (i386). If so, delegate to the dedicated 32-bit WoW64 pipeline.
	if is32BitExecutable(exePath) {
		a.logInfo("[Launch] Detected 32-bit Windows executable. Delegating to dedicated 32-bit WoW64 pipeline.")
		return a.launchSteamTranslated32(gameID, appID, rt, exePath, gameDir, winePrefix)
	}

	// Engine resolution
	var wineBinary string
	gptkDir := filepath.Join(a.EnginesDir(), "gptk")
	sharedWineDir := filepath.Join(a.EnginesDir(), "wine")
	switch rt.Engine {
	case "d3dmetal", "gptk":
		wineBinary = filepath.Join(gptkDir, "bin", "wine64")
	case "dxvk":
		wineBinary = filepath.Join(sharedWineDir, "bin", "wine")
	default:
		wineBinary = filepath.Join(sharedWineDir, "bin", "wine")
	}

	// Pre-flight: verify the Wine binary actually exists on disk
	if _, err := os.Stat(wineBinary); os.IsNotExist(err) {
		a.logError("[Launch] Wine binary not found at %s. Engine '%s' is not installed.", wineBinary, rt.Engine)
		a.emitEvent("launch_failed", map[string]interface{}{
			"gameId": gameID,
			"error":  fmt.Sprintf("Translation engine '%s' is not installed. Open Settings → Engine Setup to install it.", rt.Engine),
		})
		return fmt.Sprintf("Error: Translation engine '%s' is not installed.", rt.Engine)
	}

	// Apply winetricks
	if len(rt.Winetricks) > 0 {
		marker := filepath.Join(a.ResolveGameDir(gameID), ".winetricks_applied")
		missingRequiredDLL := false
		for _, dep := range rt.Winetricks {
			if strings.HasPrefix(strings.ToLower(dep), "vcrun") {
				vcDLL := filepath.Join(winePrefix, "drive_c", "windows", "system32", "vcruntime140.dll")
				if _, err := os.Stat(vcDLL); os.IsNotExist(err) {
					missingRequiredDLL = true
					os.Remove(marker)
					break
				}
			}
		}

		if data, err := os.ReadFile(marker); err == nil && strings.Join(rt.Winetricks, ",") == string(data) && !missingRequiredDLL {
			a.logInfo("[Launch] Winetricks dependencies already applied: %v", rt.Winetricks)
		} else {
			a.logInfo("[Launch] Applying winetricks deps %v to prefix %s", rt.Winetricks, winePrefix)
			a.emitEvent("launch_status", map[string]interface{}{
				"gameId": gameID,
				"status": "Injecting C++ DLLs so you never have to open Terminal...",
			})
			a.applyWinetricksDependencies(wineBinary, winePrefix, rt.Winetricks)
			os.WriteFile(marker, []byte(strings.Join(rt.Winetricks, ",")), 0644)
		}
	}

	a.emitEvent("launch_status", map[string]interface{}{
		"gameId": gameID,
		"status": "Bribing DirectX to speak Apple Metal...",
	})

	// Determine Steam Path (Option 2: Direct Prefix Path vs Z: escape)
	// At this point, the download pipeline will place steam games inside prefix/drive_c/Program Files (x86)/Steam
	unixDriveC := filepath.Join(winePrefix, "drive_c")
	var winPath string
	if strings.HasPrefix(exePath, unixDriveC) {
		relPath, _ := filepath.Rel(unixDriveC, exePath)
		winPath = `C:\` + strings.ReplaceAll(relPath, "/", `\`)
	} else {
		// Fallback just in case old capsule format
		winPath = "Z:" + strings.ReplaceAll(exePath, "/", `\`)
	}

	launchArgs := []string{winPath}
	if _, err := os.Stat(filepath.Join(filepath.Dir(exePath), "steam.dll")); err == nil {
		launchArgs = append(launchArgs, "-steam")
	}
	launchArgs = append(launchArgs, rt.LaunchArgs...)

	a.emitEvent("launch_status", map[string]interface{}{
		"gameId": gameID,
		"status": "Activating Translation Layer (D3DMetal / DXVK)...",
	})
	a.InjectEngineDLLs(winePrefix, rt.Engine)
	
	// Removed toxic macOS Wine3D legacy registry overrides (UseGLSL=disabled) which cause silent hard-freezes on Apple Silicon.

	// Failsafe: Kill any stale wineserver instances to prevent ESync shared memory crashes
	wineServerBin := filepath.Join(filepath.Dir(wineBinary), "wineserver")
	cleanup := exec.Command(wineServerBin, "-k")
	cleanup.Env = append(engine.SafeEnviron(), "WINEPREFIX="+winePrefix)
	_ = cleanup.Run()

	a.emitEvent("launch_status", map[string]interface{}{
		"gameId": gameID,
		"status": "Launching Game Executable...",
	})

	cmd := exec.Command(wineBinary, launchArgs...)
	cmd.Dir = filepath.Dir(exePath)

	// Setup Env
	envMap := make(map[string]string)
	for _, e := range engine.SafeEnviron() {
		if idx := strings.Index(e, "="); idx != -1 {
			envMap[e[:idx]] = e[idx+1:]
		}
	}
	envMap["HOME"] = home
	envMap["WINEPREFIX"] = winePrefix

	for k, v := range rt.EnvVars {
		envMap[k] = v
	}

	var overrides []string
	for dll, mode := range rt.DLLOverrides {
		overrides = append(overrides, dll+"="+mode)
	}

	if rt.Engine == "dxvk" {
		overrides = append(overrides, "d3d9=n,b", "d3d11=n,b", "d3d10core=n,b")
		// DXVK-macOS may not ship standalone dxgi.dll — Wine's builtin dxgi delegates to DXVK's d3d11.dll
		dxvkDxgi := filepath.Join(a.EnginesDir(), "dxvk", "x64", "dxgi.dll")
		if _, err := os.Stat(dxvkDxgi); err == nil {
			overrides = append(overrides, "dxgi=n,b")
		} else {
			overrides = append(overrides, "dxgi=b")
		}
	} else if rt.Engine == "d3dmetal" || rt.Engine == "gptk" {
		overrides = append(overrides, "d3d9=n,b", "d3d11=n,b", "d3d12=n,b")
		// D3DMetal ships its own dxgi — check for it
		gptkDxgi := filepath.Join(a.EnginesDir(), "gptk", "lib", "external", "D3DMetal.framework")
		if _, err := os.Stat(gptkDxgi); err == nil {
			overrides = append(overrides, "dxgi=n,b")
		} else {
			overrides = append(overrides, "dxgi=b")
		}
	} else if rt.Engine == "wine3d" {
		// SAFETY GUARD: If a prefix was previously infected by DXVK/D3DMetal, native DLLs will be sitting in the C: drive.
		// We MUST force 'builtin' (Wine3D OpenGL) to completely ignore any lingering DXVK dlls.
		overrides = append(overrides, "d3d9=b", "d3d11=b", "d3d10core=b", "dxgi=b")
	}

	overrides = append(overrides, "steam_api=n,b", "steam_api64=n,b", "steamclient=n,b", "steamclient64=n,b")

	if len(overrides) > 0 {
		envMap["WINEDLLOVERRIDES"] = strings.Join(overrides, ";")
	}

	envMap["WINEDEBUG"] = "err+all,fixme+all" // Enabled verbose Wine logging
	envMap["MTL_SHADER_VALIDATION"] = "0"
	envMap["MTL_DEBUG_LAYER"] = "0"

	shaderCacheDir := filepath.Join(a.ResolveGameDir(gameID), "shader_cache")
	os.MkdirAll(shaderCacheDir, 0755)
	envMap["DXVK_STATE_CACHE_PATH"] = shaderCacheDir
	envMap["DXVK_STATE_CACHE"] = "1"
	envMap["MESA_SHADER_CACHE_DIR"] = shaderCacheDir
	envMap["__GL_SHADER_DISK_CACHE_PATH"] = shaderCacheDir

	// === RIFT SHADER PERFORMANCE ENGINE ===
	// Generate optimized dxvk.conf for DXVK games
	if rt.Engine == "dxvk" {
		if err := engine.GenerateDXVKConf(filepath.Dir(exePath), runtime.NumCPU()); err != nil {
			a.logWarn("[Shader] Failed to generate dxvk.conf: %v", err)
		} else {
			a.logInfo("[Shader] Generated optimized dxvk.conf in %s (%d compiler threads)", filepath.Dir(exePath), engine.OptimalCompilerThreads())
		}
		// Force DXVK_ASYNC=1 always — this is the single highest-impact setting
		envMap["DXVK_ASYNC"] = "1"
		envMap["DXVK_LOG_LEVEL"] = "none"
		envMap["DXVK_NUM_COMPILER_THREADS"] = fmt.Sprintf("%d", engine.OptimalCompilerThreads())
	}

	// Restore backed-up D3DMetal shader cache (prevents re-compilation after macOS cache purge)
	if rt.Engine == "d3dmetal" || rt.Engine == "gptk" {
		if err := engine.RestoreD3DMetalCache(a.ResolveGameDir(gameID), filepath.Base(exePath)); err != nil {
			a.logWarn("[Shader] D3DMetal cache restore skipped: %v", err)
		} else {
			a.logInfo("[Shader] D3DMetal shader cache restored for %s", filepath.Base(exePath))
		}
		envMap["D3DM_SHADER_CACHE"] = "1"
	}

	retinaVal := "n"
	if rt.RetinaMode {
		retinaVal = "y"
	}
	regRetina := exec.Command(wineBinary, "reg", "add", `HKEY_CURRENT_USER\Software\Wine\Mac Driver`, "/v", "RetinaMode", "/d", retinaVal, "/f")
	regRetina.Env = append(engine.SafeEnviron(), "WINEPREFIX="+winePrefix, "WINEDEBUG=-all")
	_ = regRetina.Run()

	var envSlice []string
	for k, v := range envMap {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
	}

	// === ENGINE ENVIRONMENT INJECTION ===
	if rt.Engine == "d3dmetal" || rt.Engine == "gptk" {
		if rt.MetalHUD {
			envSlice = append(envSlice, "MTL_HUD_ENABLED=1")
		}
		if rt.MetalTrace {
			envSlice = append(envSlice, "MTL_CAPTURE_ENABLED=1")
		}
		if rt.DXREnabled {
			envSlice = append(envSlice, "D3DMetal_DXR=1")
		}
	} else if rt.Engine == "dxvk" {
		if rt.DXVKHUD != "" && rt.DXVKHUD != "off" {
			switch rt.DXVKHUD {
			case "full":
				envSlice = append(envSlice, "DXVK_HUD=full")
			case "frametimes":
				envSlice = append(envSlice, "DXVK_HUD=devinfo,fps,frametimes")
			case "fps":
				envSlice = append(envSlice, "DXVK_HUD=fps")
			default:
				envSlice = append(envSlice, "DXVK_HUD="+rt.DXVKHUD)
			}
		}
		// DXVK_ASYNC is ON by default (injected in envMap above).
		// User can explicitly disable it via the Configuration toggle.
		if !rt.DXVKAsync {
			envSlice = append(envSlice, "DXVK_ASYNC=0")
		}
		if rt.DXVKFrameRate > 0 {
			envSlice = append(envSlice, fmt.Sprintf("DXVK_FRAME_RATE=%d", rt.DXVKFrameRate))
		}
	}

	if rt.AVXEnabled {
		envSlice = append(envSlice, "ROSETTA_ADVERTISE_AVX=1")
	}

	switch rt.EnhancedSync {
	case "esync":
		envSlice = append(envSlice, "WINEESYNC=1", "WINEMSYNC=0")
	case "msync":
		envSlice = append(envSlice, "WINEMSYNC=1", "WINEESYNC=0")
	case "none":
		envSlice = append(envSlice, "WINEESYNC=0", "WINEMSYNC=0")
	default:
		if rt.Engine == "dxvk" {
			envSlice = append(envSlice, "WINEESYNC=1", "WINEMSYNC=0")
		} else {
			envSlice = append(envSlice, "WINEMSYNC=1", "WINEESYNC=0")
		}
	}

	cmd.Env = envSlice
	
	// Stream Wine output directly to the Wails terminal for debugging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	go func() {
		a.logInfo("[Launch] Starting translated Steam app: %s", exePath)
		a.emitEvent("launch_complete", gameID)
		sysProfile := GetMacHardwareSpecs()
		sysProfile["Engine"] = rt.Engine
		realGameName := rt.GameName
		if realGameName == "" {
			realGameName = queryGameNameFromDB(appID)
		}
		if realGameName == "" {
			realGameName = queryGameNameFromDB(gameID)
		}
		if realGameName == "" {
			realGameName = gameID
		}

		launchDurationMs := int(time.Since(launchStartTime).Milliseconds())
		a.MonitorGame(gameID, appID, realGameName, exePath, cmd, sysProfile, rt, launchDurationMs)

		// === POST-EXIT: Backup D3DMetal shader cache to survive macOS cache purges ===
		if rt.Engine == "d3dmetal" || rt.Engine == "gptk" {
			if err := engine.BackupD3DMetalCache(a.ResolveGameDir(gameID), filepath.Base(exePath)); err != nil {
				a.logWarn("[Shader] D3DMetal cache backup failed: %v", err)
			} else {
				a.logInfo("[Shader] D3DMetal shader cache backed up for %s", filepath.Base(exePath))
			}
		}

		sysmon.ClearActiveGameAndSave(filepath.Join(home, ".rift", "logs"))
		a.emitEvent("game_process_exited", gameID)
	}()

	return "Steam Translated Launch Executed."
}
