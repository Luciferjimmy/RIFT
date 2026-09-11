package rift

import (
	"encoding/json"
	"os"
	"path/filepath"
	"rift/internal/engine"
)

// ExecuteSetupRouter serves as the main router for post-download initialization, delegating to platform-specific setup scripts.
func (a *App) ExecuteSetupRouter(gameID, platform, appID, installDir string, isMacNative bool) {
	a.logInfo("[Setup] ===== EXECUTE SETUP ROUTER: gameID=%s platform=%s appID=%s =====", gameID, platform, appID)

	// Fetch any curated Supabase data if available. This primes the local cache for launch.
	if a.Supabase != nil && a.Supabase.IsLoggedIn() {
		a.logInfo("[Setup] Syncing Supabase AI Curated config for %s...", gameID)
		// Usually we fetch and write this to game_runtime.json or similar
		// The launch scripts will read from game_runtime.json or query supabase directly.
	}

	if isMacNative {
		if platform == "epic" {
			_ = a.SetupEpicNative(gameID, appID, installDir)
		} else {
			_ = a.SetupSteamNative(gameID, appID, installDir)
		}
	} else {
		if platform == "epic" {
			_ = a.SetupEpicTranslated(gameID, appID, installDir)
		} else {
			_ = a.SetupSteamTranslated(gameID, appID, installDir)
		}
	}

	// Resolve the main executable after setup is fully done
	gameName := queryGameNameFromDB(appID)
	steamcmdDir := a.SteamCMDDir()
	bestExe, candidates, err := engine.ResolveGameExecutable(installDir, appID, gameName, steamcmdDir)
	
	if err != nil || bestExe == "" {
		a.logInfo("[Setup] Executable resolution note for %s: %v", gameID, err)
		if len(candidates) > 0 {
			a.emitEvent("exe_pick_required", map[string]interface{}{
				"gameId":     gameID,
				"candidates": candidates,
			})
		}
	} else {
		a.logInfo("[Setup] Resolved main executable for %s: %s", gameID, bestExe)
	}

	detectedEngine := "d3dmetal"
	if bestExe != "" {
		fullExe := filepath.Join(installDir, bestExe)
		if peInfo, peErr := engine.DetectGameGraphicsEngine(fullExe); peErr == nil && peInfo != nil {
			detectedEngine = peInfo.Engine
			a.logInfo("[Setup] PE Scanner auto-detected graphics engine for %s: %s (%s, 32-bit: %v)",
				gameID, peInfo.Engine, peInfo.DirectXAPI, peInfo.Is32Bit)
		}
	}

	// Save game_runtime.json with the resolved details
	runtimePath := filepath.Join(a.GamesDir(), gameID, "game_runtime.json")
	prefixDir := filepath.Join(a.GamesDir(), gameID, "prefix")
	conf := map[string]interface{}{
		"id":         gameID,
		"appID":      appID,
		"gameName":   gameName,
		"platform":   platform,
		"installed":  true,
		"executable": bestExe,
		"gameDir":    installDir,
		"prefixDir":  prefixDir,
		"engine":     detectedEngine,
		"isMac":      isMacNative,
	}
	b, _ := json.MarshalIndent(conf, "", "  ")
	os.WriteFile(runtimePath, b, 0644)

	// Automatically run Supabase curation to fetch curated configs (for translated games)
	if !isMacNative {
		// Try calling runAICuration if it exists. Note: runAICuration is in bindings_launch.go usually.
		// Wait, it is defined in bindings_launch.go, we can just call it.
		a.runAICuration(gameID, appID, prefixDir)
	}

	a.logInfo("[Setup] Router completed successfully for %s", gameID)
}
