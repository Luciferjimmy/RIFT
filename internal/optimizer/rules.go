package optimizer

import (
	"fmt"
	"strings"
)

// RulesEngine implements the deterministic decision tree that handles 80%+
// of games without any AI call. It evaluates the aggregated data against the
// system profile and produces a preliminary configuration.
type RulesEngine struct{}

// NewRulesEngine creates a rules engine.
func NewRulesEngine() *RulesEngine {
	return &RulesEngine{}
}

// EvaluateResult is the output of the rules engine.
type EvaluateResult struct {
	Config     GameConfig
	Confidence float64
	Log        []string // explainability entries
}

// Evaluate runs the decision tree. The returned Config may be partially filled;
// confidence < 1.0 means some fields should be filled by later stages (AI).
//
// Decision priority (top wins):
//  1. Native Mac → skip Wine entirely
//  2. Game engine known → pick best GPU translation layer
//  3. ProtonDB rating → apply known-good tweaks
//  4. Lutris script → extract runner / env
//  5. macOS version / hardware → fallback defaults
func (r *RulesEngine) Evaluate(data *AggregatedGameData, sys *SystemProfile) *EvaluateResult {
	res := &EvaluateResult{
		Config: GameConfig{
			Engine:         "d3dmetal", // safest default on modern macOS
			WineVersion:    "latest",
			WindowsVersion: "win10",
			DLLOverrides:   make(map[string]string),
			EnvVars:        make(map[string]string),
		},
	}
	log := func(msg string) { res.Log = append(res.Log, msg) }

	// ---- Rule 1: Native Mac ----
	if data.HasNativeMac {
		res.Config.Engine = "native"
		res.Confidence = 1.0
		log("Native Mac port confirmed via Apple Gaming Wiki → engine=native, skip Wine")
		return res
	}

	// ---- Rule 2: Engine-based GPU layer selection ----
	engineName := ""
	if data.PCGamingWiki != nil {
		engineName = data.PCGamingWiki.Engine
	}
	applyEngineRule(engineName, &res.Config, log)

	// ---- Rule 3: ProtonDB overrides ----
	if data.ProtonDBBest != nil {
		pb := data.ProtonDBBest
		log(fmt.Sprintf("ProtonDB best report: rating=%s, version=%s", pb.Rating, pb.ProtonVersion))
		applyProtonDBRule(pb, &res.Config, log)
	}

	// ---- Rule 4: Lutris script hints ----
	if data.LutrisBest != nil {
		ls := data.LutrisBest
		log(fmt.Sprintf("Lutris script found: runner=%s, version=%s", ls.Runner, ls.Version))
		applyLutrisRule(ls, &res.Config, log)
	}

	// ---- Rule 5: macOS version + hardware fallback ----
	applySystemRule(sys, &res.Config, log)

	// ---- Confidence calculation ----
	res.Confidence = calcConfidence(data, sys, engineName)

	// ---- Enforce DXVK only on Intel Macs ----
	if !sys.AppleSilicon && res.Config.Engine != "native" {
		res.Config.Engine = "dxvk"
		log("Intel Mac detected → forcing DXVK (D3DMetal requires Apple Silicon)")
	}

	log(fmt.Sprintf("Confidence=%.2f, final engine=%s", res.Confidence, res.Config.Engine))
	return res
}

// ---------------------------------------------------------------------------
// Engine-based rule
// ---------------------------------------------------------------------------

func applyEngineRule(engine string, cfg *GameConfig, log func(string)) {
	lower := strings.ToLower(strings.TrimSpace(engine))
	switch {
	case lower == "":
		// Unknown engine — we keep the D3DMetal default but lower confidence.
		log("Engine unknown from PCGamingWiki → D3DMetal default (low confidence)")

	case strings.Contains(lower, "unity"):
		cfg.Engine = "d3dmetal"
		log("Engine=Unity → D3DMetal preferred (mature Metal translation)")

	case strings.Contains(lower, "unreal engine 4") || strings.Contains(lower, "unreal engine 5") ||
		strings.Contains(lower, "ue4") || strings.Contains(lower, "ue5"):
		cfg.Engine = "d3dmetal"
		log("Engine=Unreal Engine 4/5 → D3DMetal preferred (native Metal renderer available)")

	case strings.Contains(lower, "unreal engine 3") || strings.Contains(lower, "ue3"):
		cfg.Engine = "dxvk"
		log("Engine=Unreal Engine 3 → DXVK preferred (no Metal path in UE3)")

	case strings.Contains(lower, "godot"):
		cfg.Engine = "d3dmetal"
		log("Engine=Godot → D3DMetal preferred")

	case strings.Contains(lower, "source"):
		cfg.Engine = "d3dmetal"
		log("Engine=Source → D3DMetal preferred")

	case strings.Contains(lower, "rage") || strings.Contains(lower, "rockstar"):
		cfg.Engine = "dxvk"
		log("Engine=RAGE (Rockstar) → DXVK recommended for RAGE engine")

	default:
		log(fmt.Sprintf("Engine=%s → using D3DMetal default", engine))
	}
}

// ---------------------------------------------------------------------------
// ProtonDB rule
// ---------------------------------------------------------------------------

