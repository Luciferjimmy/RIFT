// Package optimizer implements a multi-stage game optimization pipeline that
// combines local databases, system profiling, deterministic rules, AI reasoning,
// and optional headless-browser enrichment to produce the best launch configuration
// for running Windows games on macOS via RIFT.
package optimizer

import "time"

// ---------------------------------------------------------------------------
// Identity
// ---------------------------------------------------------------------------

// GameIdentity holds every known identifier for a game across storefronts.
type GameIdentity struct {
	Name        string // canonical display name
	SteamAppID  string // preferred key for DB lookups
	GOGID       string
	EpicID      string
	LutrisSlugs []string // all matching Lutris slugs
	Aliases     []string // alternative names for fuzzy matching
}

// ---------------------------------------------------------------------------
// Database evidence (Stage 2 output)
// ---------------------------------------------------------------------------

// PCGWGame is a row from the PCGamingWiki database.
type PCGWGame struct {
	PageName    string
	Developer   string
	Publisher   string
	Engine      string // "Unity", "Unreal Engine 4", "Godot", etc.
	SteamAppID  string
	GOGID       string
	ReleaseDate string
}

// LutrisScript is a row from the Lutris install-scripts table.
type LutrisScript struct {
	GameSlug   string
	GameName   string
	Version    string // e.g. "Standard", "GOG", "Epic"
	Runner     string // "wine", "wine-proton", "steam", "linux"
	WinPrefix  string // relative prefix path hint
	Executable string
	RawScript  string // full YAML installer script
}

// ProtonDBEntry is a single user report from the ProtonDB dataset.
type ProtonDBEntry struct {
	GameName      string
	AppID         string
	ProtonVersion string
	Rating        string // "Platinum", "Gold", "Silver", "Bronze", "Borked"
	UserSpecs     string
	LaunchArgs    string
	EnvVars       string // space-separated KEY=VALUE pairs
	Winetricks    string // space-separated component names
	NotesRaw      string
}

// AGWEntry is a single report from the Apple Gaming Wiki dataset.
type AGWEntry struct {
	GameName  string
	NativeMac bool
	Rating    string // "Native", "Perfect", "Playable", "Unsupported"
	Link      string
}

// AggregatedGameData is the merged output of Stage 2. Every field is
// populated from its respective source; missing data leaves the field zero.
type AggregatedGameData struct {
	PCGamingWiki *PCGWGame
	LutrisBest   *LutrisScript  // highest-confidence match
	ProtonDBBest *ProtonDBEntry // highest rating
	AGW          *AGWEntry
	HasNativeMac bool // shortcut: true if AGW confirms native port
}

// ---------------------------------------------------------------------------
// System profile (Stage 3 output)
// ---------------------------------------------------------------------------

// SystemProfile describes the host macOS machine and installed engines.
type SystemProfile struct {
	MacOSVersion   string // "26.5.1"
	AppleSilicon   bool
	MetalGPUFamily int // 3 = Metal3, 4 = Metal4, etc.; 0 = unknown
	RAMGB          int
	Engines        EngineStates
	Rosetta2       bool
}

// EngineStates reports which translation engines are available.
type EngineStates struct {
	Wine bool
	DXVK bool
	GPTK bool // D3DMetal framework present
}

// ---------------------------------------------------------------------------
// Configuration synthesis (Stage 7 output)
// ---------------------------------------------------------------------------

// GameConfig is the per-game launch configuration written to
// game_runtime.json. Every field is populated by the pipeline.
type GameConfig struct {
	GameID         string            `json:"game_id"`
	Engine         string            `json:"engine"`          // "d3dmetal" | "dxvk" | "native" | "wine-only"
	WineVersion    string            `json:"wine_version"`    // "latest" | specific version string
	WindowsVersion string            `json:"windows_version"` // "win10" | "win8" | "win7" | "winxp"
	DLLOverrides   map[string]string `json:"dll_overrides"`   // "d3d11": "n,b"
	EnvVars        map[string]string `json:"env_vars"`        // "WINEMSYNC": "1", "DYLD_INSERT_LIBRARIES": "..."
	LaunchArgs     string            `json:"launch_args"`
	Winetricks     []string          `json:"winetricks"`    // ["vcrun2022", "dotnet48"]
	CompiledDLLs   []string          `json:"compiled_dlls"` // DLLs to copy from master cache
	Confidence     float64           `json:"confidence"`    // 0.0 – 1.0
	Source         string            `json:"source"`        // "rules" | "ai" | "hybrid" | "browser"
	PipelineLog    []string          `json:"pipeline_log"`  // explainability trail
}

// ---------------------------------------------------------------------------
// Pipeline result
// ---------------------------------------------------------------------------

// Result wraps the final configuration together with any errors that occurred
// during non-fatal stages (AI failure, browser timeout, etc.).
type Result struct {
	Config      GameConfig
	NonFatalErr []error
}

// ---------------------------------------------------------------------------
// Cache
// ---------------------------------------------------------------------------

// CachedConfig is stored on disk so the pipeline can be skipped when inputs
// have not changed.
type CachedConfig struct {
	Config     GameConfig `json:"config"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	SystemHash string     `json:"system_hash"` // hash of system profile at creation
}

// ---------------------------------------------------------------------------
// Confidence thresholds (shared constants)
// ---------------------------------------------------------------------------

const (
	ConfThresholdRulesOnly = 0.7 // ≥ : skip AI, go straight to synthesis
	ConfThresholdAIOnly    = 0.4 // ≥ : run AI, skip browser
	// below 0.4: run AI + browser enrichment

	confWeightEngineKnown  = 0.30
	confWeightProtonDBGold = 0.30
	confWeightLutrisScript = 0.20
	confWeightConsistency  = 0.20
)

// ---------------------------------------------------------------------------
// Errors
// ---------------------------------------------------------------------------

// NonFatalError is an error that does not abort the pipeline; it is collected
// and attached to the result.
type NonFatalError struct {
	Stage string
	Err   error
}

func (e *NonFatalError) Error() string {
	return e.Stage + ": " + e.Err.Error()
}
