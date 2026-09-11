package engine

import (
	"rift/internal/types"
	"testing"
)

func TestApplyHardwareHints_D3DMetal(t *testing.T) {
	config := &types.RuntimeConfig{
		Engine: "d3dmetal",
	}
	sysProfile := types.SystemProfile{
		RAMGB: 8,
	}

	ApplyHardwareHints(config, sysProfile)

	if config.EnvVars["D3DM_SHADER_CACHE"] != "1" {
		t.Errorf("Expected D3DM_SHADER_CACHE to be 1 for d3dmetal engine")
	}
}

func TestApplyHardwareHints_DXVK(t *testing.T) {
	config := &types.RuntimeConfig{
		Engine: "dxvk",
	}
	sysProfile := types.SystemProfile{
		RAMGB: 16,
	}

	ApplyHardwareHints(config, sysProfile)

	if config.EnvVars["DXVK_ASYNC"] != "1" || config.EnvVars["DXVK_STATE_CACHE"] != "1" {
		t.Errorf("Expected DXVK environment variables to be set")
	}
}
