package rift

import (
	"os"
	"path/filepath"
	"strings"
)

// SetupEpicNative performs post-download configuration for Mac Native games downloaded via Epic/Legendary.
func (a *App) SetupEpicNative(gameID, appID, installDir string) error {
	a.logInfo("[Setup] Starting setup for Epic Native game: %s", gameID)

	var appDirToRename string
	var executablePath string

	filepath.Walk(installDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() && info.Name() == "Contents" {
			parentDir := filepath.Dir(path)
			// Rename un-extensioned Mac app folders to .app
			if parentDir != installDir && !strings.HasSuffix(parentDir, ".app") {
				appDirToRename = parentDir
			}
		}

		// Look for Mac executables to ensure they have execute permissions
		if info.IsDir() && info.Name() == "MacOS" {
			parentDir := filepath.Dir(path)
			if filepath.Base(parentDir) == "Contents" {
				files, _ := os.ReadDir(path)
				for _, f := range files {
					if !f.IsDir() {
						execFile := filepath.Join(path, f.Name())
						os.Chmod(execFile, 0755)
						executablePath = execFile
					}
				}
			}
		}
		return nil
	})

	if appDirToRename != "" {
		newAppDir := appDirToRename + ".app"
		if err := os.Rename(appDirToRename, newAppDir); err == nil {
			a.logInfo("[Setup] Renamed un-extensioned Mac app to: %s", filepath.Base(newAppDir))
			if executablePath != "" {
				executablePath = strings.Replace(executablePath, appDirToRename, newAppDir, 1)
			}
		}
	}

	if executablePath != "" {
		a.logInfo("[Setup] Ensured execution permissions for: %s", executablePath)
	}

	a.logInfo("[Setup] Epic Native setup complete for %s", gameID)
	return nil
}
