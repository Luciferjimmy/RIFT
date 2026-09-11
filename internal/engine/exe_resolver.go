package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"rift/internal/parsers"
	"sort"
	"strings"
)

type ExeCandidate struct {
	Path   string `json:"path"`   // Relative path within game folder
	Score  int    `json:"score"`  // Weighted score
	Source string `json:"source"` // "runtime_config" / "epic_manifest" / "steam_appinfo" / "scan"
}

// GenericNonGameBinaries lists known helper, installer, extractor, and crash report binaries.
var GenericNonGameBinaries = []string{
	"unins", "uninstall", "setup", "installer", "install", "crashreport", "bugreport",
	"redist", "vcredist", "vc_redist", "dxsetup", "directx", "dxredist", "unitycrash", "cef",
	"update", "updater", "patch", "socialclub", "eosbootstrapper",
	"epicbootstrapper", "bootstrapper", "launcher_helper", "7z", "7za",
	"winrar", "unrar", "touchup", "easyanticheat", "battleye", "eac_",
	"cleanup", "register", "quickinstaller", "helper",
	"dotnet", "netfx", "framework", "physx", "oalinst", "openal", "msvcrt",
}

// IsNonGameBinary checks if an executable path or filename belongs to a known utility or installer.
func IsNonGameBinary(pathOrName string) bool {
	lowerPath := strings.ToLower(filepath.ToSlash(pathOrName))
	if strings.Contains(lowerPath, "/downloading/") || strings.Contains(lowerPath, "/temp/") ||
		strings.Contains(lowerPath, "/_commonredist/") || strings.Contains(lowerPath, "/redist/") ||
		strings.Contains(lowerPath, "/installers/") || strings.Contains(lowerPath, "/prerequisites/") ||
		strings.Contains(lowerPath, "downloading/") {
		return true
	}
	lower := strings.ToLower(filepath.Base(pathOrName))
	lower = strings.TrimSuffix(lower, ".exe")
	for _, ignored := range GenericNonGameBinaries {
		if strings.Contains(lower, ignored) {
			return true
		}
	}
	return false
}

// ResolveGameExecutable attempts to find the main .exe file for a game using a clean 4-tier pipeline:
// Tier 1: Check game capsule runtime config (game_runtime.json)
// Tier 2: Read official launcher manifests (Epic installed.json, Steam app_info / appmanifest)
// Tier 3: Weighted directory scan using file depth, binary size, and generic installer filter
func ResolveGameExecutable(gameDir, appID, gameName, steamcmdDir string) (string, []ExeCandidate, error) {
	var candidates []ExeCandidate

	// Strategy 0: Game Capsule Runtime Config (game_runtime.json)
	runtimePath := filepath.Join(gameDir, "game_runtime.json")
	if data, err := os.ReadFile(runtimePath); err == nil {
		var rconf map[string]interface{}
		if err := json.Unmarshal(data, &rconf); err == nil {
			relExe, ok := rconf["executable"].(string)
			if !ok || relExe == "" {
				relExe, _ = rconf["executable_path"].(string)
			}
			if relExe != "" && !IsNonGameBinary(relExe) {
				fullPath := filepath.Join(gameDir, relExe)
				if _, err := os.Stat(fullPath); err == nil {
					candidates = append(candidates, ExeCandidate{
						Path:   relExe,
						Score:  105,
						Source: "runtime_config",
					})
				}
			}
		}
	}

	// Strategy 1: Official Epic / Legendary Manifest Parsing (installed.json)
	homeDir, _ := os.UserHomeDir()
	manifestPaths := []string{
		filepath.Join(homeDir, ".config", "legendary", "installed.json"),
		filepath.Join(homeDir, ".rift", "auth", "epic", "installed.json"),
		filepath.Join(homeDir, ".config", "nile", "installed.json"),
		filepath.Join(gameDir, "..", "installed.json"),
		filepath.Join(gameDir, "installed.json"),
	}

	epicGames := parsers.ParseLegendaryInstalled(manifestPaths)
	for appKey, g := range epicGames {
		match := (appKey == appID || g.AppName == appID || strings.EqualFold(g.Title, gameName))
		if !match && g.InstallPath != "" && gameDir != "" {
			rel, err := filepath.Rel(filepath.Clean(g.InstallPath), filepath.Clean(gameDir))
			if err == nil && (rel == "." || rel == "") {
				match = true
			}
		}

		if match && g.Executable != "" && !IsNonGameBinary(g.Executable) {
			exePath := strings.ReplaceAll(g.Executable, "\\", "/")
			fullPath := filepath.Join(gameDir, exePath)
			if _, err := os.Stat(fullPath); err == nil {
				candidates = append(candidates, ExeCandidate{
					Path:   exePath,
					Score:  100,
					Source: "epic_manifest",
				})
			} else {
				baseExe := filepath.Base(exePath)
				if _, err := os.Stat(filepath.Join(gameDir, baseExe)); err == nil {
					candidates = append(candidates, ExeCandidate{
						Path:   baseExe,
						Score:  100,
						Source: "epic_manifest",
					})
				}
			}
		}
	}

	// Strategy 2: Steam App Info via SteamCMD (if steamcmd is available)
	if IsSteamCMDReady(steamcmdDir) {
		cmdScript := filepath.Join(steamcmdDir, "steamcmd.sh")
		cmd := exec.Command(cmdScript, "+app_info_print", appID, "+quit")
		cmd.Dir = steamcmdDir
		var outBuf bytes.Buffer
		cmd.Stdout = &outBuf

		if err := cmd.Run(); err == nil {
			output := outBuf.String()
			re := regexp.MustCompile(`(?i)"executable"\s+"([^"]+\.exe)"`)
			matches := re.FindAllStringSubmatch(output, -1)
			for _, m := range matches {
				if len(m) == 2 {
					exePath := strings.TrimSpace(m[1])
					exePath = strings.ReplaceAll(exePath, "\\", "/")
					if !IsNonGameBinary(exePath) {
						fullPath := filepath.Join(gameDir, exePath)
						if _, err := os.Stat(fullPath); err == nil {
							candidates = append(candidates, ExeCandidate{
								Path:   exePath,
								Score:  95,
								Source: "steam_appinfo",
							})
						}
					}
				}
			}
		}
	}

	// Strategy 3: Directory Scanning with Heuristic Weighting
	scanCandidates, err := ScanDirectoryForExes(gameDir, gameName)
	if err == nil {
		candidates = append(candidates, scanCandidates...)
	}

	if len(candidates) == 0 {
		return "", nil, fmt.Errorf("no executable files found in %s", gameDir)
	}

	// Sort candidates by score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	// Deduplicate candidates
	seen := make(map[string]bool)
	var uniqueCandidates []ExeCandidate
	for _, c := range candidates {
		relPath := filepath.Clean(c.Path)
		if !seen[relPath] {
			seen[relPath] = true
			c.Path = relPath
			uniqueCandidates = append(uniqueCandidates, c)
		}
	}

	if len(uniqueCandidates) == 0 || uniqueCandidates[0].Score <= 0 || IsNonGameBinary(uniqueCandidates[0].Path) {
		return "", uniqueCandidates, fmt.Errorf("no playable game executable found in %s (only utility or installer binaries detected)", gameDir)
	}

	bestMatch := uniqueCandidates[0].Path
	return bestMatch, uniqueCandidates, nil
}

