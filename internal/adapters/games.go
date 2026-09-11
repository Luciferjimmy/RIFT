package adapters

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	scraperCacheMu sync.RWMutex
	scraperCache   = make(map[string]string)
	scraperLoaded  bool
)

func loadScraperCacheOnce() {
	scraperCacheMu.RLock()
	if scraperLoaded {
		scraperCacheMu.RUnlock()
		return
	}
	scraperCacheMu.RUnlock()

	scraperCacheMu.Lock()
	defer scraperCacheMu.Unlock()
	if scraperLoaded {
		return
	}
	scraperLoaded = true

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	cachePath := filepath.Join(home, ".rift", "data", "scraper_cache.db")
	if _, err := os.Stat(cachePath); err != nil {
		return
	}

	db, err := sql.Open("sqlite3", cachePath+"?mode=ro")
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT app_id, cipher FROM translation_profiles")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var appID, cipher string
		if err := rows.Scan(&appID, &cipher); err == nil && appID != "" {
			scraperCache[appID] = cipher
		}
	}
}

// RIFTUIProfile is the strict contract that the frontend expects.
type RIFTUIProfile struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Cover        string   `json:"cover"`        // Vertical box art (grid cards)
	HeroCover    string   `json:"heroCover"`    // Wide banner art (hero section)
	Platform     string   `json:"platform"`     // "Steam" or "Epic"
	AppID        string   `json:"appID"`        // The platform-native identifier (Steam AppID or Epic app_name)
	Status       string   `json:"status"`       // "ready", "running", "downloading", "not_installed"
	Playtime     string   `json:"playtime"`     // e.g. "42h"
	LastPlayed   string   `json:"lastPlayed"`   // e.g. "3 days ago"
	LastPlayedTs int64    `json:"lastPlayedTs"` // Unix timestamp for sorting
	Backend      string   `json:"backend"`      // "macOS Native", "Wine + D3DMetal (DX11/12)", "Wine + DXVK (DX9/10)"
	Description  string   `json:"description"`
	Tags         []string `json:"tags"`
	IsInstalled  bool     `json:"isInstalled"`
	IsMacNative  bool     `json:"isMacNative"`
}

// EpicMetadata matches the JSON structure stored inside Legendary's metadata cache files.
type EpicMetadata struct {
	AppName  string `json:"app_name"`
	AppTitle string `json:"app_title"`
	Metadata struct {
		Description string `json:"description"`
		KeyImages   []struct {
			Type string `json:"type"`
			URL  string `json:"url"`
		} `json:"keyImages"`
		ReleaseInfo []struct {
			Platform []string `json:"platform"`
		} `json:"releaseInfo"`
	} `json:"metadata"`
}

// LoadEpicMetadata parses the local metadata cache to extract real cover art, banner art, description, and macOS native support.
func LoadEpicMetadata(globalAuthDir string, appName string) (string, string, string, bool) {
	metaFile := filepath.Join(globalAuthDir, "metadata", appName+".json")
	data, err := os.ReadFile(metaFile)
	if err != nil {
		// Fallback to default user config path if not found in isolated path
		home, _ := os.UserHomeDir()
		metaFile = filepath.Join(home, ".config", "legendary", "metadata", appName+".json")
		data, err = os.ReadFile(metaFile)
	}

	cover := ""
	hero := ""
	desc := ""
	isMacNative := false

	if err == nil {
		var meta EpicMetadata
		if err := json.Unmarshal(data, &meta); err == nil {
			desc = meta.Metadata.Description
			for _, img := range meta.Metadata.KeyImages {
				if img.Type == "DieselGameBoxTall" {
					cover = img.URL
				} else if img.Type == "DieselGameBox" {
					hero = img.URL
				}
			}

			// Scan releaseInfo to detect native macOS builds
			for _, release := range meta.Metadata.ReleaseInfo {
				for _, plat := range release.Platform {
					if strings.ToLower(plat) == "mac" || strings.ToLower(plat) == "macos" {
						isMacNative = true
					}
				}
			}
		}
	}

	// Premium Unsplash fallbacks if images aren't present in metadata
	if cover == "" {
		cover = "https://images.unsplash.com/photo-1542751371-adc38448a05e?w=600"
	}
	if hero == "" {
		hero = "https://images.unsplash.com/photo-1542751371-adc38448a05e?w=1200"
	}
	if desc == "" {
		desc = "No description available."
	}

	return cover, hero, desc, isMacNative
}

