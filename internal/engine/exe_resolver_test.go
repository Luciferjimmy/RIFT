package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestIsNonGameBinary(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"7z.exe", true},
		{"7za.exe", true},
		{"crashreport.exe", true},
		{"UnityCrashHandler64.exe", true},
		{"unins000.exe", true},
		{"uninstall.exe", true},
		{"setup.exe", true},
		{"vcredist_x64.exe", true},
		{"dxsetup.exe", true},
		{"EOSBootStrapper.exe", true},
		{"EpicBootstrapper.exe", true},
		{"EasyAntiCheat_Setup.exe", true},
		{"Touchup.exe", true},
		{"Dungeonhaven.exe", false},
		{"Cyberpunk2077.exe", false},
		{"BioShockHD.exe", false},
		{"GenshinImpact.exe", false},
		{"launcher.exe", false},
	}

	for _, tc := range tests {
		got := IsNonGameBinary(tc.name)
		if got != tc.expected {
			t.Errorf("IsNonGameBinary(%q) = %v; want %v", tc.name, got, tc.expected)
		}
	}
}

func TestResolveGameExecutable_RejectsOnlyUtilityBinaries(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rift-exe-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directory structure mirroring the Genshin partial extraction
	gameFiles := filepath.Join(tmpDir, "game_files", "GenshinImpact")
	os.MkdirAll(gameFiles, 0755)

	_ = os.WriteFile(filepath.Join(gameFiles, "7z.exe"), []byte("MZdummy7z"), 0755)
	_ = os.WriteFile(filepath.Join(gameFiles, "7z.dll"), []byte("MZdummy7zdll"), 0644)
	_ = os.WriteFile(filepath.Join(gameFiles, "crashreport.exe"), []byte("MZdummycrash"), 0755)

	bestExe, _, err := ResolveGameExecutable(tmpDir, "41869934302e4b8cafac2d3c0e7c293d", "Genshin Impact", "")
	if err == nil {
		t.Fatalf("Expected error when only utility binaries are present, got bestExe=%q", bestExe)
	}
	if bestExe != "" {
		t.Fatalf("Expected empty bestExe, got %q", bestExe)
	}
}

func TestResolveGameExecutable_PrioritizesRealGameExe(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rift-exe-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gameFiles := filepath.Join(tmpDir, "game_files")
	os.MkdirAll(gameFiles, 0755)

	// Create 7z.exe (utility) and RealGame.exe (game)
	_ = os.WriteFile(filepath.Join(gameFiles, "7z.exe"), []byte("MZdummy7z"), 0755)
	_ = os.WriteFile(filepath.Join(gameFiles, "RealGame.exe"), []byte("MZdummygame"), 0755)

	bestExe, _, err := ResolveGameExecutable(tmpDir, "12345", "Real Game", "")
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	expected := filepath.Join("game_files", "RealGame.exe")
	if bestExe != expected {
		t.Errorf("Expected bestExe=%q, got %q", expected, bestExe)
	}
}

func TestResolveGameExecutable_Strategy0RejectsUtilityInRuntimeConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rift-exe-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gameFiles := filepath.Join(tmpDir, "game_files")
	os.MkdirAll(gameFiles, 0755)
	_ = os.WriteFile(filepath.Join(gameFiles, "7z.exe"), []byte("MZdummy7z"), 0755)

	// Write game_runtime.json pointing to 7z.exe
	rconf := map[string]interface{}{
		"executable": filepath.Join("game_files", "7z.exe"),
		"installed":  true,
	}
	data, _ := json.Marshal(rconf)
	_ = os.WriteFile(filepath.Join(tmpDir, "game_runtime.json"), data, 0644)

	bestExe, _, err := ResolveGameExecutable(tmpDir, "dummy-id", "dummy-name", "")
	if err == nil {
		t.Fatalf("Expected error because runtime config points to 7z.exe and no other game exe exists, got bestExe=%q", bestExe)
	}
}
