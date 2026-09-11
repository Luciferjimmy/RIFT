package parsers

import (
	"fmt"
	"os"
	"regexp"
)

// ParseAppManifest reads a Steam appmanifest_<id>.acf file and returns
// the appID, game name, install directory, and whether it's fully installed.
func ParseAppManifest(acfPath string) (appID, name, installDir string, isInstalled bool, err error) {
	data, err := os.ReadFile(acfPath)
	if err != nil {
		return "", "", "", false, err
	}
	content := string(data)

	// Simple regex parsing for VDF format
	reAppID := regexp.MustCompile(`"appid"\s+"([^"]+)"`)
	reName := regexp.MustCompile(`"name"\s+"([^"]+)"`)
	reInstallDir := regexp.MustCompile(`"installdir"\s+"([^"]+)"`)
	reStateFlags := regexp.MustCompile(`"StateFlags"\s+"([^"]+)"`)

	if m := reAppID.FindStringSubmatch(content); len(m) > 1 {
		appID = m[1]
	}
	if m := reName.FindStringSubmatch(content); len(m) > 1 {
		name = m[1]
	}
	if m := reInstallDir.FindStringSubmatch(content); len(m) > 1 {
		installDir = m[1]
	}
	if m := reStateFlags.FindStringSubmatch(content); len(m) > 1 {
		// StateFlags 4 means fully installed
		if m[1] == "4" {
			isInstalled = true
		}
	}

	if appID == "" || name == "" {
		return "", "", "", false, fmt.Errorf("invalid manifest")
	}

	return appID, name, installDir, isInstalled, nil
}
