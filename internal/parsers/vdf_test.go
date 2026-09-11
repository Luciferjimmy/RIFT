package parsers

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCheckIfGameInstalled(t *testing.T) {
	tempDir := t.TempDir()

	// Game 730 is installed
	acfPath := filepath.Join(tempDir, "appmanifest_730.acf")
	os.WriteFile(acfPath, []byte("fake content"), 0644)

	if !CheckIfGameInstalled(tempDir, "730") {
		t.Errorf("Expected 730 to be installed")
	}

	if CheckIfGameInstalled(tempDir, "123") {
		t.Errorf("Expected 123 to not be installed")
	}
}

func TestParseSteamLoginUser(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	os.MkdirAll(configDir, 0755)

	vdfData := `"users"
{
	"76561198000000001"
	{
		"PersonaName"		"OldUser"
		"RememberPassword"		"1"
		"MostRecent"		"0"
	}
	"76561198000000002"
	{
		"PersonaName"		"ActiveUser"
		"RememberPassword"		"1"
		"MostRecent"		"1"
	}
}
`
	vdfPath := filepath.Join(configDir, "loginusers.vdf")
	os.WriteFile(vdfPath, []byte(vdfData), 0644)

	user, err := ParseSteamLoginUser(tempDir)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if user != "ActiveUser" {
		t.Errorf("Expected ActiveUser, got %s", user)
	}
}

func TestParseSteamLoginUser_NoMostRecent(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	os.MkdirAll(configDir, 0755)

	vdfData := `"users"
{
	"76561198000000001"
	{
		"PersonaName"		"OnlyUser"
	}
}
`
	vdfPath := filepath.Join(configDir, "loginusers.vdf")
	os.WriteFile(vdfPath, []byte(vdfData), 0644)

	user, err := ParseSteamLoginUser(tempDir)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if user != "OnlyUser" {
		t.Errorf("Expected OnlyUser, got %s", user)
	}
}

func TestParseOwnedSteamGames(t *testing.T) {
	tempDir := t.TempDir()
	userdataDir := filepath.Join(tempDir, "userdata", "12345", "config")
	os.MkdirAll(userdataDir, 0755)

	vdfData := `"UserLocalConfigStore"
{
	"Software"
	{
		"Valve"
		{
			"Steam"
			{
				"apps"
				{
					"730"
					{
						"Playtime" "100"
					}
					"400"
					{
						"Playtime" "20"
					}
				}
			}
		}
	}
}
`
	os.WriteFile(filepath.Join(userdataDir, "localconfig.vdf"), []byte(vdfData), 0644)

	games, err := ParseOwnedSteamGames(tempDir, "12345")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Order doesn't matter, just check length and elements
	if len(games) != 2 {
		t.Fatalf("Expected 2 games, got %d", len(games))
	}

	gameMap := make(map[string]bool)
	for _, g := range games {
		gameMap[g] = true
	}

	if !gameMap["730"] || !gameMap["400"] {
		t.Errorf("Expected 730 and 400, got %v", games)
	}
}

func TestParseLibraryFolders(t *testing.T) {
	tempDir := t.TempDir()
	vdfPath := filepath.Join(tempDir, "libraryfolders.vdf")

	vdfData := `"libraryfolders"
{
	"0"
	{
		"path"		"/Users/user/Library/Application Support/Steam"
	}
	"1"
	{
		"path"		"/Volumes/ExternalDrive/SteamLibrary"
	}
}
`
	os.WriteFile(vdfPath, []byte(vdfData), 0644)

	folders := ParseLibraryFolders(vdfPath)
	expected := []string{
		"/Users/user/Library/Application Support/Steam",
		"/Volumes/ExternalDrive/SteamLibrary",
	}

	if !reflect.DeepEqual(folders, expected) {
		t.Errorf("Expected %v, got %v", expected, folders)
	}
}
