package engine

import (
	"fmt"
	"strings"

	"rift/internal/types"
)

// ApplyHardwareHints takes the cloud config and applies local device-specific tweaks.
func ApplyHardwareHints(config *types.RuntimeConfig, sysProfile types.SystemProfile) {
	if config.EnvVars == nil {
		config.EnvVars = make(map[string]string)
	}

	// WINE_LARGE_ADDRESS_AWARE is dynamically set per-executable by the launcher pipeline

	// M1/M2 (Apple7/Apple8): conservative Metal settings
	if sysProfile.GPUMetalFamily == "Apple7" || sysProfile.GPUMetalFamily == "Apple8" {
		config.EnvVars["MTL_HUD_ENABLED"] = "0"
	}

	// DXVK specific tweaks
	if strings.ToLower(config.Engine) == "dxvk" {
		config.EnvVars["DXVK_ASYNC"] = "1"
		config.EnvVars["DXVK_STATE_CACHE"] = "1"
		config.EnvVars["DXVK_LOG_LEVEL"] = "none"
		// Multi-core pipeline compilation — use all Apple Silicon performance cores
		config.EnvVars["DXVK_NUM_COMPILER_THREADS"] = fmt.Sprintf("%d", OptimalCompilerThreads())
	}

	// D3DMetal / GPTK: ensure shader JIT cache is active
	engineLower := strings.ToLower(config.Engine)
	if engineLower == "d3dmetal" || engineLower == "gptk" {
		config.EnvVars["D3DM_SHADER_CACHE"] = "1"
	}
}
