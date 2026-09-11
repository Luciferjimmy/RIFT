package adapters

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setupMockHome(t *testing.T) string {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})
	return tempDir
}

func TestGetActualBackend(t *testing.T) {
	tests := []struct {
		name        string
		gameID      string
		isInstalled bool
		isMacNative bool
		dbBackend   string
		setupMock   func(homeDir string)
		expected    string
	}{
		{
			name:        "Mac Native Game",
			gameID:      "game1",
			isInstalled: true,
			isMacNative: true,
			dbBackend:   "",
			setupMock:   func(h string) {},
			expected:    "macOS Native",
		},
		{
			name:        "Not Installed, No Cache, No DB",
			gameID:      "game2",
			isInstalled: false,
			isMacNative: false,
			dbBackend:   "",
			setupMock:   func(h string) {},
			expected:    "Awaiting Translation Profile",
		},
		{
			name:        "Not Installed, DB Says DXVK",
			gameID:      "game3",
			isInstalled: false,
			isMacNative: false,
			dbBackend:   "requires dxvk",
			setupMock:   func(h string) {},
			expected:    "Wine + DXVK (DX9/10/11) - Pending",
		},
		{
			name:        "Installed, Fallback to DB (D3DMetal)",
			gameID:      "game4",
			isInstalled: true,
			isMacNative: false,
			dbBackend:   "dx12 game",
			setupMock:   func(h string) {},
			expected:    "Wine + D3DMetal (DX11/12)",
		},
		{
			name:        "Installed, Local Config Says DXVK",
			gameID:      "game5",
			isInstalled: true,
			isMacNative: false,
			dbBackend:   "dx12", // Should be overridden by local config
			setupMock: func(h string) {
				cfgDir := filepath.Join(h, ".rift", "games", "game5")
				os.MkdirAll(cfgDir, 0755)
				cfgData := `{"engine": "dxvk"}`
				os.WriteFile(filepath.Join(cfgDir, "rift_config.json"), []byte(cfgData), 0644)
			},
			expected: "Wine + DXVK (DX9/10/11)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := setupMockHome(t)
			tt.setupMock(home)

			result := getActualBackend(tt.gameID, tt.isInstalled, tt.isMacNative, tt.dbBackend)
			if result != tt.expected {
				t.Errorf("getActualBackend() = %q; want %q", result, tt.expected)
			}
		})
	}
}

func TestAdaptSteamGame(t *testing.T) {
	setupMockHome(t) // Ensure clean playtime
	profile := AdaptSteamGame("730", "CSGO", true, false, "dx9")

	if profile.ID != "steam-730" {
		t.Errorf("Expected ID steam-730, got %s", profile.ID)
	}
	if profile.Platform != "Steam" {
		t.Errorf("Expected Platform Steam, got %s", profile.Platform)
	}
	if profile.Backend != "Wine + DXVK (DX9/10/11)" {
		t.Errorf("Expected DXVK backend, got %s", profile.Backend)
	}
}

func TestLoadEpicMetadata(t *testing.T) {
	tempDir := t.TempDir()
	metaDir := filepath.Join(tempDir, "metadata")
	os.MkdirAll(metaDir, 0755)

	mockMeta := EpicMetadata{
		AppName:  "Sugar",
		AppTitle: "Rocket League",
	}
	mockMeta.Metadata.Description = "Car soccer"

	// Add an image
	mockMeta.Metadata.KeyImages = append(mockMeta.Metadata.KeyImages, struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	}{"DieselGameBoxTall", "http://example.com/cover.jpg"})

	// Add Mac Native release info
	mockMeta.Metadata.ReleaseInfo = append(mockMeta.Metadata.ReleaseInfo, struct {
		Platform []string `json:"platform"`
	}{Platform: []string{"Mac"}})

	data, _ := json.Marshal(mockMeta)
	os.WriteFile(filepath.Join(metaDir, "Sugar.json"), data, 0644)

	cover, hero, desc, isMacNative := LoadEpicMetadata(tempDir, "Sugar")

	if cover != "http://example.com/cover.jpg" {
		t.Errorf("Expected custom cover, got %s", cover)
	}
	if hero == "" {
		t.Errorf("Expected fallback hero to be populated")
	}
	if desc != "Car soccer" {
		t.Errorf("Expected custom description, got %s", desc)
	}
	if !isMacNative {
		t.Errorf("Expected isMacNative to be true based on ReleaseInfo")
	}
}