func applyProtonDBRule(pb *ProtonDBEntry, cfg *GameConfig, log func(string)) {
	switch pb.Rating {
	case "Platinum", "Gold":
		log("ProtonDB rating is Platinum/Gold → high confidence in Wine path")
	case "Silver", "Bronze":
		log("ProtonDB rating is Silver/Bronze → some tweaks may be needed")
	case "Borked":
		// Try D3DMetal instead of DXVK as a different translation path.
		if cfg.Engine != "native" {
			cfg.Engine = "d3dmetal"
			log("ProtonDB=Borked → switching to D3DMetal as alternative path")
		}
	}

	// Apply known-working tweaks from the report.
	if pb.LaunchArgs != "" {
		cfg.LaunchArgs = pb.LaunchArgs
		log(fmt.Sprintf("ProtonDB launch args: %s", pb.LaunchArgs))
	}
	if pb.EnvVars != "" {
		for _, pair := range strings.Fields(pb.EnvVars) {
			k, v, ok := strings.Cut(pair, "=")
			if ok {
				cfg.EnvVars[k] = v
			}
		}
		log(fmt.Sprintf("ProtonDB env vars: %s", pb.EnvVars))
	}
	if pb.Winetricks != "" {
		for _, wt := range strings.Fields(pb.Winetricks) {
			wt = strings.TrimSpace(wt)
			if wt != "" {
				cfg.Winetricks = append(cfg.Winetricks, wt)
			}
		}
		log(fmt.Sprintf("ProtonDB winetricks: %s", strings.Join(cfg.Winetricks, ", ")))
	}
}

// ---------------------------------------------------------------------------
// Lutris rule
// ---------------------------------------------------------------------------

func applyLutrisRule(ls *LutrisScript, cfg *GameConfig, log func(string)) {
	// Runner selection.
	runner := strings.ToLower(ls.Runner)
	if strings.Contains(runner, "wine-proton") || strings.Contains(runner, "proton") {
		log("Lutris uses Proton runner → ensuring DXVK/D3DMetal compatibility")
	}
	if strings.Contains(runner, "linux") || strings.Contains(runner, "native") {
		// This is a native Linux script — not directly applicable, but hints at
		// game engine and capabilities.
		log("Lutris runner is native Linux — useful for engine hints only")
	}

	// Parse raw script for DLL overrides and environment variables.
	if ls.RawScript != "" {
		extractLutrisOverrides(ls.RawScript, cfg, log)
	}
}

func extractLutrisOverrides(raw string, cfg *GameConfig, log func(string)) {
	// Lutris YAML scripts are not perfectly structured, but common patterns
	// include "wine:" → "dll_overrides:" or "environment:" → variables.
	// We perform simple line-based extraction here.
	lines := strings.Split(raw, "\n")
	inOverrides := false
	inEnv := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trimmed, "dll_overrides"):
			inOverrides = true
			inEnv = false
		case strings.HasPrefix(trimmed, "environment"):
			inEnv = true
			inOverrides = false
		case strings.HasPrefix(trimmed, "wine:"):
			inOverrides = false
			inEnv = false
		case inOverrides && strings.Contains(trimmed, ":"):
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				dll := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				if dll != "" && val != "" {
					cfg.DLLOverrides[dll] = val
				}
			}
		case inEnv && strings.Contains(trimmed, "="):
			k, v, ok := strings.Cut(trimmed, "=")
			if ok {
				cfg.EnvVars[strings.TrimSpace(k)] = strings.TrimSpace(v)
			}
		case inEnv && strings.Contains(trimmed, ":"):
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				cfg.EnvVars[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}
}

// ---------------------------------------------------------------------------
// System-based rule
// ---------------------------------------------------------------------------

func applySystemRule(sys *SystemProfile, cfg *GameConfig, log func(string)) {
	// Windows version.
	major := majorVersion(sys.MacOSVersion)
	if major >= 20 {
		log("macOS 20+ → D3DMetal v4 is the best translation layer")
	}

	// RAM constraint.
	if sys.RAMGB > 0 && sys.RAMGB < 8 {
		log(fmt.Sprintf("Only %dGB RAM — DXVK preferred (lighter than D3DMetal)", sys.RAMGB))
		if cfg.Engine != "native" {
			cfg.Engine = "dxvk"
		}
	}

	// Engine availability overrides.
	if !sys.Engines.GPTK && cfg.Engine == "d3dmetal" {
		if sys.Engines.DXVK {
			cfg.Engine = "dxvk"
			log("GPTK not installed, DXVK available → falling back to DXVK")
		} else {
			cfg.Engine = "wine-only"
			log("No GPU translation layer available → Wine-only (no D3D acceleration)")
		}
	}
	if !sys.Engines.DXVK && cfg.Engine == "dxvk" {
		if sys.Engines.GPTK {
			cfg.Engine = "d3dmetal"
			log("DXVK not installed, GPTK available → switching to D3DMetal")
		} else {
			cfg.Engine = "wine-only"
			log("No GPU translation layer available → Wine-only (no D3D acceleration)")
		}
	}
}

// ---------------------------------------------------------------------------
// Confidence calculation
// ---------------------------------------------------------------------------

func calcConfidence(data *AggregatedGameData, sys *SystemProfile, engineName string) float64 {
	var c float64

	if engineName != "" {
		c += confWeightEngineKnown
	}
	if data.ProtonDBBest != nil {
		rating := data.ProtonDBBest.Rating
		switch rating {
		case "Platinum":
			c += 0.35
		case "Gold":
			c += confWeightProtonDBGold
		case "Silver":
			c += 0.20
		case "Bronze":
			c += 0.10
		}
	}
	if data.LutrisBest != nil {
		c += confWeightLutrisScript
	}
	if data.PCGamingWiki != nil && data.ProtonDBBest != nil {
		c += confWeightConsistency // both sources agree on something
	}
	// System profile bonus.
	if sys.MetalGPUFamily >= 4 && sys.Engines.GPTK {
		c += 0.10 // modern hardware gives more confidence in D3DMetal
	}
	if c > 1.0 {
		c = 1.0
	}
	return c
}