// AdaptEpicGame maps a Legendary JSON game entry to the UI contract.
func AdaptEpicGame(appName string, title string, isInstalled bool, globalAuthDir string, dbBackend string) RIFTUIProfile {
	cover, hero, desc, isMacNative := LoadEpicMetadata(globalAuthDir, appName)

	backend := getActualBackend(appName, isInstalled, isMacNative, dbBackend)
	playtime := getPlaytimeString(appName)

	return RIFTUIProfile{
		ID:          appName,
		Name:        title,
		Cover:       cover,
		HeroCover:   hero,
		Platform:    "Epic",
		AppID:       appName,
		Status:      statusFromInstalled(isInstalled),
		Playtime:    playtime,
		LastPlayed:  "—",
		Backend:     backend,
		Description: desc,
		Tags:        []string{"Epic Games"},
		IsInstalled: isInstalled,
		IsMacNative: isMacNative,
	}
}

// AdaptSteamGame maps a Steam AppID to the UI contract.
func AdaptSteamGame(appID string, title string, isInstalled bool, isMacNative bool, dbBackend string) RIFTUIProfile {
	cover := "https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/" + appID + "/header.jpg"
	hero := "https://steamcdn-a.akamaihd.net/steam/apps/" + appID + "/library_hero.jpg"

	backend := getActualBackend(appID, isInstalled, isMacNative, dbBackend)
	playtime := getPlaytimeString("steam-" + appID)

	return RIFTUIProfile{
		ID:          "steam-" + appID,
		Name:        title,
		Cover:       cover,
		HeroCover:   hero,
		Platform:    "Steam",
		AppID:       appID,
		Status:      statusFromInstalled(isInstalled),
		Playtime:    playtime,
		LastPlayed:  "—",
		Backend:     backend,
		Description: title,
		Tags:        []string{"Steam"},
		IsInstalled: isInstalled,
		IsMacNative: isMacNative,
	}
}

func statusFromInstalled(isInstalled bool) string {
	if isInstalled {
		return "ready"
	}
	return "not_installed"
}

// getActualBackend determines the display string for the game's engine
func getActualBackend(gameID string, isInstalled bool, isMacNative bool, dbBackend string) string {
	if isMacNative {
		return "macOS Native"
	}
	if !isInstalled {
		// First check scraper_cache.db in-memory cache
		loadScraperCacheOnce()
		scraperCacheMu.RLock()
		if cipher, ok := scraperCache[gameID]; ok && cipher != "" {
			scraperCacheMu.RUnlock()
			return cipher
		}
		scraperCacheMu.RUnlock()

		if dbBackend != "" {
			if strings.Contains(strings.ToLower(dbBackend), "dxvk") || strings.Contains(strings.ToLower(dbBackend), "dx9") || strings.Contains(strings.ToLower(dbBackend), "d9vk") || strings.Contains(strings.ToLower(dbBackend), "opengl") {
				return "Wine + DXVK (DX9/10/11) - Pending"
			}
			return "Wine + D3DMetal (DX11/12) - Pending"
		}
		return "Awaiting Translation Profile"
	}

	// Try reading rift_config.json
	home, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(home, ".rift", "games", gameID, "rift_config.json")
		if data, err := os.ReadFile(configPath); err == nil {
			var cfg struct {
				Engine string `json:"engine"`
			}
			if json.Unmarshal(data, &cfg) == nil {
				if cfg.Engine == "dxvk" {
					return "Wine + DXVK (DX9/10/11)"
				} else {
					return "Wine + D3DMetal (DX11/12)"
				}
			}
		}
	}

	// Fallback to dbBackend or D3DMetal
	if strings.Contains(strings.ToLower(dbBackend), "dxvk") || strings.Contains(strings.ToLower(dbBackend), "dx9") || strings.Contains(strings.ToLower(dbBackend), "d9vk") || strings.Contains(strings.ToLower(dbBackend), "opengl") {
		return "Wine + DXVK (DX9/10/11)"
	}
	return "Wine + D3DMetal (DX11/12)"
}

func getPlaytimeString(gameID string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "—"
	}
	playtimeFile := filepath.Join(home, ".rift", "data", "playtime.json")
	data, err := os.ReadFile(playtimeFile)
	if err != nil {
		return "—"
	}

	var pt struct {
		PerGame map[string]int64 `json:"per_game"`
	}
	if err := json.Unmarshal(data, &pt); err == nil && pt.PerGame != nil {
		secs, ok := pt.PerGame[gameID]
		if ok && secs > 0 {
			hours := float64(secs) / 3600.0
			if hours < 1.0 {
				return fmt.Sprintf("%.1f min", float64(secs)/60.0)
			}
			return fmt.Sprintf("%.1f hrs", hours)
		}
	}
	return "—"
}
