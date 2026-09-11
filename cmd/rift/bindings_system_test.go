package rift

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlatformFromAppID(t *testing.T) {
	tests := []struct {
		appID    string
		expected string
	}{
		{"730", "Steam"},
		{"123456", "Steam"},
		{"appID_Sugar", "Epic"},                      // Typical Legendary format for old games
		{"1234567890abcdef1234567890abcdef", "Epic"}, // 32 char md5 is typical for Epic app names
		{"Sugar", "Epic"},
		{"MyCoolGame", "Epic"},
	}

	for _, tt := range tests {
		t.Run(tt.appID, func(t *testing.T) {
			result := platformFromAppID(tt.appID)
			if result != tt.expected {
				t.Errorf("platformFromAppID(%q) = %q; want %q", tt.appID, result, tt.expected)
			}
		})
	}
}

func TestResolveDataDir(t *testing.T) {
	// Let's create a temporary ".rift/data" folder and set HOME to it.
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	expectedDir := filepath.Join(tempDir, ".rift", "data")
	os.MkdirAll(expectedDir, 0755)

	// Since we can't easily mock `os.Executable`, and assuming CWD/data doesn't exist during go test,
	// it should fall back to HOME/.rift/data
	resolved := resolveDataDir()

	if resolved != expectedDir && resolved != "data" {
		t.Errorf("Expected %s or 'data', got %s", expectedDir, resolved)
	}
}

func TestCheckEnginesStatus(t *testing.T) {
	app := &App{}
	statusJSON := app.CheckEnginesStatus()
	if statusJSON == "" {
		t.Fatalf("expected non-empty JSON response from CheckEnginesStatus")
	}

	t.Logf("CheckEnginesStatus output: %s", statusJSON)
}

func TestCheckSystemDependencies(t *testing.T) {
	app := &App{}
	isReady := app.CheckSystemDependencies()
	t.Logf("CheckSystemDependencies isReady: %v", isReady)
}


