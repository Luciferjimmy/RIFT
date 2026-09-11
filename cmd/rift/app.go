package rift

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"rift/internal/cloud"
	"rift/internal/engine"
	"rift/internal/scraper"
	"rift/internal/sysmon"
	"strings"
	"sync"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// GameOverride stores manual tweaks for a specific game
type GameOverride struct {
	Engine    string `json:"engine"`
	Retina    string `json:"retina"`
	HUD       string `json:"hud"`
	Args      string `json:"args"`
	VRAMLimit int    `json:"vram_limit"`
	FPSCap    int    `json:"fps_cap"`
	PrePurge       bool   `json:"pre_purge"`
	GameMode       bool   `json:"game_mode"`
	ExecutablePath string `json:"executable_path"`
}

// AppConfig stores user settings persisting inside ~/.rift/config.json
type AppConfig struct {
	LibraryFolders   []string                `json:"library_folders"`
	DefaultLibrary   int                     `json:"default_library"`
	LibraryDir       string                  `json:"library_dir,omitempty"` // Legacy, kept for migration
	ImportedGames    map[string]string       `json:"imported_games"`        // gameID -> localPath of native app / exec
	GlobalHUD        bool                    `json:"global_hud"`
	GlobalRetina     bool                    `json:"global_retina"`
	GlobalEsync      bool                    `json:"global_esync"`
	GlobalSafeMode   bool                    `json:"global_safe_mode"`
	DisableTelemetry bool                    `json:"disable_telemetry"`
	GameOverrides    map[string]GameOverride `json:"game_overrides"` // gameID -> GameOverride
}

type AuthStatus struct {
	EpicConnected  bool   `json:"epicConnected"`
	EpicUsername   string `json:"epicUsername"`
	SteamConnected bool   `json:"steamConnected"`
	SteamUsername  string `json:"steamUsername"`
	SteamInstalled bool   `json:"steamInstalled"`
}

// App is the wails application struct that holds application state and exposes bindings.
type App struct {
	EngineManager  *engine.EngineManager
	RiftBaseDir    string
	GlobalAuthDir  string
	LibraryFolders []string
	DefaultLibrary int
	ctx            context.Context
	dls            *DownloadManager // tracks active download for cancel/progress
	Log            *RiftLogger      // session-persistent structured logger
	Supabase       *cloud.SupabaseClient
	Scraper        *scraper.BackgroundScraper
	metadataDB     *sql.DB
	metadataDBMu   sync.RWMutex
	coverArtCache  sync.Map
}

// Thread-safe cache for Steam platform compatibility queries
var steamPlatformCache = struct {
	sync.RWMutex
	data map[string]bool
}{data: make(map[string]bool)}

// RiftDir returns the base .rift directory (defaults to ~/.rift).
func (a *App) RiftDir() string {
	if a.RiftBaseDir != "" {
		return a.RiftBaseDir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".rift")
}

// EnginesDir returns the engines directory (defaults to ~/.rift/engines).
func (a *App) EnginesDir() string {
	return filepath.Join(a.RiftDir(), "engines")
}

// SteamCMDDir returns the steamcmd directory (defaults to ~/.rift/engines/steamcmd).
func (a *App) SteamCMDDir() string {
	return filepath.Join(a.EnginesDir(), "steamcmd")
}

// AuthDir returns the auth directory (defaults to ~/.rift/auth).
func (a *App) AuthDir() string {
	if a.GlobalAuthDir != "" {
		return a.GlobalAuthDir
	}
	return filepath.Join(a.RiftDir(), "auth")
}

// UserDir returns the directory scoped to the currently authenticated user (~/.rift/users/<user_id>).
// If no user is logged in, falls back to ~/.rift/users/default.
func (a *App) UserDir() string {
	subDir := "default"
	if a.Supabase != nil && a.Supabase.IsLoggedIn() && a.Supabase.UserID != "" {
		subDir = a.Supabase.UserID
	}
	return filepath.Join(a.RiftDir(), "users", subDir)
}

