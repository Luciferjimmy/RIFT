package rift

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
	"rift/internal/sysmon"

	_ "github.com/mattn/go-sqlite3"
)

// GetLastSessionTelemetry returns the postmortem data from the last game session
func (a *App) GetLastSessionTelemetry() string {
	home, _ := os.UserHomeDir()
	logsDir := filepath.Join(home, ".rift", "logs")
	return sysmon.GetLastSessionTelemetry(logsDir)
}

// GetPlaytimeStats returns the playtime data scoped to the active user.
func (a *App) GetPlaytimeStats() string {
	playtimeFile := filepath.Join(a.UserDir(), "playtime.json")
	if data, err := os.ReadFile(playtimeFile); err == nil {
		return string(data)
	}
	return `{"total": 0, "per_game": {}}`
}

// GetLastSessionSummary returns the full, formatted telemetry payload of the most recent session
func (a *App) GetLastSessionSummary() string {
	home, _ := os.UserHomeDir()
	summaryPath := filepath.Join(home, ".rift", "logs", "last_session_summary.json")
	if data, err := os.ReadFile(summaryPath); err == nil {
		return string(data)
	}
	return "{}"
}

// LastSessionCoreInfo holds dynamic details about the engine/core used on the last played title
type LastSessionCoreInfo struct {
	GameID       string `json:"game_id"`
	GameName     string `json:"game_name"`
	Engine       string `json:"engine"`
	CoreTitle    string `json:"core_title"`
	CoreBadge    string `json:"core_badge"`
	Subtitle     string `json:"subtitle"`
	LayerDetails string `json:"layer_details"`
	Platform     string `json:"platform"`
	IsMacNative  bool   `json:"is_mac_native"`
}

// GetLastSessionCoreInfo detects the most recently played game and returns its translation core configuration
func (a *App) GetLastSessionCoreInfo() string {
	home, _ := os.UserHomeDir()
	logsDir := filepath.Join(home, ".rift", "logs")

	var newestFile string
	var newestTime time.Time

	files, err := os.ReadDir(logsDir)
	if err == nil {
		for _, f := range files {
			if !f.IsDir() && filepath.Ext(f.Name()) == ".json" && strings.HasSuffix(f.Name(), "_postmortem.json") {
				info, err := f.Info()
				if err == nil && info.ModTime().After(newestTime) {
					newestTime = info.ModTime()
					newestFile = f.Name()
				}
			}
		}
	}

	gameID := ""
	if newestFile != "" {
		gameID = strings.TrimSuffix(newestFile, "_postmortem.json")
	}

	// Fallback to top played title if no postmortem exists
	if gameID == "" {
		playtimeFile := filepath.Join(a.UserDir(), "playtime.json")
		if b, err := os.ReadFile(playtimeFile); err == nil {
			var pt sysmon.PlaytimeData
			if json.Unmarshal(b, &pt) == nil && pt.PerGame != nil {
				var maxSec int64 = -1
				for gid, sec := range pt.PerGame {
					if sec > maxSec {
						maxSec = sec
						gameID = gid
					}
				}
			}
		}
	}

	gameName := ""
	engineName := "d3dmetal"
	isMac := false
	platform := "steam"

	// Known title dictionary for clean names
	knownNames := map[string]string{
		"steam-11020":                      "TrackMania Nations Forever",
		"0055e45ce7654c55aade646467349e83": "Mortal Shell",
		"c98c934106994f2d8afc36c3f1872549": "Undying",
		"23cbdd44de914e9caec216acefc51f9b": "Deadtime Defenders",
		"308461dd5d8248058a65a2667f92e330": "Breathedge",
		"0a20ccd3f1b3464da750a4dbf8c80d7c": "Emily is Away",
		"steam-417860":                     "Emily is Away",
		"steam-2762230":                    "Citizen Sleeper",
		"steam-223710":                     "Cry of Fear",
		"steam-698780":                     "Doki Doki Literature Club!",
		"steam-1671340":                    "KILL KNIGHT",
		"Batfish":                          "Telltale Batman Season 1",
	}

	if gameID != "" {
		if name, ok := knownNames[gameID]; ok {
			gameName = name
		}

		// Read game_runtime.json if it exists
		runtimePath := filepath.Join(home, ".rift", "games", gameID, "game_runtime.json")
		if data, err := os.ReadFile(runtimePath); err == nil {
			var rconf map[string]interface{}
			if json.Unmarshal(data, &rconf) == nil {
				if gn, ok := rconf["gameName"].(string); ok && gn != "" {
					gameName = gn
				}
				if eng, ok := rconf["engine"].(string); ok && eng != "" {
					engineName = eng
				}
				// Check if 32-bit executable
				exeName := ""
				if en, ok := rconf["executable"].(string); ok && en != "" {
					exeName = en
				} else if ep, ok := rconf["executable_path"].(string); ok && ep != "" {
					exeName = ep
				}
				if exeName != "" {
					gDir := filepath.Join(home, ".rift", "games", gameID, "game_files")
					if gd, ok := rconf["gameDir"].(string); ok && gd != "" {
						gDir = gd
					}
					targetExe := exeName
					if !filepath.IsAbs(targetExe) {
						targetExe = filepath.Join(gDir, targetExe)
					}
					if is32BitExecutable(targetExe) {
						engineName = "wow64_32"
					}
				}
			}
		}
	}

	if gameName == "" {
		if gameID != "" {
			gameName = gameID
		} else {
			gameName = "TrackMania Nations Forever"
		}
	}

	// Fallback check for TrackMania Nations Forever which is always 32-bit i386
	if gameID == "steam-11020" || strings.Contains(strings.ToLower(gameName), "trackmania") {
		engineName = "wow64_32"
	}

	res := LastSessionCoreInfo{
		GameID:      gameID,
		GameName:    gameName,
		Engine:      engineName,
		Platform:    platform,
		IsMacNative: isMac,
	}

	// Determine presentation strings based on engine and architecture
	lowerEng := strings.ToLower(engineName)
	if isMac || lowerEng == "native" || lowerEng == "macos" {
		res.CoreTitle = "macOS Native"
		res.CoreBadge = "NATIVE"
		res.Subtitle = gameName
		res.LayerDetails = "Apple Silicon ARM64 Direct"
	} else if lowerEng == "wow64_32" || lowerEng == "wow32" || lowerEng == "wow64" {
		// 32-bit Windows games run via Wine WoW64 (Apple Silicon lacks 32-bit D3DMetal)
		res.CoreTitle = "Wine WoW64 (32-bit)"
		res.CoreBadge = "WOW64 (32-BIT)"
		res.Subtitle = gameName
		res.LayerDetails = "32-on-64 WoW64 · DirectX 9 Layer"
	} else if lowerEng == "dxvk" {
		res.CoreTitle = "DXVK + MoltenVK"
		res.CoreBadge = "DXVK"
		res.Subtitle = gameName
		res.LayerDetails = "DirectX 9/10/11 → Metal"
	} else if lowerEng == "wined3d" || lowerEng == "wine" {
		res.CoreTitle = "WineD3D"
		res.CoreBadge = "WINED3D"
		res.Subtitle = gameName
		res.LayerDetails = "DirectX → OpenGL/Metal"
	} else {
		// gptk / d3dmetal (64-bit Apple Game Porting Toolkit translation layer)
		res.CoreTitle = "D3DMetal (GPTK)"
		res.CoreBadge = "D3DMETAL"
		res.Subtitle = gameName
		res.LayerDetails = "DirectX 11/12 → Metal 4"
	}

	b, _ := json.Marshal(res)
	return string(b)
}
