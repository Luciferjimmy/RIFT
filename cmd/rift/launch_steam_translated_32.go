package rift

import (
	"debug/pe"
	"fmt"
	"io"
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

// launchSteamTranslated32 handles dedicated execution of 32-bit Windows executables (IMAGE_FILE_MACHINE_I386).
// On Apple Silicon macOS, 32-bit games run via Wine 11 WoW64.
// Apple's D3DMetal does NOT support 32-bit binaries, so 32-bit games cleanly use DXVK (Vulkan over Metal).
func (a *App) launchSteamTranslated32(gameID, appID string, rt *types.RuntimeConfig, exePath, gameDir, winePrefix string) string {
	a.logInfo("[Launch 32-bit] Initializing dedicated 32-bit WoW64 launch pipeline for %s (appID: %s)", gameID, appID)
	launchStartTime := time.Now()

	home, _ := os.UserHomeDir()

	a.emitEvent("launch_status", map[string]interface{}{
		"gameId": gameID,
		"status": "Configuring Sandboxed Wine Capsule...",
	})

	// Binary selection for 32-bit execution:
	// When gptk or d3dmetal is requested, use GPTK's wine64 which supports 32-bit WoW64 execution on Apple Silicon.
	var wineBinary string
	if rt.Engine == "gptk" || rt.Engine == "d3dmetal" {
		wineBinary = filepath.Join(a.EnginesDir(), "gptk", "bin", "wine64")
	} else if rt.Engine == "wine" {
		wineBinary = filepath.Join(a.EnginesDir(), "wine", "bin", "wine")
	} else {
		// Default: prefer gptk if available, fallback to wine
		wineBinary = filepath.Join(a.EnginesDir(), "gptk", "bin", "wine64")
		if _, err := os.Stat(wineBinary); err != nil {
			wineBinary = filepath.Join(a.EnginesDir(), "wine", "bin", "wine")
		}
	}
	if _, err := os.Stat(wineBinary); err != nil {
		if strings.Contains(wineBinary, "gptk") {
			wineBinary = filepath.Join(a.EnginesDir(), "wine", "bin", "wine")
		} else {
			wineBinary = filepath.Join(a.EnginesDir(), "gptk", "bin", "wine64")
		}
	}

	a.emitEvent("launch_status", map[string]interface{}{
		"gameId": gameID,
		"status": "Verifying Essential VC++ & Wine Runtimes...",
	})

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

	// Determine Windows path inside prefix vs Z: escape
	unixDriveC := filepath.Join(winePrefix, "drive_c")
	var winPath string
	if strings.HasPrefix(exePath, unixDriveC) {
		relPath, _ := filepath.Rel(unixDriveC, exePath)
		winPath = `C:\` + strings.ReplaceAll(relPath, "/", `\`)
	} else {
		winPath = "Z:" + strings.ReplaceAll(exePath, "/", `\`)
	}

	launchArgs := []string{winPath}
	if _, err := os.Stat(filepath.Join(filepath.Dir(exePath), "steam.dll")); err == nil {
		launchArgs = append(launchArgs, "-steam")
	}
	launchArgs = append(launchArgs, rt.LaunchArgs...)

	// TrackMania Nations Forever requires /useexedir to read/write local configurations
	if appID == "11020" || strings.Contains(strings.ToLower(exePath), "tmforever") {
		hasUseExeDir := false
		for _, arg := range launchArgs {
			if strings.EqualFold(arg, "/useexedir") {
				hasUseExeDir = true
				break
			}
		}
		if !hasUseExeDir {
			launchArgs = append(launchArgs, "/useexedir")
		}
	}

	// Manage 32-bit engine DLLs (inject for dxvk, clean for gptk/d3dmetal)
	a.emitEvent("launch_status", map[string]interface{}{
		"gameId": gameID,
		"status": "Activating Translation Layer (D3DMetal / DXVK)...",
	})
	a.inject32BitEngineDLLs(winePrefix, gameDir, rt.Engine)

	// Failsafe: Terminate any lingering wineserver instances
	for _, eng := range []string{"wine", "gptk"} {
		ws := filepath.Join(a.EnginesDir(), eng, "bin", "wineserver")
		if _, err := os.Stat(ws); err == nil {
			cleanup := exec.Command(ws, "-k")
			cleanup.Env = append(engine.SafeEnviron(), "WINEPREFIX="+winePrefix)
			_ = cleanup.Run()
		}
	}

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

	for k, v := range rt.EnvVars {
		envMap[k] = v
	}

	// Safely detect if the 32-bit executable was compiled with /LARGEADDRESSAWARE (0x0020).
	// Forcing 4GB address space on legacy 32-bit binaries that lack LAA causes pointer truncation,
	// sign extension errors, stack corruption, and fatal SEH invalid frame crashes (status c0000005).
	if engine.IsLargeAddressAware(exePath) {
		a.logInfo("[Launch 32-bit] Executable has LARGE_ADDRESS_AWARE flag set. Enabling 4GB virtual address space.")
		envMap["WINE_LARGE_ADDRESS_AWARE"] = "1"
	} else {
		a.logInfo("[Launch 32-bit] Executable is NOT Large Address Aware. Constraining to safe 2GB virtual address space.")
		envMap["WINE_LARGE_ADDRESS_AWARE"] = "0"
	}

	var overrides []string
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
	} else if rt.Engine == "gptk" || rt.Engine == "d3dmetal" {
		overrides = append(overrides, "d3d9=b,n", "dxgi=b,n", "d3d11=b,n", "d3d12=b,n", "d3dx9_30=n,b")
	}

	overrides = append(overrides, "steam_api=n,b", "steam_api64=n,b", "steamclient=n,b", "steamclient64=n,b")
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
	regEnv := append(engine.SafeEnviron(), "WINEPREFIX="+winePrefix, "WINEDEBUG=-all")

	regRetina := exec.Command(wineBinary, "reg", "add", `HKEY_CURRENT_USER\Software\Wine\Mac Driver`, "/v", "RetinaMode", "/d", retinaVal, "/f")
	regRetina.Env = regEnv
	_ = regRetina.Run()

	// Default WineD3D OpenGL with GLSL is used on Apple Silicon for D3D9 titles.

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
		if rt.Engine == "gptk" || rt.Engine == "d3dmetal" {
			envSlice = append(envSlice, "WINEMSYNC=1", "WINEESYNC=0")
		} else {
			envSlice = append(envSlice, "WINEESYNC=0", "WINEMSYNC=0")
		}
	}

	cmd.Env = envSlice
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	a.emitEvent("launch_status", map[string]interface{}{
		"gameId": gameID,
		"status": "Launching Game Executable...",
	})

	go func() {
		a.logInfo("[Launch 32-bit] Executing 32-bit translated Steam game: %s", exePath)
		a.emitEvent("launch_complete", gameID)
		sysProfile := GetMacHardwareSpecs()
		sysProfile["Engine"] = rt.Engine
		sysProfile["Architecture"] = "x86_32"
		sysProfile["WineVersion"] = "RIFT-Wine-11.0-WoW64"

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

		sysmon.ClearActiveGameAndSave(filepath.Join(home, ".rift", "logs"))
		a.emitEvent("game_process_exited", gameID)
	}()

	return "Steam 32-bit Translated Launch Executed."
}

// inject32BitEngineDLLs copies 32-bit DXVK/graphics DLLs into both syswow64 and the game folder.
func (a *App) inject32BitEngineDLLs(prefixPath, gameDir, engineType string) {
	syswow64 := filepath.Join(prefixPath, "drive_c", "windows", "syswow64")
	os.MkdirAll(syswow64, 0755)

	copyFile := func(src, dst string) {
		s, err := os.Open(src)
		if err != nil {
			return
		}
		defer s.Close()
		d, err := os.Create(dst)
		if err != nil {
			return
		}
		defer d.Close()
		io.Copy(d, s)
	}

	if strings.ToLower(engineType) == "dxvk" {
		// 1. Direct3D 9, 8 & DirectDraw: dgVoodoo2 translates D3D9/D3D8/DDraw to Direct3D 11
		dgVoodooX32 := filepath.Join(a.EnginesDir(), "dgvoodoo", "x32")
		dgDlls := []string{"d3d9.dll", "ddraw.dll", "d3dimm.dll", "d3d8.dll"}
		for _, dll := range dgDlls {
			src := filepath.Join(dgVoodooX32, dll)
			if _, err := os.Stat(src); err == nil {
				copyFile(src, filepath.Join(syswow64, dll))
			}
		}

		// 2. Direct3D 11 & 10: DXVK-macOS translates Direct3D 11 into Vulkan (MoltenVK) -> Metal
		dxvkX32 := filepath.Join(a.EnginesDir(), "dxvk", "x32")
		dxvkDlls := []string{"d3d11.dll", "d3d10core.dll"}
		for _, dll := range dxvkDlls {
			src := filepath.Join(dxvkX32, dll)
			if _, err := os.Stat(src); err == nil {
				copyFile(src, filepath.Join(syswow64, dll))
			}
		}
	} else if strings.ToLower(engineType) == "gptk" || strings.ToLower(engineType) == "d3dmetal" {
		// Clean up any leftover dgVoodoo/DXVK DLLs from the game directory so GPTK doesn't attempt to load Vulkan
		cleanupDlls := []string{"d3d9.dll", "d3d11.dll", "d3d10core.dll", "d3d8.dll", "d3dimm.dll", "ddraw.dll", "dxgi.dll", "dxvk.conf"}
		for _, dll := range cleanupDlls {
			_ = os.Remove(filepath.Join(gameDir, dll))
		}
	}
}

// is32BitExecutable checks if a PE binary has the 32-bit i386 machine type.
func is32BitExecutable(exePath string) bool {
	if exePath == "" {
		return false
	}
	f, err := pe.Open(exePath)
	if err != nil {
		return false
	}
	defer f.Close()
	return f.Machine == pe.IMAGE_FILE_MACHINE_I386
}
