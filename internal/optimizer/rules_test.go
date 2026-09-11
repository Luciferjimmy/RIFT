package optimizer

import (
	"reflect"
	"testing"
)

func TestApplyEngineRule(t *testing.T) {
	tests := []struct {
		engine   string
		expected string
	}{
		{"Unreal Engine 4", "d3dmetal"},
		{"Unreal Engine 5", "d3dmetal"},
		{"UE3", "dxvk"},
		{"Unity", "d3dmetal"},
		{"RAGE", "dxvk"},
		{"UnknownEngine", "d3dmetal"},
	}

	for _, tt := range tests {
		t.Run(tt.engine, func(t *testing.T) {
			cfg := &GameConfig{Engine: "d3dmetal"} // Default
			log := func(msg string) {}

			applyEngineRule(tt.engine, cfg, log)

			if cfg.Engine != tt.expected {
				t.Errorf("Expected engine %s, got %s", tt.expected, cfg.Engine)
			}
		})
	}
}

func TestApplyProtonDBRule(t *testing.T) {
	pb := &ProtonDBEntry{
		Rating:     "Borked",
		LaunchArgs: "-novid",
		EnvVars:    "DXVK_HUD=1 WINEFSYNC=1",
		Winetricks: "corefonts d3dcompiler_47",
	}

	cfg := &GameConfig{
		Engine:     "dxvk",
		EnvVars:    make(map[string]string),
		Winetricks: []string{},
	}
	log := func(msg string) {}

	applyProtonDBRule(pb, cfg, log)

	if cfg.Engine != "d3dmetal" {
		t.Errorf("Expected Engine to switch to d3dmetal for Borked game, got %s", cfg.Engine)
	}
	if cfg.LaunchArgs != "-novid" {
		t.Errorf("Expected LaunchArgs -novid, got %s", cfg.LaunchArgs)
	}
	if cfg.EnvVars["DXVK_HUD"] != "1" || cfg.EnvVars["WINEFSYNC"] != "1" {
		t.Errorf("Expected EnvVars parsed properly, got %v", cfg.EnvVars)
	}

	expectedTricks := []string{"corefonts", "d3dcompiler_47"}
	if !reflect.DeepEqual(cfg.Winetricks, expectedTricks) {
		t.Errorf("Expected Winetricks %v, got %v", expectedTricks, cfg.Winetricks)
	}
}

func TestApplySystemRule(t *testing.T) {
	// Test low RAM forcing DXVK
	sys1 := &SystemProfile{
		RAMGB: 4, // Low RAM
		Engines: EngineStates{
			GPTK: true,
			DXVK: true,
		},
	}
	cfg1 := &GameConfig{Engine: "d3dmetal"}
	applySystemRule(sys1, cfg1, func(s string) {})
	if cfg1.Engine != "dxvk" {
		t.Errorf("Expected low RAM to force DXVK, got %s", cfg1.Engine)
	}

	// Test missing GPTK fallback
	sys2 := &SystemProfile{
		RAMGB: 16,
		Engines: EngineStates{
			GPTK: false, // Missing
			DXVK: true,
		},
	}
	cfg2 := &GameConfig{Engine: "d3dmetal"}
	applySystemRule(sys2, cfg2, func(s string) {})
	if cfg2.Engine != "dxvk" {
		t.Errorf("Expected missing GPTK to force DXVK, got %s", cfg2.Engine)
	}
}