// StoresDir returns the stores directory where store prefixes live, scoped per user.
func (a *App) StoresDir() string {
	return filepath.Join(a.UserDir(), "stores")
}

// UserAuthDir returns the auth directory scoped per user (~/.rift/users/<user_id>/auth).
func (a *App) UserAuthDir() string {
	return filepath.Join(a.UserDir(), "auth")
}

// GamesDir returns the games directory (defaults to ~/.rift/games).
func (a *App) GamesDir() string {
	return filepath.Join(a.RiftDir(), "games")
}

// EnsureBaseDirectories ensures that all critical RIFT directories exist.
// If the user deletes ~/.rift or any subfolder in Finder, this auto-reconstructs them.
func (a *App) EnsureBaseDirectories() {
	dirs := []string{
		a.RiftDir(),
		a.GamesDir(),
		a.EnginesDir(),
		a.AuthDir(),
		a.UserDir(),
		a.StoresDir(),
		a.UserAuthDir(),
		filepath.Join(a.StoresDir(), "steam"),
		filepath.Join(a.EnginesDir(), "dxvk"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			a.logError("[Self-Healing] Failed to ensure directory %s: %v", dir, err)
		}
	}
	a.migrateLegacyUserData()
}

func (a *App) migrateLegacyUserData() {
	legacyUserID := "87e8c939-e12c-4430-aef4-582ecb39ad18"
	legacyUserDir := filepath.Join(a.RiftDir(), "users", legacyUserID)
	if _, err := os.Stat(legacyUserDir); err == nil {
		return // already migrated
	}

	legacyPlaytime := filepath.Join(a.RiftDir(), "data", "playtime.json")
	legacyStoresSteam := filepath.Join(a.RiftDir(), "stores", "steam")
	legacyAuthEpic := filepath.Join(a.RiftDir(), "auth", "epic")

	hasLegacy := false
	if _, err := os.Stat(legacyPlaytime); err == nil {
		hasLegacy = true
	}
	if _, err := os.Stat(legacyStoresSteam); err == nil {
		hasLegacy = true
	}

	if !hasLegacy {
		return
	}

	_ = os.MkdirAll(filepath.Join(legacyUserDir, "stores"), 0755)
	_ = os.MkdirAll(filepath.Join(legacyUserDir, "auth"), 0755)

	if data, err := os.ReadFile(legacyPlaytime); err == nil {
		_ = os.WriteFile(filepath.Join(legacyUserDir, "playtime.json"), data, 0644)
	}
	if _, err := os.Stat(legacyStoresSteam); err == nil {
		_ = os.Rename(legacyStoresSteam, filepath.Join(legacyUserDir, "stores", "steam"))
	}
	if _, err := os.Stat(legacyAuthEpic); err == nil {
		_ = os.Rename(legacyAuthEpic, filepath.Join(legacyUserDir, "auth", "epic"))
	}
}

// emitEvent safely emits a Wails event if app context is initialized.
func (a *App) emitEvent(eventName string, optionalData ...interface{}) {
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, eventName, optionalData...)
	}
}

// runtimeConfigPath returns the per-game runtime config file path.
func (a *App) runtimeConfigPath(gameID string) string {
	return filepath.Join(a.ResolveGameDir(gameID), "game_runtime.json")
}

// Convenience logging helpers that use the session logger (fall back to fmt).
func (a *App) logInfo(format string, args ...interface{}) {
	if a.Log != nil {
		a.Log.Info(format, args...)
	} else {
		fmt.Printf(format, args...)

		if !strings.HasSuffix(format, "\n") {
			fmt.Println()
		}
	}
}

func (a *App) logWarn(format string, args ...interface{}) {
	if a.Log != nil {
		a.Log.Warn(format, args...)
	} else {
		fmt.Printf("WARN: "+format, args...)
		if !strings.HasSuffix(format, "\n") {
			fmt.Println()
		}
	}
}

