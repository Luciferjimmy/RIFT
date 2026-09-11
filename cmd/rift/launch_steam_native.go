package rift

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"rift/internal/sysmon"
)

// LaunchSteamNative launches natively-compiled macOS Steam games by locating and executing their .app bundle.
func (a *App) LaunchSteamNative(gameID, appID, gameDir string) string {
	a.logInfo("[Launch] ===== MAC NATIVE STEAM LAUNCH: gameID=%s appID=%s =====", gameID, appID)
	launchStartTime := time.Now()

	appFolder := ""
	filepath.Walk(gameDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() && strings.HasSuffix(info.Name(), ".app") {
			appFolder = path
			return filepath.SkipDir
		}
		return nil
	})

	if appFolder != "" {
		go func() {
			a.logInfo("[Launch] Starting native macOS Steam app (blocking): %s", appFolder)
			cmd := exec.Command("open", "-W", appFolder)
			home, _ := os.UserHomeDir()
			
			sysProfile := GetMacHardwareSpecs()
			sysProfile["Engine"] = "Native macOS"

			realGameName := queryGameNameFromDB(appID)
			if realGameName == "" {
				realGameName = queryGameNameFromDB(gameID)
			}
			if realGameName == "" {
				realGameName = gameID
			}

			launchDurationMs := int(time.Since(launchStartTime).Milliseconds())
			a.MonitorGame(gameID, appID, realGameName, appFolder, cmd, sysProfile, nil, launchDurationMs)
			
			sysmon.ClearActiveGameAndSave(filepath.Join(home, ".rift", "logs"))
			a.emitEvent("game_process_exited", gameID)
		}()
		a.emitEvent("launch_complete", gameID)
		return "Native Mac Launch Executed."
	}

	return fmt.Sprintf("Error: Could not resolve a native Mac executable or .app bundle in %s", gameDir)
}
