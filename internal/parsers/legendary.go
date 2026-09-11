package parsers

import (
	"encoding/json"
	"os"
)

// LegendaryInstalledGame represents the installation data of an Epic game in installed.json.
type LegendaryInstalledGame struct {
	AppName          string `json:"app_name"`
	Title            string `json:"title"`
	InstallPath      string `json:"install_path"`
	Version          string `json:"version"`
	Executable       string `json:"executable"`
	LaunchParameters string `json:"launch_parameters"`
}

// ParseLegendaryInstalled reads one or more installed.json files and merges them into a single map.
// The map key is the AppName (Epic's internal ID), and the value is the full game struct.
func ParseLegendaryInstalled(configPaths []string) map[string]LegendaryInstalledGame {
	installed := make(map[string]LegendaryInstalledGame)

	for _, path := range configPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue // File might not exist, which is fine
		}

		var parsedData map[string]LegendaryInstalledGame
		if err := json.Unmarshal(data, &parsedData); err == nil {
			for appName, game := range parsedData {
				if game.InstallPath != "" {
					// Only keep it if the install_path physically exists on the SSD
					if info, errStat := os.Stat(game.InstallPath); errStat == nil && info.IsDir() {
						installed[appName] = game
					}
				}
			}
		}
	}

	return installed
}
