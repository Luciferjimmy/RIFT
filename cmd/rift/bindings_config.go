package rift

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"rift/internal/engine"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// loadConfig ensures default configuration exists.
func (a *App) loadConfig() {
	configPath := filepath.Join(a.RiftDir(), "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(configPath), 0755)
		defaultConfig := AppConfig{
			LibraryFolders: []string{filepath.Join(a.RiftDir(), "games")},
			DefaultLibrary: 0,
			ImportedGames:  make(map[string]string),
			GameOverrides:  make(map[string]GameOverride),
		}
		data, _ := json.MarshalIndent(defaultConfig, "", "  ")
		_ = atomicWriteFile(configPath, data, 0644)
		a.LibraryFolders = defaultConfig.LibraryFolders
		a.DefaultLibrary = defaultConfig.DefaultLibrary
		return
	}

	data, err := os.ReadFile(configPath)
	if err == nil {
		var cfg AppConfig
		if err := json.Unmarshal(data, &cfg); err == nil {

			if len(cfg.LibraryFolders) == 0 && cfg.LibraryDir != "" {
				cfg.LibraryFolders = []string{cfg.LibraryDir}
				cfg.LibraryDir = ""
				a.SaveFullConfig(cfg)
			}

			if len(cfg.LibraryFolders) > 0 {
				a.LibraryFolders = cfg.LibraryFolders
				a.DefaultLibrary = cfg.DefaultLibrary
				if a.DefaultLibrary >= len(a.LibraryFolders) {
					a.DefaultLibrary = 0
				}
			} else {
				a.LibraryFolders = []string{filepath.Join(a.RiftDir(), "games")}
				a.DefaultLibrary = 0
			}
			return
		}
	}
	a.LibraryFolders = []string{filepath.Join(a.RiftDir(), "games")}
	a.DefaultLibrary = 0
}

// saveConfig is a legacy function for quickly appending a new library folder
func (a *App) saveConfig(libraryDir string) {
	cfg := a.GetFullConfig()

	found := false
	for i, folder := range cfg.LibraryFolders {
		if folder == libraryDir {
			cfg.DefaultLibrary = i
			found = true
			break
		}
	}
	if !found {
		cfg.LibraryFolders = append(cfg.LibraryFolders, libraryDir)
		cfg.DefaultLibrary = len(cfg.LibraryFolders) - 1
	}
	a.SaveFullConfig(cfg)
}

// GetFullConfig reads and returns the entire AppConfig from disk
func (a *App) GetFullConfig() AppConfig {
	configPath := filepath.Join(a.RiftDir(), "config.json")
	var cfg AppConfig
	cfg.ImportedGames = make(map[string]string)
	cfg.GameOverrides = make(map[string]GameOverride)

	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &cfg)
	}
	if cfg.ImportedGames == nil {
		cfg.ImportedGames = make(map[string]string)
	}
	if cfg.GameOverrides == nil {
		cfg.GameOverrides = make(map[string]GameOverride)
	}
	return cfg
}

// atomicWriteFile writes data to a temporary file first and renames it atomically
// to prevent file corruption if the app crashes or the machine loses power mid-write.
func atomicWriteFile(filePath string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmpFile := fmt.Sprintf("%s.tmp.%d", filePath, os.Getpid())
	if err := os.WriteFile(tmpFile, data, perm); err != nil {
		return err
	}
	return os.Rename(tmpFile, filePath)
}

// SaveFullConfig saves the entire AppConfig struct to disk atomically.
func (a *App) SaveFullConfig(cfg AppConfig) {
	configPath := filepath.Join(a.RiftDir(), "config.json")
	a.LibraryFolders = cfg.LibraryFolders
	a.DefaultLibrary = cfg.DefaultLibrary
	newData, _ := json.MarshalIndent(cfg, "", "  ")
	if err := atomicWriteFile(configPath, newData, 0644); err != nil {
		a.logError("[Config] Failed to atomically write config.json: %v", err)
	}

	if cfg.DisableTelemetry {
		home, _ := os.UserHomeDir()
		_ = os.RemoveAll(filepath.Join(home, ".rift", "data", "pending_telemetry"))
	}
}

// SelectLibraryFolder lets the user choose a directory and appends it to LibraryFolders
func (a *App) SelectLibraryFolder() string {
	path, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select RIFT Games Directory",
	})
	if err != nil || path == "" {
		if len(a.LibraryFolders) > 0 {
			return a.LibraryFolders[a.DefaultLibrary]
		}
		return ""
	}

	a.saveConfig(path)
	a.logInfo("[Config] Library directory updated to: %s", path)
	return path
}

// GetGamesDir returns the default games library directory path.
func (a *App) GetGamesDir() string {
	if len(a.LibraryFolders) == 0 {
		return filepath.Join(a.RiftDir(), "games")
	}
	if a.DefaultLibrary >= len(a.LibraryFolders) {
		a.DefaultLibrary = 0
	}
	return a.LibraryFolders[a.DefaultLibrary]
}

