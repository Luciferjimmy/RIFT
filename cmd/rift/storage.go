package rift

import (
	"os"
	"path/filepath"
)

type StorageStats struct {
	TotalGamesSize  int64 `json:"totalGamesSize"`
	TotalCacheSize  int64 `json:"totalCacheSize"`
	TotalEngineSize int64 `json:"totalEngineSize"`
}

func (a *App) GetStorageStats() StorageStats {
	var stats StorageStats

	home, _ := os.UserHomeDir()
	appDir := filepath.Join(home, ".rift")
	enginesDir := filepath.Join(appDir, "engines")

	// Game files (and their prefixes) across all library folders
	for _, folder := range a.LibraryFolders {
		stats.TotalGamesSize += getDirSize(folder)
	}

	// Caches (e.g. downloaded installers/temp)
	stats.TotalCacheSize = getDirSize(filepath.Join(appDir, "cache"))

	// Engines (e.g. wine builds, crossover layers)
	stats.TotalEngineSize = getDirSize(enginesDir)

	return stats
}

func (a *App) PurgeCache() bool {
	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".rift", "cache")
	os.RemoveAll(cacheDir)
	os.MkdirAll(cacheDir, 0755)
	return true
}

func getDirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // ignore errors
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}
