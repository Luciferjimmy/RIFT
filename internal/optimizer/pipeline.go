// Package optimizer implements a multi-stage game optimization pipeline that
// combines local databases, system profiling, deterministic rules, AI reasoning,
// and optional headless-browser enrichment to produce the best launch configuration
// for running Windows games on macOS via RIFT.
//
// Pipeline stages:
//
//	Stage 1 — Identity Resolution:   map any game ID to a canonical identity
//	Stage 2 — Local DB Aggregation:  parallel queries to all local databases
//	Stage 3 — System Profile:        hardware + OS + engine availability
//	Stage 4 — Rules Engine:          deterministic decision tree (80%+ coverage)
//	Stage 5 — AI Reasoning:          optional LLM call for low-confidence gaps
//	Stage 6 — Browser Enrichment:    optional web scrape for unknown games
//	Stage 7 — Config Synthesis:      merge all outputs into final game_runtime.json
//	Stage 8 — Cache & Persist:       store result for fast replay
package optimizer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Pipeline is the top-level orchestrator. It runs the eight stages in order,
// with early-exit fast paths where possible.
type Pipeline struct {
	db         *DBProvider
	resolver   *IdentityResolver
	aggregator *Aggregator
	profiler   *SystemProfiler
	rules      *RulesEngine
	ai         *AIReasoner
	browser    *BrowserEnricher
	synth      *Synthesizer
	cache      *Cache
}

// PipelineOption configures optional behaviour.
type PipelineOption func(*Pipeline)

// WithAIKey sets the Groq API key for the AI reasoning stage.
// If not set, AI stage is skipped entirely.
func WithAIKey(key string) PipelineOption {
	return func(p *Pipeline) {
		if key != "" {
			p.ai = NewAIReasoner(key)
		}
	}
}