func (a *App) logError(format string, args ...interface{}) {
	if a.Log != nil {
		a.Log.Error(format, args...)
	} else {
		fmt.Printf("ERROR: "+format, args...)
		if !strings.HasSuffix(format, "\n") {
			fmt.Println()
		}
	}
}

func loadEnvVars() (string, string) {
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_ANON_KEY")

	if url != "" && key != "" {
		return url, key
	}

	home, _ := os.UserHomeDir()
	envPath := filepath.Join(home, ".rift", ".env")
	if data, err := os.ReadFile(envPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "SUPABASE_URL=") {
				url = strings.TrimPrefix(line, "SUPABASE_URL=")
				url = strings.Trim(url, "\"'")
			}
			if strings.HasPrefix(line, "SUPABASE_ANON_KEY=") {
				key = strings.TrimPrefix(line, "SUPABASE_ANON_KEY=")
				key = strings.Trim(key, "\"'")
			}
		}
	}
	if url == "" {
		url = cloud.DefaultSupabaseURL
	}
	if key == "" {
		key = cloud.DefaultSupabaseAnonKey
	}
	return url, key
}

func raiseFileLimit() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err == nil {
		max := uint64(65536)
		if rLimit.Max < max {
			max = rLimit.Max
		}
		if rLimit.Cur < max {
			rLimit.Cur = max
			_ = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		}
	}
}

// NewApp creates a new App struct and loads config options from disk.
func NewApp() *App {
	raiseFileLimit()
	home, _ := os.UserHomeDir()
	baseDir := filepath.Join(home, ".rift")
	sbURL, sbKey := loadEnvVars()

	app := &App{
		EngineManager: engine.NewEngineManager(),
		RiftBaseDir:   baseDir,
		GlobalAuthDir: filepath.Join(baseDir, "auth"),
		dls:           NewDownloadManager(),
		Log:           InitLogger(),
		Supabase:      cloud.NewSupabaseClient(sbURL, sbKey, filepath.Join(baseDir, "auth")),
		Scraper:       scraper.NewBackgroundScraper(),
	}

	app.Log.Info("RIFT app initializing")
	app.EnsureBaseDirectories()
	app.loadConfig()
	app.Log.Info("GamesDir: %s", app.GetGamesDir())
	return app
}

// Startup is called when the app starts. The context is saved.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.EnsureBaseDirectories()
	a.Log.Info("RIFT app startup complete")
	sysmon.StartTelemetry(ctx)
	if a.Scraper != nil {
		a.Scraper.Start(ctx)
	}
	go a.FlushPendingTelemetry()
	go a.SyncCloudPlaytimeOnLogin()
}

// Shutdown cleans up the session logger and background workers on app exit.
func (a *App) Shutdown(ctx context.Context) {
	a.StopActiveDownloadsOnExit()
	if a.Scraper != nil {
		a.Scraper.Stop()
	}
	if a.Log != nil {
		a.Log.Close()
	}
}

// BeforeClose handles window close events when the user clicks the red (x) button.
// It stops active download workers cleanly (preserving partial download chunks on disk
// so the user can resume them later) and allows the application to quit completely.
func (a *App) BeforeClose(ctx context.Context) (prevent bool) {
	a.logInfo("[App] Window close requested (red cross clicked). Stopping active workers and quitting.")
	a.StopActiveDownloadsOnExit()
	return false // Always allow the app to close and quit!
}

// RestoreWindow un-minimises, un-hides, and brings the main window to the front.
func (a *App) RestoreWindow() {
	if a.ctx != nil {
		wailsRuntime.Show(a.ctx)
		wailsRuntime.WindowShow(a.ctx)
		wailsRuntime.WindowUnminimise(a.ctx)
	}
}

// UninstallGame completely removes a game's capsule (prefix, files, configuration) from disk.
func (a *App) LogMessage(level string, message string) {
	switch level {
	case "info":
		a.logInfo("%s", message)
	case "warn":
		a.logWarn("%s", message)
	case "error":
		a.logError("%s", message)
	default:
		a.logInfo("%s", message)
	}
}
