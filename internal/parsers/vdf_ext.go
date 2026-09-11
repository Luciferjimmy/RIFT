package parsers

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ParseInstalledAppIDs reads libraryfolders.vdf and returns all AppIDs found in the "apps" blocks.
func ParseInstalledAppIDs(vdfPath string) []string {
	content, err := os.ReadFile(vdfPath)
	if err != nil {
		return nil
	}

	text := string(content)
	var appIDs []string
	
	appsIdx := strings.Index(strings.ToLower(text), "\"apps\"")
	if appsIdx == -1 {
		return nil
	}
	
	// Find all "apps" blocks
	appsBlocks := regexp.MustCompile(`(?i)"apps"\s*\{([^}]*)\}`).FindAllStringSubmatch(text, -1)
	for _, block := range appsBlocks {
		if len(block) > 1 {
			inner := block[1]
			// Extract the keys (AppIDs)
			matches := regexp.MustCompile(`"([0-9]+)"\s+"[^"]+"`).FindAllStringSubmatch(inner, -1)
			for _, match := range matches {
				if len(match) > 1 {
					appIDs = append(appIDs, match[1])
				}
			}
		}
	}

	return appIDs
}

// ParseNonSteamGames reads localconfig.vdf and extracts Non-Steam games from the friends block.
// Returns a map of AppID -> Name
func ParseNonSteamGames(globalAuthDir string, steam32ID string) map[string]string {
	vdfPath := filepath.Join(globalAuthDir, "userdata", steam32ID, "config", "localconfig.vdf")
	content, err := os.ReadFile(vdfPath)
	if err != nil {
		return nil
	}

	text := string(content)
	nonSteamGames := make(map[string]string)

	// Non-steam games are stored in the friends block with 17-digit IDs.
	// e.g. "103582791475223930" { "name" "EA SPORTS FC 26" ... }
	re := regexp.MustCompile(`"([0-9]{16,20})"\s*\{\s*"name"\s+"([^"]+)"`)
	matches := re.FindAllStringSubmatch(text, -1)
	
	for _, match := range matches {
		if len(match) > 2 {
			appID := match[1]
			name := match[2]
			nonSteamGames[appID] = name
		}
	}
	
	return nonSteamGames
}
