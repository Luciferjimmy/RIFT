package optimizer

import (
	"fmt"
	"strings"
)

// Synthesizer merges outputs from the rules engine, AI reasoner, and browser
// enricher into a single coherent GameConfig. Rules-engine outputs are used
// as the base; AI outputs override specific fields when the AI was confident;
// browser enrichment adds a confidence boost.
type Synthesizer struct{}

// NewSynthesizer creates a synthesizer.
func NewSynthesizer() *Synthesizer {
	return &Synthesizer{}
}

// Merge takes the rules-engine config (always present) and optionally merges
// AI and browser results on top.
func (s *Synthesizer) Merge(
	rules *EvaluateResult,
	ai *AIFillResult,
	browser *EnrichResult,
) *GameConfig {

	cfg := &GameConfig{
		Engine:         rules.Config.Engine,
		WineVersion:    rules.Config.WineVersion,
		WindowsVersion: rules.Config.WindowsVersion,
		DLLOverrides:   copyMap(rules.Config.DLLOverrides),
		EnvVars:        copyMap(rules.Config.EnvVars),
		LaunchArgs:     rules.Config.LaunchArgs,
		Winetricks:     append([]string{}, rules.Config.Winetricks...),
	}
	cfg.PipelineLog = append(cfg.PipelineLog, rules.Log...)

	// ---- AI merge (overrides rules when confident) ----
	if ai != nil {
		cfg.PipelineLog = append(cfg.PipelineLog, "AI reasoning applied")

		if ai.TranslationLayer != "SKIP" {
			layer := strings.ToLower(strings.TrimSpace(ai.TranslationLayer))
			if layer == "d3dmetal" || layer == "dxvk" {
				cfg.Engine = layer
				cfg.PipelineLog = append(cfg.PipelineLog, fmt.Sprintf("AI → engine=%s", layer))
			}
		}
		if ai.WineVersion != "SKIP" && ai.WineVersion != "" {
			cfg.WineVersion = ai.WineVersion
		}
		if ai.WindowsVersion != "SKIP" && ai.WindowsVersion != "" {
			cfg.WindowsVersion = ai.WindowsVersion
		}
		if ai.LaunchArgs != "SKIP" && ai.LaunchArgs != "" {
			cfg.LaunchArgs = ai.LaunchArgs
		}
		for _, dll := range ai.DLLOverrides {
			dll = strings.TrimSpace(dll)
			if dll == "" {
				continue
			}
			k, v, ok := strings.Cut(dll, "=")
			if ok {
				cfg.DLLOverrides[strings.TrimSpace(k)] = strings.TrimSpace(v)
			} else {
				cfg.DLLOverrides[dll] = "n,b" // default override
			}
		}
		for _, wt := range ai.Winetricks {
			wt = strings.TrimSpace(wt)
			if wt != "" {
				cfg.Winetricks = append(cfg.Winetricks, wt)
			}
		}
		for _, ev := range ai.EnvVars {
			ev = strings.TrimSpace(ev)
			if ev == "" {
				continue
			}
			k, v, ok := strings.Cut(ev, "=")
			if ok {
				cfg.EnvVars[strings.TrimSpace(k)] = strings.TrimSpace(v)
			}
		}
		// Confidence: use AI confidence if higher than rules.
		if ai.Confidence > rules.Confidence {
			cfg.Confidence = ai.Confidence
		} else {
			cfg.Confidence = rules.Confidence
		}
		cfg.Source = "hybrid"
	} else {
		cfg.Confidence = rules.Confidence
		cfg.Source = "rules"
	}

	// ---- Browser enrichment (confidence boost) ----
	if browser != nil {
		cfg.Confidence += browser.ConfidenceDelta
		if cfg.Confidence > 1.0 {
			cfg.Confidence = 1.0
		}
		cfg.PipelineLog = append(cfg.PipelineLog, browser.Log...)
		if browser.ProtonDBHint != "" {
			cfg.PipelineLog = append(cfg.PipelineLog, "Browser: "+browser.ProtonDBHint)
		}
		if browser.ConfidenceDelta > 0 {
			cfg.Source = "browser"
		}
	}

	// ---- Final cleanup ----

	// Remove duplicate winetricks.
	cfg.Winetricks = uniqueStrings(cfg.Winetricks)

	// Engine-specific overrides.
	switch cfg.Engine {
	case "d3dmetal":
		if _, ok := cfg.DLLOverrides["d3d11"]; !ok {
			cfg.DLLOverrides["d3d11"] = "n,b"
		}
		if _, ok := cfg.DLLOverrides["dxgi"]; !ok {
			cfg.DLLOverrides["dxgi"] = "n,b"
		}
		if _, ok := cfg.EnvVars["WINEMSYNC"]; !ok {
			cfg.EnvVars["WINEMSYNC"] = "1"
		}

	case "dxvk":
		if _, ok := cfg.DLLOverrides["d3d11"]; !ok {
			cfg.DLLOverrides["d3d11"] = "n,b"
		}
		if _, ok := cfg.DLLOverrides["d3d10core"]; !ok {
			cfg.DLLOverrides["d3d10core"] = "n,b"
		}
		if _, ok := cfg.EnvVars["WINEESYNC"]; !ok {
			cfg.EnvVars["WINEESYNC"] = "1"
		}

	case "native":
		// No DLL overrides or env vars needed.
		cfg.DLLOverrides = nil
		cfg.EnvVars = nil
		cfg.Winetricks = nil
	}

	cfg.PipelineLog = append(cfg.PipelineLog,
		fmt.Sprintf("Final: engine=%s, winver=%s, confidence=%.2f, source=%s",
			cfg.Engine, cfg.WindowsVersion, cfg.Confidence, cfg.Source))

	return cfg
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func copyMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func uniqueStrings(s []string) []string {
	seen := make(map[string]bool, len(s))
	uniq := make([]string, 0, len(s))
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			uniq = append(uniq, v)
		}
	}
	return uniq
}
