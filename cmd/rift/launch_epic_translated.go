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

// LaunchEpicTranslated launches Windows-native Epic games via Wine/GPTK/DXVK.
func (a *App) LaunchEpicTranslated(gameID, appID, gameDir string, rt *types.RuntimeConfig) string {
	a.logInfo("[Launch] ===== EPIC TRANSLATED LAUNCH: gameID=%s appID=%s =====", gameID, appID)
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
	a.runAICuration(gameID, appID, winePrefix)

	targetAppName := appID
	if targetAppName == "" {
		targetAppName = strings.TrimPrefix(gameID, "epic-")
	}

	var exePath string
	targetExeName := rt.ExecutablePath

	runtimePath := filepath.Join(a.ResolveGameDir(gameID), "game_runtime.json")
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

	if targetExeName != "" {
		if engine.IsNonGameBinary(targetExeName) {
			a.logWarn("[Launch] Target executable %q is a known utility/installer binary, deferring to manifest...", targetExeName)
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
		a.logInfo("[Launch] Skipping early executable check for Epic game; deferring to Legendary manifest...")
	} else {
		a.logInfo("[Launch] Found executable at %s", exePath)
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

	// Epic specifics
	legendaryBin := filepath.Join(a.EnginesDir(), "legendary", "venv", "bin", "legendary")
	if _, err := os.Stat(legendaryBin); err != nil {
		if path, err := exec.LookPath("legendary"); err == nil {
			legendaryBin = path
		}
	}

	epicGlobalAuth := filepath.Join(a.UserAuthDir(), "epic")
	legendaryConfig := filepath.Join(a.ResolveGameDir(gameID), ".legendary")
	os.MkdirAll(legendaryConfig, 0755)
	for _, name := range []string{"user.json", "config.ini"} {
		src := filepath.Join(epicGlobalAuth, name)
		dest := filepath.Join(legendaryConfig, name)
		if _, err := os.Stat(dest); os.IsNotExist(err) {
			if data, err := os.ReadFile(src); err == nil {
				os.WriteFile(dest, data, 0600)
			}
		}
	}

	var legGameParams []string
	var legEGLParams []string
	legCmd := exec.Command(legendaryBin, "launch", targetAppName, "--no-wine", "--json")
	legCmd.Env = append(engine.SafeEnviron(), "LEGENDARY_CONFIG_PATH="+legendaryConfig)
	outBytes, err := legCmd.Output()
	if err != nil {
		legCmdFallback := exec.Command(legendaryBin, "launch", targetAppName, "--no-wine", "--json")
		legCmdFallback.Env = append(engine.SafeEnviron(), "LEGENDARY_CONFIG_PATH="+epicGlobalAuth)
		outBytes, err = legCmdFallback.Output()
	}

	if err == nil {
		lines := strings.Split(string(outBytes), "\n")
		var jsonLine string
		for i := len(lines) - 1; i >= 0; i-- {
			if strings.HasPrefix(strings.TrimSpace(lines[i]), "{") {
				jsonLine = lines[i]
				break
			}
		}
		var legData struct {
			GameParams    []string `json:"game_parameters"`
			EGLParams     []string `json:"egl_parameters"`
			GameExe       string   `json:"game_executable"`
			GameDirectory string   `json:"game_directory"`
		}
		if err := json.Unmarshal([]byte(jsonLine), &legData); err == nil {
			legGameParams = legData.GameParams
			legEGLParams = legData.EGLParams
			a.logInfo("[Launch] Legendary generated %d game params and %d EGL params for %s", len(legGameParams), len(legEGLParams), targetAppName)
			if exePath == "" && legData.GameExe != "" && legData.GameDirectory != "" && !engine.IsNonGameBinary(legData.GameExe) {
				exeCandidate := filepath.Join(legData.GameDirectory, legData.GameExe)
				if _, statErr := os.Stat(exeCandidate); statErr == nil {
					exePath = exeCandidate
				}
			}
		}
	} else {
		a.logWarn("[Launch] Legendary auth query failed or offline: %v", err)
	}

	// OFFLINE FALLBACK: If Legendary failed (e.g. offline, expired auth, DNS drop), scan game files directly on disk
	if exePath == "" {
		a.logInfo("[Launch] Legendary exe query unavailable; falling back to direct disk scanner in %s...", gameDir)
		gameName := queryGameNameFromDB(appID)
		bestExe, candidates, err := engine.ResolveGameExecutable(gameDir, appID, gameName, a.SteamCMDDir())
		if err == nil && bestExe != "" {
			exePath = filepath.Join(gameDir, bestExe)
			a.logInfo("[Launch] Offline disk scanner resolved main executable: %s", exePath)
		} else if len(candidates) > 0 {
			a.emitEvent("exe_pick_required", map[string]interface{}{
				"gameId":     gameID,
				"candidates": candidates,
			})
			return "Please select game executable in the dialog."
		}
	}

	// HARD GUARD: Never execute Wine with an empty executable!
	if exePath == "" {
		a.logError("[Launch] Error: No valid Windows executable (.exe) found in %s. Is the game installed?", gameDir)
		a.emitEvent("launch_status", map[string]interface{}{
			"gameId": gameID,
			"status": "Executable not found. Please verify game files.",
		})
		return fmt.Sprintf("Error: Executable was not found in %s.", gameDir)
	}

	// PERMANENT CACHE: Save resolved executable to game_runtime.json so future launches (including offline) are instant!
	var rmap map[string]interface{}
	if data, err := os.ReadFile(runtimePath); err == nil {
		json.Unmarshal(data, &rmap)
	}
	if rmap == nil {
		rmap = make(map[string]interface{})
	}
	relExe, _ := filepath.Rel(gameDir, exePath)
	if relExe == "" || strings.HasPrefix(relExe, "..") {
		relExe = filepath.Base(exePath)
	}
	rmap["executable_path"] = relExe
	rmap["executable"] = relExe
	rmap["installed"] = true
	if b, err := json.MarshalIndent(rmap, "", "  "); err == nil {
		os.WriteFile(runtimePath, b, 0644)
		a.logInfo("[Launch] Saved resolved executable '%s' to game_runtime.json for offline reliability", relExe)
	}

	// Check if the executable is 32-bit (i386). If so, delegate to the dedicated 32-bit WoW64 pipeline.
	if is32BitExecutable(exePath) {
		a.logInfo("[Launch] Detected 32-bit Windows executable for Epic game. Delegating to dedicated 32-bit WoW64 pipeline.")
		return a.launchEpicTranslated32(gameID, appID, rt, exePath, gameDir, winePrefix, targetAppName, legGameParams, legEGLParams, legendaryConfig)
	}

	launchArgs := []string{exePath}
	launchArgs = append(launchArgs, rt.LaunchArgs...)
	launchArgs = append(launchArgs, legGameParams...)
	launchArgs = append(launchArgs, legEGLParams...)

	a.emitEvent("launch_status", map[string]interface{}{
		"gameId": gameID,
		"status": "Activating Translation Layer (D3DMetal / DXVK)...",
	})
	a.InjectEngineDLLs(winePrefix, rt.Engine)

	retinaVal := "n"
	if rt.RetinaMode {
		retinaVal = "y"
	}
	reg4 := exec.Command(wineBinary, "reg", "add", `HKEY_CURRENT_USER\Software\Wine\Mac Driver`, "/v", "RetinaMode", "/d", retinaVal, "/f")
	reg4.Env = append(engine.SafeEnviron(), "WINEPREFIX="+winePrefix, "WINEDEBUG=-all")
	_ = reg4.Run()

	// Failsafe: Kill any stale wineserver instances to prevent ESync shared memory crashes
	wineServerBin := filepath.Join(filepath.Dir(wineBinary), "wineserver")
	cleanup := exec.Command(wineServerBin, "-k")
	cleanup.Env = append(engine.SafeEnviron(), "WINEPREFIX="+winePrefix)
	_ = cleanup.Run()

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
	envMap["LEGENDARY_CONFIG_PATH"] = legendaryConfig

	for k, v := range rt.EnvVars {
		envMap[k] = v
	}

	overrides := []string{}
	for dll, mode := range rt.DLLOverrides {
		overrides = append(overrides, dll+"="+mode)
	}

	if rt.Engine == "dxvk" {
		overrides = append(overrides, "d3d9=n,b", "d3d11=n,b", "d3d10core=n,b")
		dxvkDxgi := filepath.Join(a.EnginesDir(), "dxvk", "x64", "dxgi.dll")
		if _, err := os.Stat(dxvkDxgi); err == nil {
			overrides = append(overrides, "dxgi=n,b")
		} else {
			overrides = append(overrides, "dxgi=b")
		}
	} else if rt.Engine == "d3dmetal" || rt.Engine == "gptk" {
		overrides = append(overrides, "d3d9=n,b", "d3d11=n,b", "d3d12=n,b")
		gptkDxgi := filepath.Join(a.EnginesDir(), "gptk", "lib", "external", "D3DMetal.framework")
		if _, err := os.Stat(gptkDxgi); err == nil {
			overrides = append(overrides, "dxgi=n,b")
		} else {
			overrides = append(overrides, "dxgi=b")
		}
	}

	if len(overrides) > 0 {
		envMap["WINEDLLOVERRIDES"] = strings.Join(overrides, ";")
	}

	envMap["WINEDEBUG"] = "-all,err+all"
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

	var envSlice []string
	for k, v := range envMap {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
	}

	// === WHISKY LOGIC INJECTION ===
	if rt.Engine == "dxvk" {
		if rt.DXVKHUD != "" && rt.DXVKHUD != "off" {
			switch rt.DXVKHUD {
			case "full":
				envSlice = append(envSlice, "DXVK_HUD=full")
			case "partial":
				envSlice = append(envSlice, "DXVK_HUD=devinfo,fps,frametimes")
			case "fps":
				envSlice = append(envSlice, "DXVK_HUD=fps")
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

	if rt.MetalHUD {
		envSlice = append(envSlice, "MTL_HUD_ENABLED=1")
	}
	if rt.MetalTrace {
		envSlice = append(envSlice, "METAL_CAPTURE_ENABLED=1")
	}
	if rt.AVXEnabled {
		envSlice = append(envSlice, "ROSETTA_ADVERTISE_AVX=1")
	}
	if rt.DXREnabled {
		envSlice = append(envSlice, "D3DM_SUPPORT_DXR=1")
	}
	if rt.ResolutionScale != "" && rt.ResolutionScale != "1.0" {
		envSlice = append(envSlice, "WINE_FULLSCREEN_FSR=1")
	}

	cmd.Env = envSlice

	a.emitEvent("launch_status", map[string]interface{}{
		"gameId": gameID,
		"status": "Launching Game Executable...",
	})

	go func() {
		a.logInfo("[Launch] Starting translated Epic app: %s", exePath)
		a.emitEvent("launch_complete", gameID)
		sysProfile := GetMacHardwareSpecs()
		sysProfile["Engine"] = rt.Engine
		launchDurationMs := int(time.Since(launchStartTime).Milliseconds())
		a.MonitorGame(gameID, appID, targetAppName, exePath, cmd, sysProfile, rt, launchDurationMs)

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

	return "Epic Translated Launch Executed."
}
