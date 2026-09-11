package optimizer

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// DBProvider opens and caches connections to the local game databases.
// Each database is opened once and reused; callers must not close the
// returned handles (they are managed by DBProvider).
type DBProvider struct {
	mu    sync.Mutex
	conns map[string]*sql.DB

	dataDir string // resolved once from resolveDataDir
}

// NewDBProvider creates a DBProvider that looks for database files under dataDir.
func NewDBProvider(dataDir string) *DBProvider {
	return &DBProvider{
		conns:   make(map[string]*sql.DB),
		dataDir: dataDir,
	}
}

// ResolveDataDir finds the data directory by checking (in order):
//  1. Executable-relative "data/"
//  2. Working-directory "data/"
//  3. ~/.rift/data/
func ResolveDataDir() string {
	if execPath, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(execPath), "data")
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			return candidate
		}
	}
	for _, dir := range []string{"data"} {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(dir)
			return abs
		}
	}
	home, _ := os.UserHomeDir()
	candidate := filepath.Join(home, ".rift", "data")
	if info, err := os.Stat(candidate); err == nil && info.IsDir() {
		return candidate
	}
	return "data"
}

// open opens (or returns a cached) SQLite connection for the given file name
// inside dataDir.
func (p *DBProvider) open(name string) (*sql.DB, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if db, ok := p.conns[name]; ok {
		return db, nil
	}
	path := filepath.Join(p.dataDir, name)
	db, err := sql.Open("sqlite3", path+"?mode=ro&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	db.SetMaxOpenConns(1) // SQLite WAL with 1 writer is safest for concurrent reads
	p.conns[name] = db
	return db, nil
}

// ---------------------------------------------------------------------------
// PCGamingWiki
// ---------------------------------------------------------------------------

// QueryPCGamingWiki looks up a game by Steam AppID or page name.
func (p *DBProvider) QueryPCGamingWiki(steamID, name string) (*PCGWGame, error) {
	db, err := p.open("pcgamingwiki/pcgamingwiki.db")
	if err != nil {
		return nil, err
	}

	var row PCGWGame
	q := `SELECT page_name, developer, publisher, engine, steam_appid, gog_id, release_date
		  FROM pcgw_games WHERE steam_appid = ? OR page_name = ? LIMIT 1`
	err = db.QueryRow(q, steamID, name).Scan(
		&row.PageName, &row.Developer, &row.Publisher,
		&row.Engine, &row.SteamAppID, &row.GOGID, &row.ReleaseDate,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pcgw query: %w", err)
	}
	return &row, nil
}

// ---------------------------------------------------------------------------
// Lutris
// ---------------------------------------------------------------------------

// QueryLutrisGames looks for matching Lutris games by name. Returns the best
// match and its associated install script (if any).
func (p *DBProvider) QueryLutrisGames(name string) (games []struct {
	Slug          string
	Name          string
	HasInstallers bool
}, _ error) {
	db, err := p.open("lutris/lutris.db")
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(
		`SELECT slug, name, has_installers FROM lutris_games WHERE name LIKE ? LIMIT 5`,
		"%"+name+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("lutris games query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var g struct {
			Slug          string
			Name          string
			HasInstallers bool
		}
		if err := rows.Scan(&g.Slug, &g.Name, &g.HasInstallers); err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, rows.Err()
}

// QueryLutrisScript returns the first install script for the given game slug.
func (p *DBProvider) QueryLutrisScript(slug string) (*LutrisScript, error) {
	db, err := p.open("lutris/lutris.db")
	if err != nil {
		return nil, err
	}

	var s LutrisScript
	q := `SELECT game_slug, game_name, script_slug, version, runner, wine_prefix, executable, raw_script
		  FROM lutris_scripts WHERE game_slug = ? LIMIT 1`
	err = db.QueryRow(q, slug).Scan(
		&s.GameSlug, &s.GameName, &s.Version,
		&s.Runner, &s.WinPrefix, &s.Executable, &s.RawScript,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("lutris script query: %w", err)
	}
	// script_slug is never null in the schema but we keep the field name aligned
	// with the struct; we re-assign to Version for clarity.
	if s.Version == "" {
		s.Version = s.GameSlug
	}
	return &s, nil
}

// ---------------------------------------------------------------------------
// ProtonDB
// ---------------------------------------------------------------------------

// QueryProtonDB returns the highest-rated fix entries for a Steam AppID.
func (p *DBProvider) QueryProtonDB(steamID string) ([]ProtonDBEntry, error) {
	db, err := p.open("protondb/rift.db")
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(
		`SELECT game_name, app_id, proton_version, rating,
		        COALESCE(launch_args,''), COALESCE(env_vars,''),
		        COALESCE(winetricks,''), COALESCE(notes_raw,'')
		 FROM game_fixes WHERE app_id = ?
		 ORDER BY CASE rating
		   WHEN 'Platinum' THEN 1 WHEN 'Gold' THEN 2
		   WHEN 'Silver'  THEN 3 WHEN 'Bronze' THEN 4
		   ELSE 5 END
		 LIMIT 5`, steamID,
	)
	if err != nil {
		return nil, fmt.Errorf("protondb query: %w", err)
	}
	defer rows.Close()

	var entries []ProtonDBEntry
	for rows.Next() {
		var e ProtonDBEntry
		if err := rows.Scan(&e.GameName, &e.AppID, &e.ProtonVersion,
			&e.Rating, &e.LaunchArgs, &e.EnvVars,
			&e.Winetricks, &e.NotesRaw); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// ---------------------------------------------------------------------------
// Apple Gaming Wiki
// ---------------------------------------------------------------------------

// QueryAGW loads the AGW reports dataset and looks for the given game name.
func (p *DBProvider) QueryAGW(name string) (*AGWEntry, error) {
	path := filepath.Join(p.dataDir, "applegamingwiki", "agw_reports.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil // silently skip — dataset is optional
	}

	var raw []struct {
		Title  string `json:"title"`
		Status string `json:"status"`
		Link   string `json:"link"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, nil
	}

	// Simple substring match; in production we would want token-aware matching.
	nameLower := toLower(name)
	for _, r := range raw {
		if containsFold(r.Title, nameLower) {
			native := r.Status == "Native" || r.Status == "Perfect"
			return &AGWEntry{
				GameName:  r.Title,
				NativeMac: native,
				Rating:    r.Status,
				Link:      r.Link,
			}, nil
		}
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// Bottles
// ---------------------------------------------------------------------------

// BottlesDependency describes a single Bottles dependency (vcrun, dotnet, etc.).
type BottlesDependency struct {
	Name          string
	Description   string
	Category      string
	DownloadURL   string
	InstallAction string
}

// QueryBottlesDependencies returns all known Bottles dependencies.
func (p *DBProvider) QueryBottlesDependencies() ([]BottlesDependency, error) {
	db, err := p.open("bottles/bottles.db")
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(
		`SELECT dep_name, COALESCE(description,''), COALESCE(category,''),
		        COALESCE(download_url,''), COALESCE(install_action,'')
		 FROM bottles_dependencies`)
	if err != nil {
		return nil, fmt.Errorf("bottles query: %w", err)
	}
	defer rows.Close()

	var deps []BottlesDependency
	for rows.Next() {
		var d BottlesDependency
		if err := rows.Scan(&d.Name, &d.Description, &d.Category, &d.DownloadURL, &d.InstallAction); err != nil {
			return nil, err
		}
		deps = append(deps, d)
	}
	return deps, rows.Err()
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func toLower(s string) string {
	buf := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		buf[i] = c
	}
	return string(buf)
}

func containsFold(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return len(s) >= len(substr) && containsString(s, substr)
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && indexString(s, substr) >= 0
}

func indexString(s, substr string) int {
	n := len(substr)
	if n == 0 {
		return 0
	}
	for i := 0; i <= len(s)-n; i++ {
		if s[i:i+n] == substr {
			return i
		}
	}
	return -1
}
