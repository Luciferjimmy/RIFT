package parsers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseLegendaryInstalled(t *testing.T) {
	tempDir := t.TempDir()

	// Create fake game directories so os.Stat passes
	game1Dir := filepath.Join(tempDir, "Games", "Game1")
	game2Dir := filepath.Join(tempDir, "Games", "Game2")
	os.MkdirAll(game1Dir, 0755)
	os.MkdirAll(game2Dir, 0755)

	// Game3Dir doesn't exist, should be skipped

	jsonPath1 := filepath.Join(tempDir, "installed_1.json")
	jsonPath2 := filepath.Join(tempDir, "installed_2.json")

	jsonData1 := `{
		"Game1": {
			"app_name": "Game1",
			"title": "Cool Game",
			"install_path": "` + game1Dir + `",
			"version": "1.0",
			"executable": "game.exe"
		},
		"Game3": {
			"app_name": "Game3",
			"title": "Missing Game",
			"install_path": "/path/does/not/exist",
			"version": "1.0"
		}
	}`

	jsonData2 := `{
		"Game2": {
			"app_name": "Game2",
			"title": "Another Game",
			"install_path": "` + game2Dir + `",
			"version": "2.0"
		}
	}`

	os.WriteFile(jsonPath1, []byte(jsonData1), 0644)
	os.WriteFile(jsonPath2, []byte(jsonData2), 0644)

	installed := ParseLegendaryInstalled([]string{jsonPath1, jsonPath2, filepath.Join(tempDir, "missing.json")})

	if len(installed) != 2 {
		t.Fatalf("Expected 2 installed games, got %d", len(installed))
	}

	g1, ok := installed["Game1"]
	if !ok || g1.Title != "Cool Game" {
		t.Errorf("Game1 missing or incorrect")
	}

	g2, ok := installed["Game2"]
	if !ok || g2.Title != "Another Game" {
		t.Errorf("Game2 missing or incorrect")
	}

	if _, ok := installed["Game3"]; ok {
		t.Errorf("Game3 should be skipped because path does not exist")
	}
}