// NewPipeline creates a fully wired pipeline. If no Groq API key is provided
// via WithAIKey, the AI stage is skipped.
func NewPipeline(opts ...PipelineOption) *Pipeline {
	dataDir := ResolveDataDir()
	db := NewDBProvider(dataDir)

	p := &Pipeline{
		db:         db,
		resolver:   NewIdentityResolver(db),
		aggregator: NewAggregator(db),
		profiler:   NewSystemProfiler(),
		rules:      NewRulesEngine(),
		synth:      NewSynthesizer(),
		browser:    NewBrowserEnricher(),
		cache:      NewCache(),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// Run executes the full pipeline and returns the optimal GameConfig for the
// given game. A system profile is collected fresh on every call because engine
// availability changes over time (user may install/uninstall).
//
// The pipeline has three tiers:
//   - Fast path (±500ms):  DB + rules only (confidence ≥ 0.7)
//   - AI path (±3-5s):     adds AI stage when confidence ∈ [0.4, 0.7)
//   - Browser path (±10s): adds headless scrape when confidence < 0.4
func (p *Pipeline) Run(ctx context.Context, gameID, steamID, gameName string) (*Result, error) {
	result := &Result{
		Config: GameConfig{
			GameID:         gameID,
			Engine:         "d3dmetal",
			WindowsVersion: "win10",
			DLLOverrides:   make(map[string]string),
			EnvVars:        make(map[string]string),
		},
	}

	// ---- System profile (Stage 3, needed early for cache key) ----
	sys := p.profiler.Profile()
	sysHash := sys.SystemHash()

	// ---- Cache lookup ----
	if cached, ok := p.cache.Get(gameID, sysHash); ok {
		result.Config = *cached
		result.Config.PipelineLog = append(result.Config.PipelineLog,
			"CACHE HIT: returning cached config (system unchanged)")
		return result, nil
	}

	log := func(s string) { result.Config.PipelineLog = append(result.Config.PipelineLog, s) }

	// ---- Stage 1: Identity Resolution ----
	log("Stage 1: resolving game identity")
	ident, err := p.resolver.Resolve(steamID, gameName)
	if err != nil {
		return nil, fmt.Errorf("identity resolution: %w", err)
	}
	log(fmt.Sprintf("Identity: name=%q steamID=%s", ident.Name, ident.SteamAppID))

	// ---- Stage 2: Local DB Aggregation ----
	log("Stage 2: aggregating local databases (parallel)")
	data := p.aggregator.Aggregate(ctx, ident)
	if data.PCGamingWiki != nil {
		log(fmt.Sprintf("PCGamingWiki: engine=%s", data.PCGamingWiki.Engine))
	}
	if data.ProtonDBBest != nil {
		log(fmt.Sprintf("ProtonDB: rating=%s", data.ProtonDBBest.Rating))
	}
	if data.LutrisBest != nil {
		log(fmt.Sprintf("Lutris: runner=%s", data.LutrisBest.Runner))
	}
	if data.HasNativeMac {
		log("AGW: native Mac port confirmed")
	}

	// ---- Stage 4: Rules Engine ----
	log("Stage 4: running deterministic rules engine")
	rulesResult := p.rules.Evaluate(data, sys)
	log(fmt.Sprintf("Rules confidence: %.2f, engine=%s", rulesResult.Confidence, rulesResult.Config.Engine))

	// ---- Stage 5: AI Reasoning (if confidence too low) ----
	var aiResult *AIFillResult
	if rulesResult.Confidence < ConfThresholdRulesOnly && p.ai != nil {
		log(fmt.Sprintf("Stage 5: confidence %.2f < %.2f, invoking AI", rulesResult.Confidence, ConfThresholdRulesOnly))
		aiCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		aiResult = p.ai.Fill(aiCtx, ident.Name, extractEngineName(data), ident.SteamAppID, data, sys)
		cancel()

		if aiResult != nil {
			log(fmt.Sprintf("AI confidence: %.2f", aiResult.Confidence))
		} else {
			log("AI returned nil (API error or skipped) — using rules-only")
		}
	}

	// ---- Stage 6: Browser Enrichment (if very low confidence) ----
	var browserResult *EnrichResult
	if rulesResult.Confidence < ConfThresholdAIOnly {
		log(fmt.Sprintf("Stage 6: confidence %.2f < %.2f, invoking browser enrichment", rulesResult.Confidence, ConfThresholdAIOnly))
		browserCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		browserResult = p.browser.Enrich(browserCtx, ident)
		cancel()
		if browserResult != nil && browserResult.ConfidenceDelta > 0 {
			log(fmt.Sprintf("Browser confidence delta: +%.2f", browserResult.ConfidenceDelta))
		}
	}

	// ---- Stage 7: Config Synthesis ----
	log("Stage 7: synthesizing final configuration")
	result.Config = *p.synth.Merge(rulesResult, aiResult, browserResult)
	result.Config.GameID = gameID

	// ---- Stage 8: Cache ----
	log("Stage 8: caching result")
	if err := p.cache.Set(gameID, sysHash, &result.Config); err != nil {
		log(fmt.Sprintf("Cache write failed (non-fatal): %v", err))
		result.NonFatalErr = append(result.NonFatalErr, err)
	}

	return result, nil
}

// Run is a convenience wrapper that performs the full pipeline with the given
// game name and optional Steam AppID.
func Run(ctx context.Context, gameID, steamID, gameName, groqAPIKey string) (*Result, error) {
	p := NewPipeline(WithAIKey(groqAPIKey))
	return p.Run(ctx, gameID, steamID, gameName)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func extractEngineName(data *AggregatedGameData) string {
	if data.PCGamingWiki != nil {
		return data.PCGamingWiki.Engine
	}
	return ""
}

// ---------------------------------------------------------------------------
// Integration: writing the config to the game directory
// ---------------------------------------------------------------------------

// WriteConfig persists the GameConfig to the game's directory as
// game_runtime.json, matching the format expected by existing code in
// cmd/rift/bindings.go.
func WriteConfig(gamesDir, gameID string, cfg *GameConfig) error {
	dir := filepath.Join(gamesDir, gameID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create game dir: %w", err)
	}

	type runtimeV2 struct {
		Engine         string            `json:"engine"`
		WineVersion    string            `json:"wine_version"`
		WindowsVersion string            `json:"windows_version"`
		LaunchArgs     []string          `json:"launch_args"`
		DLLOverrides   map[string]string `json:"dll_overrides"`
		EnvVars        map[string]string `json:"env_vars"`
		Winetricks     []string          `json:"winetricks"`
		CompiledDLLs   []string          `json:"compiled_dlls"`
		Confidence     float64           `json:"confidence"`
		Source         string            `json:"source"`
	}

	rv := runtimeV2{
		Engine:         cfg.Engine,
		WineVersion:    cfg.WineVersion,
		WindowsVersion: cfg.WindowsVersion,
		LaunchArgs:     splitArgs(cfg.LaunchArgs),
		DLLOverrides:   cfg.DLLOverrides,
		EnvVars:        cfg.EnvVars,
		Winetricks:     cfg.Winetricks,
		CompiledDLLs:   cfg.CompiledDLLs,
		Confidence:     cfg.Confidence,
		Source:         cfg.Source,
	}

	data, err := json.MarshalIndent(rv, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal runtime config: %w", err)
	}

	outPath := filepath.Join(dir, "game_runtime.json")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return fmt.Errorf("write runtime config: %w", err)
	}
	return nil
}

func splitArgs(s string) []string {
	if s == "" || s == "SKIP" {
		return nil
	}
	// Simple whitespace splitting — fails for args containing spaces, but
	// that's extremely rare in game launch arguments.
	var args []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ' ' {
			if i > start {
				args = append(args, s[start:i])
			}
			start = i + 1
		}
	}
	return args
}
