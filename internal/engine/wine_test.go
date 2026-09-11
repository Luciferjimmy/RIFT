package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFlattenDir_SingleSubdir(t *testing.T) {
	tempDir := t.TempDir()

	// Create a nested dir: tempDir/wrapped/file.txt
	wrappedDir := filepath.Join(tempDir, "wrapped")
	os.MkdirAll(wrappedDir, 0755)

	os.WriteFile(filepath.Join(wrappedDir, "file.txt"), []byte("data"), 0644)

	err := flattenDir(tempDir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// "wrapped" should be gone, "file.txt" should be in tempDir
	if _, err := os.Stat(filepath.Join(tempDir, "wrapped")); !os.IsNotExist(err) {
		t.Errorf("Expected wrapped directory to be deleted")
	}
	if _, err := os.Stat(filepath.Join(tempDir, "file.txt")); err != nil {
		t.Errorf("Expected file.txt to be moved up")
	}
}

func TestFlattenDir_GcenxWine(t *testing.T) {
	tempDir := t.TempDir()

	// Create: tempDir/Wine Crossover.app/Contents/Resources/wine/bin/wine
	wineRoot := filepath.Join(tempDir, "Wine Crossover.app", "Contents", "Resources", "wine")
	wineBin := filepath.Join(wineRoot, "bin")
	os.MkdirAll(wineBin, 0755)
	os.WriteFile(filepath.Join(wineBin, "wine"), []byte("binary"), 0755)
	os.WriteFile(filepath.Join(wineRoot, "lib"), []byte("libdata"), 0644)

	err := flattenDir(tempDir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// The contents of `wine` (bin, lib) should now be inside tempDir directly
	if _, err := os.Stat(filepath.Join(tempDir, "bin", "wine")); err != nil {
		t.Errorf("Expected bin/wine to be at the root of tempDir")
	}
	if _, err := os.Stat(filepath.Join(tempDir, "lib")); err != nil {
		t.Errorf("Expected lib to be at the root of tempDir")
	}

	// Wine Crossover.app should be gone or not matter since we pulled it up
	if _, err := os.Stat(filepath.Join(tempDir, "Wine Crossover.app")); !os.IsNotExist(err) {
		t.Errorf("Expected Wine Crossover.app to be overwritten/removed")
	}
}

func TestIsSonomaOrLater(t *testing.T) {
	// Just test that it runs without panicking
	res := IsSonomaOrLater()

	// Reset the global cache variable to test idempotency
	isSonoma = nil
	res2 := IsSonomaOrLater()

	if res != res2 {
		t.Errorf("Expected cached result to match")
	}
}
