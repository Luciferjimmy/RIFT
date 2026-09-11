package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGameIDSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Cyberpunk 2077", "cyberpunk-2077"},
		{"The Witcher 3: Wild Hunt", "the-witcher-3-wild-hunt"},
		{"GTA V", "gta-v"},
		{"Half-Life 2", "half-life-2"},
		{"Game_with_underscores", "game_with_underscores"},
	}

	for _, tc := range tests {
		actual := GameID(tc.input)
		if actual != tc.expected {
			t.Errorf("GameID(%q) = %q; expected %q", tc.input, actual, tc.expected)
		}
	}
}

func TestConfigPaths(t *testing.T) {
	// Not an isolated test because RiftDir relies on os.UserHomeDir, but we can test relativity.
	home, _ := os.UserHomeDir()
	expectedRift := filepath.Join(home, ".rift")

	if RiftDir() != expectedRift {
		t.Errorf("Expected %s, got %s", expectedRift, RiftDir())
	}

	gameID := "test-game"
	expectedGameDir := filepath.Join(expectedRift, "games", gameID)
	if GameDir(gameID) != expectedGameDir {
		t.Errorf("Expected %s, got %s", expectedGameDir, GameDir(gameID))
	}

	expectedGameConfig := filepath.Join(expectedGameDir, "config.json")
	if GameConfigPath(gameID) != expectedGameConfig {
		t.Errorf("Expected %s, got %s", expectedGameConfig, GameConfigPath(gameID))
	}
}
