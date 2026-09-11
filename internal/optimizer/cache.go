package optimizer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	cacheDirName  = "game_configs"
	cacheFilePerm = 0644
	// cacheExpiry is how long a cached config is considered valid.
	cacheExpiry = 30 * 24 * time.Hour // 30 days
)

// Cache persists computed game configurations to disk so the pipeline can
// be skipped on subsequent launches when inputs have not changed.
type Cache struct {
	rootDir string
}

// NewCache creates a cache rooted under ~/.rift/cache/game_configs.
func NewCache() *Cache {
	home, _ := os.UserHomeDir()
	return &Cache{
		rootDir: filepath.Join(home, ".rift", "cache", cacheDirName),
	}
}

// cachePath returns the on-disk path for a given game.
func (c *Cache) cachePath(gameID string) string {
	cleanID := filepath.Clean(gameID)
	// Additional check for traversal attempts
	if strings.Contains(cleanID, "..") || strings.Contains(cleanID, string(filepath.Separator)) {
		cleanID = "invalid_id"
	}
	return filepath.Join(c.rootDir, cleanID+".json")
}

// Get returns the cached config for gameID if it exists and has not expired.
// The systemHash is compared to detect hardware/engine changes.
func (c *Cache) Get(gameID, systemHash string) (*GameConfig, bool) {
	path := c.cachePath(gameID)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var cached CachedConfig
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, false
	}

	if time.Now().After(cached.ExpiresAt) {
		return nil, false // expired
	}
	if cached.SystemHash != systemHash {
		return nil, false // hardware or engine state changed
	}

	return &cached.Config, true
}

// Set stores a configuration in the cache.
func (c *Cache) Set(gameID, systemHash string, cfg *GameConfig) error {
	if err := os.MkdirAll(c.rootDir, 0755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	cached := CachedConfig{
		Config:     *cfg,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(cacheExpiry),
		SystemHash: systemHash,
	}

	data, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}

	if err := os.WriteFile(c.cachePath(cfg.GameID), data, cacheFilePerm); err != nil {
		return fmt.Errorf("write cache: %w", err)
	}
	return nil
}

// Invalidate removes the cache entry for a game, forcing re-computation.
func (c *Cache) Invalidate(gameID string) error {
	path := c.cachePath(gameID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove cache: %w", err)
	}
	return nil
}
