package rift

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetGamesDir_EmptyFolders(t *testing.T) {
	app := &App{
		LibraryFolders: []string{},
	}

	dir := app.GetGamesDir()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".rift", "games")

	if dir != expected {
		t.Errorf("Expected %s, got %s", expected, dir)
	}
}

func TestResolveGameDir_ExistingFolder(t *testing.T) {
	tempDir := t.TempDir()
	folder1 := filepath.Join(tempDir, "Library1")
	folder2 := filepath.Join(tempDir, "Library2")

	os.MkdirAll(folder1, 0755)
	os.MkdirAll(folder2, 0755)

	// Create the game inside folder2
	os.MkdirAll(filepath.Join(folder2, "GameX"), 0755)

	app := &App{
		LibraryFolders: []string{folder1, folder2},
		DefaultLibrary: 0,
	}

	resolved := app.ResolveGameDir("GameX")
	expected := filepath.Join(folder2, "GameX")

	if resolved != expected {
		t.Errorf("Expected %s, got %s", expected, resolved)
	}
}