// ScanDirectoryForExes recursively finds and scores all .exe files in a game directory.
func ScanDirectoryForExes(gameDir, gameName string) ([]ExeCandidate, error) {
	var candidates []ExeCandidate

	lowerGameName := strings.ToLower(gameName)
	cleanGameName := regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(lowerGameName, "")

	err := filepath.WalkDir(gameDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := strings.ToLower(d.Name())

		relPath, err := filepath.Rel(gameDir, path)
		if err != nil {
			relPath = d.Name()
		}

		if d.IsDir() {
			lowerDir := strings.ToLower(d.Name())
			if lowerDir == "downloading" || lowerDir == "temp" || lowerDir == "_commonredist" ||
				lowerDir == "redist" || lowerDir == "installers" || lowerDir == "support" ||
				lowerDir == "prerequisites" || (strings.HasPrefix(lowerDir, ".") && !strings.HasSuffix(lowerDir, ".app")) {
				return filepath.SkipDir
			}
			if strings.HasSuffix(name, ".app") {
				candidates = append(candidates, ExeCandidate{
					Path:   relPath,
					Score:  150, // Mac native apps get highest priority
					Source: "scan",
				})
			}
			return nil
		}

		if !strings.HasSuffix(name, ".exe") {
			return nil
		}

		// Calculate weight score
		score := 50

		// Penalize standard generic non-game installer binaries
		if IsNonGameBinary(relPath) || IsNonGameBinary(name) {
			score -= 150
		}

		// Reward root directory or engine binary directory (Binaries/Win64, Binaries/Win32)
		toSlashRel := filepath.ToSlash(relPath)
		depth := strings.Count(toSlashRel, "/")
		if depth == 0 {
			score += 20
		} else if strings.Contains(toSlashRel, "Binaries/Win64") || strings.Contains(toSlashRel, "Binaries/Win32") {
			score += 25
		} else if depth == 1 {
			score += 10
		}

		// Reward name matches with game title
		cleanExeName := regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(strings.TrimSuffix(name, ".exe"), "")
		if cleanGameName != "" && cleanExeName != "" {
			if cleanExeName == cleanGameName {
				score += 30
			} else if strings.Contains(cleanExeName, cleanGameName) || strings.Contains(cleanGameName, cleanExeName) {
				score += 15
			}
		}

		// Reward larger file size (main game executables are significantly larger than helper utilities)
		if info, err := d.Info(); err == nil {
			sizeMB := info.Size() / (1024 * 1024)
			if sizeMB > 50 {
				score += 25
			} else if sizeMB > 10 {
				score += 10
			}
		}

		candidates = append(candidates, ExeCandidate{
			Path:   relPath,
			Score:  score,
			Source: "scan",
		})

		return nil
	})

	return candidates, err
}
