package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"rift/internal/engine"
	"strings"
	"time"

	"rift/internal/config"
)

// Backend describes a rendering backend.
type Backend struct {
	Name   string
	Env    map[string]string
	Args   []string
	Prefix string // Wine DLL overrides
}

// SelectConfig builds the backend configuration for a game.
func SelectConfig(gameCfg *config.GameConfig, globalCfg *config.GlobalConfig) *Backend {
	switch gameCfg.Backend {
	case "d3dmetal":
		return d3dmetalBackend(globalCfg)
	case "dxvk":
		return dxvkBackend(gameCfg, globalCfg)
	case "moltenvk":
		return moltenvkBackend(globalCfg, gameCfg)
	case "wined3d":
		return wined3dBackend()
	default:
		return dxvkBackend(gameCfg, globalCfg)
	}
}

func d3dmetalBackend(globalCfg *config.GlobalConfig) *Backend {
	env := map[string]string{}

	if globalCfg.Runtimes.GPTKLib != "" {
		gptkDir := filepath.Dir(globalCfg.Runtimes.GPTKLib)

		// Build library path: framework dir, MoltenVK dir, system paths
		libPath := gptkDir + ":/usr/lib:/usr/local/lib:/opt/homebrew/lib"

		// For framework bundles, include Resources dir so libdxcompiler etc. can be found
		resourcesDir := filepath.Join(gptkDir, "Resources")
		if info, err := os.Stat(resourcesDir); err == nil && info.IsDir() {
			libPath = resourcesDir + ":" + libPath
		}

		// Also add MoltenVK path if available (D3DMetal may use MoltenVK for some paths)
		if globalCfg.Runtimes.MoltenVKLib != "" {
			mvkDir := filepath.Dir(globalCfg.Runtimes.MoltenVKLib)
			libPath = mvkDir + ":" + libPath
		}

		env["DYLD_FALLBACK_LIBRARY_PATH"] = libPath
		env["DYLD_INSERT_LIBRARIES"] = globalCfg.Runtimes.GPTKLib
	}

	env["MTL_HUD_ENABLED"] = "0"
	env["WINEESYNC"] = "1"
	env["WINEDLLOVERRIDES"] = "d3d9,d3d10core,d3d11,d3d12,d3d12core,dxgi=n,b"

	return &Backend{
		Name: "D3DMetal",
		Env:  env,
		Args: []string{},
	}
}

func dxvkBackend(gameCfg *config.GameConfig, globalCfg *config.GlobalConfig) *Backend {
	env := map[string]string{
		"DXVK_ASYNC":       "1",
		"DXVK_LOG_LEVEL":   "warn",
		"WINEESYNC":        "1",
		"MTL_HUD_ENABLED":  "0",
		"WINEDLLOVERRIDES": "d3d9,d3d10core,d3d11,dxgi=n,b",
	}

	// DXVK-macOS needs MoltenVK for Vulkan→Metal translation
	if globalCfg.Runtimes.MoltenVKLib != "" {
		mvkDir := filepath.Dir(globalCfg.Runtimes.MoltenVKLib)
		existingPath := os.Getenv("DYLD_FALLBACK_LIBRARY_PATH")
		if existingPath != "" {
			env["DYLD_FALLBACK_LIBRARY_PATH"] = mvkDir + ":" + existingPath
		} else {
			env["DYLD_FALLBACK_LIBRARY_PATH"] = mvkDir
		}
	}

	return &Backend{
		Name: "DXVK-macOS",
		Env:  env,
	}
}

func moltenvkBackend(globalCfg *config.GlobalConfig, gameCfg *config.GameConfig) *Backend {
	env := map[string]string{
		"MTL_HUD_ENABLED":          "0",
		"WINEESYNC":                "1",
		"DISABLE_VK_LAYER_VALVE_1": "1",
	}

	if globalCfg.Runtimes.MoltenVKLib != "" {
		env["DYLD_INSERT_LIBRARIES"] = globalCfg.Runtimes.MoltenVKLib
		env["DYLD_FALLBACK_LIBRARY_PATH"] = filepath.Dir(globalCfg.Runtimes.MoltenVKLib)
	}

	return &Backend{
		Name: "MoltenVK",
		Env:  env,
	}
}

