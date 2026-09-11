package rift

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGetAuthStatus_SteamMock(t *testing.T) {
	tempDir := t.TempDir()

	app := &App{
		RiftBaseDir: tempDir,
	}

	steamDir := filepath.Join(app.StoresDir(), "steam")
	steamPrefix := filepath.Join(steamDir, "prefix", "drive_c", "Program Files (x86)", "Steam")
	os.MkdirAll(filepath.Join(steamPrefix, "config"), 0755)

	// Create steam.exe to satisfy "SteamInstalled"
	os.WriteFile(filepath.Join(steamPrefix, "steam.exe"), []byte("mock binary"), 0755)

	conf := StoreConfig{Ready: true, Username: "MockSteamUser"}
	data, _ := json.Marshal(conf)
	os.MkdirAll(filepath.Dir(filepath.Join(steamDir, "store.json")), 0755)
	os.WriteFile(filepath.Join(steamDir, "store.json"), data, 0644)

	// Create loginusers.vdf inside the store prefix
	vdfContent := `"users"
{
	"76561198000000001"
	{
		"PersonaName"		"MockSteamUser"
		"MostRecent"		"1"
	}
}
`
	os.WriteFile(filepath.Join(steamPrefix, "config", "loginusers.vdf"), []byte(vdfContent), 0644)

	status := app.GetAuthStatus()

	if !status.SteamInstalled {
		t.Errorf("Expected SteamInstalled = true")
	}
	if !status.SteamConnected {
		t.Errorf("Expected SteamConnected = true")
	}
	if status.SteamUsername != "MockSteamUser" {
		t.Errorf("Expected SteamUsername = MockSteamUser, got %s", status.SteamUsername)
	}
}