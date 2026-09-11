package parsers

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ParseOwnedSteamGames reads the localconfig.vdf file from the Global Auth directory,
// isolates the "apps" block using brace balancing, and extracts all unique Steam AppIDs the user owns.
func ParseOwnedSteamGames(globalAuthDir string, steam32ID string) ([]string, error) {
	vdfPath := filepath.Join(globalAuthDir, "userdata", steam32ID, "config", "localconfig.vdf")

	content, err := os.ReadFile(vdfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read localconfig.vdf: %w", err)
	}

	text := string(content)
	appsIdx := strings.Index(strings.ToLower(text), "\"apps\"")
	if appsIdx == -1 {
		return nil, fmt.Errorf("apps block not found in VDF")
	}

	// Find the opening brace of the apps block
	braceStart := strings.Index(text[appsIdx:], "{")
	if braceStart == -1 {
		return nil, fmt.Errorf("opening brace of apps block not found")
	}
	braceStart = appsIdx + braceStart

	// Balance braces to extract the content inside the "apps" node
	depth := 1
	endIdx := -1
	for i := braceStart + 1; i < len(text); i++ {
		if text[i] == '{' {
			depth++
		} else if text[i] == '}' {
			depth--
			if depth == 0 {
				endIdx = i
				break
			}
		}
	}
	if endIdx == -1 {
		return nil, fmt.Errorf("closing brace of apps block not found")
	}

	appsBlock := text[braceStart+1 : endIdx]

	// Match keys under the apps block that look like `"AppID" {`
	reKey := regexp.MustCompile(`"([0-9]+)"\s*\{`)
	matches := reKey.FindAllStringSubmatch(appsBlock, -1)

	appMap := make(map[string]bool)
	var ownedGames []string

	for _, match := range matches {
		if len(match) > 1 {
			appID := match[1]
			// Deduplicate AppIDs
			if appID != "0" && !appMap[appID] {
				appMap[appID] = true
				ownedGames = append(ownedGames, appID)
			}
		}
	}

	return ownedGames, nil
}

// CheckIfGameInstalled scans the steamapps folder for the appmanifest file.
// If the ACF file exists, the game is installed locally.
func CheckIfGameInstalled(steamappsDir string, appID string) bool {
	acfPath := filepath.Join(steamappsDir, fmt.Sprintf("appmanifest_%s.acf", appID))
	_, err := os.Stat(acfPath)
	return err == nil
}

// ParseSteamLoginUser reads loginusers.vdf and returns the persona name of the most recent logged-in user.
func ParseSteamLoginUser(globalAuthDir string) (string, error) {
	// First try loginusers.vdf (legacy SteamCMD session storage)
	vdfPath := filepath.Join(globalAuthDir, "config", "loginusers.vdf")
	content, err := os.ReadFile(vdfPath)
	if err == nil {
		text := string(content)
		rePersona := regexp.MustCompile(`"PersonaName"\s+"([^"]+)"`)
		reMostRecent := regexp.MustCompile(`"MostRecent"\s+"1"`)

		// Split by '}' followed by '"' to isolate each user block, since each user starts with an ID in quotes
		blocks := regexp.MustCompile(`\}\s*"`).Split(text, -1)
		for _, block := range blocks {
			if reMostRecent.MatchString(block) {
				matches := rePersona.FindStringSubmatch(block)
				if len(matches) > 1 {
					return matches[1], nil
				}
			}
		}

		// Fallback to first persona name found if no "MostRecent" "1" is explicitly matched
		matches := rePersona.FindStringSubmatch(text)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	// Fall back to config.vdf (modern SteamCMD stores "Accounts" section with username)
	configPath := filepath.Join(globalAuthDir, "config", "config.vdf")
	cfgContent, err := os.ReadFile(configPath)
	if err == nil {
		// Match: "Accounts" { "username" { ... } }
		re := regexp.MustCompile(`"Accounts"\s*\{[^}]*"([^"\s]+)"\s*\{`)
		matches := re.FindStringSubmatch(string(cfgContent))
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("no user logged in")
}

// ParseLibraryFolders reads libraryfolders.vdf and returns all paths where Steam libraries exist.
func ParseLibraryFolders(vdfPath string) []string {
	content, err := os.ReadFile(vdfPath)
	if err != nil {
		return nil
	}

	text := string(content)
	rePath := regexp.MustCompile(`(?i)"path"\s+"([^"]+)"`)
	matches := rePath.FindAllStringSubmatch(text, -1)

	var paths []string
	for _, match := range matches {
		if len(match) > 1 {
			// Windows paths in VDF have double backslashes, e.g. "D:\\SteamLibrary"
			// Unescape them to standard Go paths if needed, or filepath.Clean
			// Because Mac natively mounts disks at /Volumes, it should just be clean
			cleanPath := match[1]
			// Some VDFs on Mac might have escapes like \\ but usually they don't
			paths = append(paths, cleanPath)
		}
	}
	return paths
}
