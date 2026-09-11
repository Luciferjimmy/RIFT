package rift

import (
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

// launchEpicTranslated32 handles dedicated execution of 32-bit Windows executables for Epic Games.
// On Apple Silicon macOS, 32-bit games run via Wine 11 WoW64.
// Apple's D3DMetal does NOT support 32-bit binaries, so 32-bit games cleanly use DXVK (Vulkan over Metal).
func (a *App) launchEpicTranslated32(gameID, appID string, rt *types.RuntimeConfig, exePath, gameDir, winePrefix, targetAppName string, legGameParams, legEGLParams []string, legendaryConfig string) string {
	a.logInfo("[Launch 32-bit] Initializing dedicated 32-bit Epic launch pipeline for %s (appID: %s)", gameID, appID)
	launchStartTime := time.Now()

	home, _ := os.UserHomeDir()

	// Apple D3DMetal is strictly 64-bit. Fallback cleanly to DXVK.
	if rt.Engine == "d3dmetal" || rt.Engine == "gptk" {
		a.logInfo("[Launch 32-bit] Note: D3DMetal is 64-bit only on Apple Silicon. Using DXVK for 32-bit Epic game.")
		rt.Engine = "dxvk"
	} else if rt.Engine == "" {
		rt.Engine = "dxvk"
	}

	// Always use shared Wine 11 WoW64 binary for 32-bit execution
	wineBinary := filepath.Join(a.EnginesDir(), "wine", "bin", "wine")
	if _, err := os.Stat(wineBinary); err != nil {
		wineBinary = filepath.Join(a.EnginesDir(), "gptk", "bin", "wine64")
	}

	// Apply winetricks if specified
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
			a.logInfo("[Launch 32-bit] Winetricks dependencies already applied: %v", rt.Winetricks)
		} else {
			a.logInfo("[Launch 32-bit] Applying winetricks deps %v to prefix %s", rt.Winetricks, winePrefix)
			a.emitEvent("launch_status", map[string]interface{}{
				"gameId": gameID,
				"status": fmt.Sprintf("Applying Winetricks Dependencies: %v...", rt.Winetricks),
			})
			a.applyWinetricksDependencies(wineBinary, winePrefix, rt.Winetricks)
			os.WriteFile(marker, []byte(strings.Join(rt.Winetricks, ",")), 0644)
		}
	}

	launchArgs := []string{exePath}
	launchArgs = append(launchArgs, rt.LaunchArgs...)
	launchArgs = append(launchArgs, legGameParams...)
	launchArgs = append(launchArgs, legEGLParams...)

	// Inject 32-bit engine DLLs into syswow64 and game directory
	a.emitEvent("launch_status", map[string]interface{}{
		"gameId": gameID,
		"status": "Activating Translation Layer (D3DMetal / DXVK)...",
	})
	a.inject32BitEngineDLLs(winePrefix, gameDir, rt.Engine)

	// Failsafe: Terminate any lingering wineserver instances
	wineServerBin := filepath.Join(filepath.Dir(wineBinary), "wineserver")
	cleanup := exec.Command(wineServerBin, "-k")
	cleanup.Env = append(engine.SafeEnviron(), "WINEPREFIX="+winePrefix)
	_ = cleanup.Run()

	cmd := exec.Command(wineBinary, launchArgs...)
	cmd.Dir = filepath.Dir(exePath)

	// Setup Environment Map
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

	// Safely detect if the 32-bit executable was compiled with /LARGEADDRESSAWARE (0x0020).
	if engine.IsLargeAddressAware(exePath) {
		a.logInfo("[Launch 32-bit] Executable has LARGE_ADDRESS_AWARE flag set. Enabling 4GB virtual address space.")
		envMap["WINE_LARGE_ADDRESS_AWARE"] = "1"
	} else {
		a.logInfo("[Launch 32-bit] Executable is NOT Large Address Aware. Constraining to safe 2GB virtual address space.")
		envMap["WINE_LARGE_ADDRESS_AWARE"] = "0"
	}

	overrides := []string{}
	for dll, mode := range rt.DLLOverrides {
		overrides = append(overrides, dll+"="+mode)
	}

	if rt.Engine == "dxvk" {
		// 32-bit graphics translation pipeline on Apple Silicon (Wine WoW64):
		// - dgVoodoo2 translates Direct3D 9, Direct3D 8, and DirectDraw calls to Direct3D 11
		// - DXVK-macOS translates Direct3D 11 calls into Vulkan / MoltenVK / Metal
		// - Native d3dx9 math/shader helpers are prioritized over Wine's OpenGL-based builtins
		overrides = append(overrides,
			"d3d9=n,b",
			"ddraw=n,b",
			"d3dimm=n,b",
			"d3d8=n,b",
			"d3d11=n,b",
			"d3d10core=n,b",
			"dxgi=b",
			"*d3dx9*=n,b",
			"d3dx9_24=n,b", "d3dx9_25=n,b", "d3dx9_26=n,b", "d3dx9_27=n,b", "d3dx9_28=n,b",
			"d3dx9_29=n,b", "d3dx9_30=n,b", "d3dx9_31=n,b", "d3dx9_32=n,b", "d3dx9_33=n,b",
			"d3dx9_34=n,b", "d3dx9_35=n,b", "d3dx9_36=n,b", "d3dx9_37=n,b", "d3dx9_38=n,b",
			"d3dx9_39=n,b", "d3dx9_40=n,b", "d3dx9_41=n,b", "d3dx9_42=n,b", "d3dx9_43=n,b",
		)
		dxvkDxgi := filepath.Join(a.EnginesDir(), "dxvk", "x32", "dxgi.dll")
		if _, err := os.Stat(dxvkDxgi); err == nil {
			overrides = append(overrides, "dxgi=n,b")
		}
	} else if rt.Engine == "wine3d" {
		overrides = append(overrides, "d3d9=b", "d3d11=b", "d3d10core=b", "dxgi=b")
	}

	if len(overrides) > 0 {
		envMap["WINEDLLOVERRIDES"] = strings.Join(overrides, ";")
	}

	envMap["WINEDEBUG"] = "err+all,fixme-all"
	envMap["MTL_SHADER_VALIDATION"] = "0"
	envMap["MTL_DEBUG_LAYER"] = "0"

	shaderCacheDir := filepath.Join(a.ResolveGameDir(gameID), "shader_cache")
	os.MkdirAll(shaderCacheDir, 0755)
	envMap["DXVK_STATE_CACHE_PATH"] = shaderCacheDir
	envMap["DXVK_STATE_CACHE"] = "1"
	envMap["MESA_SHADER_CACHE_DIR"] = shaderCacheDir
	envMap["__GL_SHADER_DISK_CACHE_PATH"] = shaderCacheDir

	if rt.Engine == "dxvk" {
		if err := engine.GenerateDXVKConf(filepath.Dir(exePath), runtime.NumCPU()); err != nil {
			a.logWarn("[Launch 32-bit] Failed to generate dxvk.conf: %v", err)
		}
		envMap["DXVK_ASYNC"] = "1"
		envMap["DXVK_LOG_LEVEL"] = "none"
		envMap["DXVK_NUM_COMPILER_THREADS"] = fmt.Sprintf("%d", engine.OptimalCompilerThreads())
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

	if rt.Engine == "dxvk" {
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
		envSlice = append(envSlice, "WINEESYNC=1", "WINEMSYNC=0")
	}

	cmd.Env = envSlice
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	a.emitEvent("launch_status", map[string]interface{}{
		"gameId": gameID,
		"status": "Launching Game Executable...",
	})

	go func() {
		a.logInfo("[Launch 32-bit] Executing 32-bit translated Epic game: %s", exePath)
		a.emitEvent("launch_complete", gameID)
		sysProfile := GetMacHardwareSpecs()
		sysProfile["Engine"] = rt.Engine
		sysProfile["Architecture"] = "x86_32"
		sysProfile["WineVersion"] = "RIFT-Wine-11.0-WoW64"

		launchDurationMs := int(time.Since(launchStartTime).Milliseconds())
		a.MonitorGame(gameID, appID, targetAppName, exePath, cmd, sysProfile, rt, launchDurationMs)

		sysmon.ClearActiveGameAndSave(filepath.Join(home, ".rift", "logs"))
		a.emitEvent("game_process_exited", gameID)
	}()

	return "Epic 32-bit Translated Launch Executed."
}
