package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const RIFT_DIR = ".rift"
const GAMES_DIR = "games"
const RUNTIMES_DIR = "runtimes"
const GLOBAL_CONFIG = "rift.json"
const GAME_CONFIG = "config.json"

type RuntimePaths struct {
	WineBin     string `json:"wine_bin"`
	Winetricks  string `json:"winetricks_bin"`
	MoltenVKLib string `json:"moltenvk_lib"`
	GPTKLib     string `json:"gptk_lib"`
	DXVKPath    string `json:"dxvk_path"`
}

type GlobalConfig struct {
	GamesPath    string       `json:"games_path"`
	RuntimesPath string       `json:"runtimes_path"`
	Runtimes     RuntimePaths `json:"runtimes"`
}

type GameConfig struct {
	Name       string            `json:"name"`
	Exe        string            `json:"exe"`
	Gamefiles  string            `json:"gamefiles"`
	Engine     string            `json:"engine"`
	DX         string            `json:"dx"`
	Backend    string            `json:"backend"`
	Env        map[string]string `json:"env"`
	Args       []string          `json:"args"`
	Winetricks []string          `json:"winetricks"`
}

func RiftDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, RIFT_DIR)
}

func GamesDir() string {
	return filepath.Join(RiftDir(), GAMES_DIR)
}

func RuntimesDir() string {
	return filepath.Join(RiftDir(), RUNTIMES_DIR)
}

func GlobalConfigPath() string {
	return filepath.Join(RiftDir(), GLOBAL_CONFIG)
}

func GameDir(gameID string) string {
	return filepath.Join(GamesDir(), gameID)
}

func GameConfigPath(gameID string) string {
	return filepath.Join(GameDir(gameID), GAME_CONFIG)
}

func GamePrefixDir(gameID string) string {
	return filepath.Join(GameDir(gameID), "prefix")
}

func GameFilesDir(gameID string) string {
	return filepath.Join(GameDir(gameID), "gamefiles")
}

func GameLogPath(gameID string) string {
	return filepath.Join(GameDir(gameID), "last_run.log")
}

func LoadGlobal() (*GlobalConfig, error) {
	data, err := os.ReadFile(GlobalConfigPath())
	if err != nil {
		return nil, fmt.Errorf("read global config: %w", err)
	}
	var cfg GlobalConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse global config: %w", err)
	}
	return &cfg, nil
}

func SaveGlobal(cfg *GlobalConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal global config: %w", err)
	}
	if err := os.WriteFile(GlobalConfigPath(), data, 0644); err != nil {
		return fmt.Errorf("write global config: %w", err)
	}
	return nil
}

func LoadGame(gameID string) (*GameConfig, error) {
	data, err := os.ReadFile(GameConfigPath(gameID))
	if err != nil {
		return nil, fmt.Errorf("read game config: %w", err)
	}
	var cfg GameConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse game config: %w", err)
	}
	return &cfg, nil
}

func SaveGame(gameID string, cfg *GameConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal game config: %w", err)
	}
	if err := os.WriteFile(GameConfigPath(gameID), data, 0644); err != nil {
		return fmt.Errorf("write game config: %w", err)
	}
	return nil
}

func ListGames() ([]string, error) {
	entries, err := os.ReadDir(GamesDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var games []string
	for _, e := range entries {
		if e.IsDir() {
			cfgPath := filepath.Join(GamesDir(), e.Name(), GAME_CONFIG)
			if _, err := os.Stat(cfgPath); err == nil {
				games = append(games, e.Name())
			}
		}
	}
	return games, nil
}

func GameID(name string) string {
	// Simple slug: lowercase, replace spaces with hyphens
	s := ""
	for _, c := range name {
		if c >= 'A' && c <= 'Z' {
			s += string(c + 32)
		} else if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '_' {
			s += string(c)
		} else if c == ' ' {
			s += "-"
		}
	}
	return s
}
