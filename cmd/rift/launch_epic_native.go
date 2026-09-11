package rift

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"rift/internal/engine"
	"rift/internal/sysmon"
)

// LaunchEpicNative natively launches macOS versions of Epic Games titles.
func (a *App) LaunchEpicNative(gameID, appID, gameDir string) string {
	a.logInfo("[Launch] ===== MAC NATIVE EPIC LAUNCH: gameID=%s appID=%s =====", gameID, appID)
	launchStartTime := time.Now()

	legendaryBin := filepath.Join(a.EnginesDir(), "legendary", "venv", "bin", "legendary")
	if _, err := os.Stat(legendaryBin); err != nil {
		if path, err := exec.LookPath("legendary"); err == nil {
			legendaryBin = path
		}
	}

	targetAppName := appID
	if targetAppName == "" {
		targetAppName = strings.TrimPrefix(gameID, "epic-")
	}

	epicGlobalAuth := filepath.Join(a.UserAuthDir(), "epic")
	legendaryConfig := filepath.Join(a.ResolveGameDir(gameID), ".legendary")
	
	// Attempt to fetch tokens
	legCmd := exec.Command(legendaryBin, "launch", targetAppName, "--no-wine", "--json")
	legCmd.Env = append(engine.SafeEnviron(), "LEGENDARY_CONFIG_PATH="+legendaryConfig)
	outBytes, err := legCmd.Output()
	if err != nil {
		// Fallback to global auth
		legCmdFallback := exec.Command(legendaryBin, "launch", targetAppName, "--no-wine", "--json")
		legCmdFallback.Env = append(engine.SafeEnviron(), "LEGENDARY_CONFIG_PATH="+epicGlobalAuth)
		outBytes, err = legCmdFallback.Output()
	}

	if err != nil {
		a.logWarn("[Launch] Could not fetch Epic DRM tokens for native Mac game %s: %v", gameID, err)
		return fmt.Sprintf("Error: Could not authenticate Epic Native Mac game. %v", err)
	}

	// Parse the output
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

	var exePath string
	var launchArgs []string

	if err := json.Unmarshal([]byte(jsonLine), &legData); err == nil {
		if legData.GameExe != "" && legData.GameDirectory != "" {
			exePath = filepath.Join(legData.GameDirectory, legData.GameExe)
		}
		launchArgs = append(launchArgs, legData.GameParams...)
		launchArgs = append(launchArgs, legData.EGLParams...)
	}

	if exePath == "" {
		return "Error: Legendary manifest did not return an executable path."
	}

	go func() {
		a.logInfo("[Launch] Starting native Epic Mac app directly: %s", exePath)
		cmd := exec.Command(exePath, launchArgs...)
		home, _ := os.UserHomeDir()
		cmd.Dir = legData.GameDirectory
		
		// Inherit environment variables but inject EOS parameters
		cmd.Env = engine.SafeEnviron()

		sysProfile := GetMacHardwareSpecs()
		sysProfile["Engine"] = "Native macOS"

		launchDurationMs := int(time.Since(launchStartTime).Milliseconds())
		a.MonitorGame(gameID, appID, targetAppName, exePath, cmd, sysProfile, nil, launchDurationMs)
		
		sysmon.ClearActiveGameAndSave(filepath.Join(home, ".rift", "logs"))
		a.emitEvent("game_process_exited", gameID)
	}()
	a.emitEvent("launch_complete", gameID)
	return "Native Mac Launch Executed."
}