// ResolveGameDir scans all registered library folders and returns the path to the game capsule.
// If the game does not exist anywhere (or drives are unplugged), it returns the default path so new downloads go there.
func (a *App) ResolveGameDir(gameID string) string {
	for _, folder := range a.LibraryFolders {

		if _, err := os.Stat(folder); os.IsNotExist(err) {
			continue
		}

		gamePath := filepath.Join(folder, gameID)
		if _, err := os.Stat(gamePath); err == nil {
			return gamePath
		}
	}

	return filepath.Join(a.GetGamesDir(), gameID)
}

// GameExecutableInfo provides metadata about an executable binary found inside a game capsule.
type GameExecutableInfo struct {
	RelativePath string `json:"relativePath"`
	FileName     string `json:"fileName"`
	SizeBytes    int64  `json:"sizeBytes"`
	IsSelected   bool   `json:"isSelected"`
	Engine       string `json:"engine"`
	DirectXAPI   string `json:"directXAPI"`
	Is32Bit      bool   `json:"is32Bit"`
}

// GetGameExecutables scans the game folder for all valid executable candidates.
func (a *App) GetGameExecutables(gameID string) ([]GameExecutableInfo, error) {
	gameDir := a.ResolveGameDir(gameID)
	gameFilesDir := filepath.Join(gameDir, "game_files")
	if _, err := os.Stat(gameFilesDir); os.IsNotExist(err) {
		gameFilesDir = gameDir
	}

	// Read current selected executable from game_runtime.json
	runtimePath := a.runtimeConfigPath(gameID)
	currentSelected := ""
	if data, err := os.ReadFile(runtimePath); err == nil {
		var rt struct {
			Executable     string `json:"executable"`
			ExecutablePath string `json:"executable_path"`
		}
		if json.Unmarshal(data, &rt) == nil {
			if rt.Executable != "" {
				currentSelected = filepath.Clean(rt.Executable)
			} else if rt.ExecutablePath != "" {
				currentSelected = filepath.Clean(rt.ExecutablePath)
			}
		}
	}

	var results []GameExecutableInfo
	ignoredNames := map[string]bool{
		"crashreportclient.exe":   true,
		"crashreport.exe":         true,
		"unins000.exe":            true,
		"uninstall.exe":           true,
		"uninstaller.exe":         true,
		"easyanticheat_setup.exe": true,
		"unrealcefsubprocess.exe": true,
		"dxsetup.exe":             true,
		"vcredist_x64.exe":        true,
		"vcredist_x86.exe":        true,
		"ue4prereqsetup_x64.exe":  true,
		"epicgameslauncher.exe":   true,
		"7z.exe":                  true,
	}

	filepath.Walk(gameFilesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}

		lowerName := strings.ToLower(info.Name())
		if strings.HasSuffix(lowerName, ".exe") && !ignoredNames[lowerName] {
			rel, _ := filepath.Rel(gameFilesDir, path)
			cleanRel := filepath.Clean(rel)

			peInfo, _ := engine.DetectGameGraphicsEngine(path)
			engineType := "d3dmetal"
			directXAPI := "Unknown"
			is32Bit := false
			if peInfo != nil {
				engineType = peInfo.Engine
				directXAPI = peInfo.DirectXAPI
				is32Bit = peInfo.Is32Bit
			}

			isSelected := false
			if currentSelected != "" {
				if cleanRel == currentSelected || filepath.Base(cleanRel) == currentSelected || strings.HasSuffix(currentSelected, cleanRel) {
					isSelected = true
				}
			}

			results = append(results, GameExecutableInfo{
				RelativePath: cleanRel,
				FileName:     info.Name(),
				SizeBytes:    info.Size(),
				IsSelected:   isSelected,
				Engine:       engineType,
				DirectXAPI:   directXAPI,
				Is32Bit:      is32Bit,
			})
		}
		return nil
	})

	// If none was marked selected, mark the first as selected
	if len(results) > 0 {
		hasSelected := false
		for _, r := range results {
			if r.IsSelected {
				hasSelected = true
				break
			}
		}
		if !hasSelected {
			results[0].IsSelected = true
		}
	}

	return results, nil
}

// SetGameExecutable updates the active executable for a game in game_runtime.json.
func (a *App) SetGameExecutable(gameID, relativeExePath string) error {
	runtimePath := a.runtimeConfigPath(gameID)
	raw := make(map[string]interface{})
	if data, err := os.ReadFile(runtimePath); err == nil {
		json.Unmarshal(data, &raw)
	}

	raw["executable"] = relativeExePath
	raw["executable_path"] = relativeExePath

	// Detect graphics engine for the new executable
	gameDir := a.ResolveGameDir(gameID)
	gameFilesDir := filepath.Join(gameDir, "game_files")
	fullPath := filepath.Join(gameFilesDir, relativeExePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		fullPath = filepath.Join(gameDir, relativeExePath)
	}

	if peInfo, err := engine.DetectGameGraphicsEngine(fullPath); err == nil && peInfo != nil {
		// Only update engine if manual override is not enabled
		if mo, ok := raw["manual_override"].(bool); !ok || !mo {
			raw["engine"] = peInfo.Engine
		}
	}

	b, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal runtime config: %w", err)
	}

	if err := atomicWriteFile(runtimePath, b, 0644); err != nil {
		a.logError("[Config] Failed to save runtime config to %s: %v", runtimePath, err)
	}

	a.logInfo("[Config] Primary executable for %s set to %s", gameID, relativeExePath)
	return nil
}