func wined3dBackend() *Backend {
	return &Backend{
		Name: "WineD3D",
		Env: map[string]string{
			"MTL_HUD_ENABLED": "0",
			"WINEESYNC":       "1",
		},
	}
}

// LaunchResult describes the outcome of a launch attempt.
type LaunchResult struct {
	Success  bool
	Backend  string
	Duration time.Duration
	Error    string
	PID      int
}

// Launch attempts to run a game with a specific backend.
func Launch(
	wineBin string,
	prefixDir string,
	exePath string,
	gameCfg *config.GameConfig,
	backend *Backend,
	workDir string,
) *LaunchResult {
	result := &LaunchResult{Backend: backend.Name}

	// Find wine bin dir for PATH
	wineDir := filepath.Dir(wineBin)
	// Try to find the Resources/wine/bin directory
	resourcesDir := filepath.Join(filepath.Dir(wineDir), "Resources", "wine", "bin")
	if info, err := os.Stat(resourcesDir); err == nil && info.IsDir() {
		wineDir = resourcesDir
	}

	// Build environment — START with the parent process env, then overlay ours
	// (duplicate keys: LAST wins, so our specific vars must come AFTER engine.SafeEnviron())
	env := engine.SafeEnviron()

	env = append(env,
		"WINEPREFIX="+prefixDir,
		"WINEARCH=win64",
		"PATH="+wineDir+":"+os.Getenv("PATH"),
	)

	// Add runtime library paths if needed
	if libPaths := getLibraryPaths(); libPaths != "" {
		env = append(env, "DYLD_FALLBACK_LIBRARY_PATH="+libPaths)
	}

	// Add game-specific env vars from config
	for k, v := range gameCfg.Env {
		env = append(env, k+"="+v)
	}

	// Add backend env vars (overrides game-specific if conflict)
	for k, v := range backend.Env {
		env = append(env, k+"="+v)
	}

	// Build command
	args := []string{exePath}
	args = append(args, gameCfg.Args...)
	args = append(args, backend.Args...)

	cmd := exec.Command(wineBin, args...)
	if workDir != "" {
		cmd.Dir = workDir
	} else {
		cmd.Dir = filepath.Dir(exePath)
	}
	cmd.Env = env

	// Start the game
	start := time.Now()
	if err := cmd.Start(); err != nil {
		result.Error = fmt.Sprintf("failed to start: %v", err)
		return result
	}

	result.PID = cmd.Process.Pid

	// Monitor for crash window: wait up to 30 seconds
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		duration := time.Since(start)
		result.Duration = duration
		if err != nil {
			result.Error = fmt.Sprintf("exited with error after %v: %v", duration.Round(time.Second), err)
		} else {
			// Exited cleanly — game might have finished normally
			result.Success = true
		}
	case <-time.After(30 * time.Second):
		// Game has been running for 30+ seconds — assume success
		result.Success = true
		result.Duration = 30 * time.Second
		// Detach from the process — let it keep running
		cmd.Process.Release()
	}

	return result
}

func getLibraryPaths() string {
	paths := []string{
		"/usr/lib",
		"/usr/local/lib",
		"/opt/homebrew/lib",
	}

	// Check for MoltenVK in common locations
	moltenVKPaths := []string{
		"/opt/homebrew/lib/libMoltenVK.dylib",
		"/usr/local/lib/libMoltenVK.dylib",
	}
	for _, p := range moltenVKPaths {
		if _, err := os.Stat(p); err == nil {
			paths = append(paths, filepath.Dir(p))
		}
	}

	return strings.Join(paths, ":")
}

// LogOutput saves the launch output to a log file.

// LogOutput saves the launch output to a log file.
func LogOutput(gameID, logLine string) error {
	logPath := config.GameLogPath(gameID)
	logDir := filepath.Dir(logPath)
	os.MkdirAll(logDir, 0755)

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(logLine + "\n")
	return err
}
