package rift

import (
	"rift/internal/engine/goldberg"
)

// SetupSteamNative performs post-download configuration for Mac Native games downloaded via SteamCMD.
func (a *App) SetupSteamNative(gameID, appID, installDir string) error {
	a.logInfo("[Setup] Starting setup for Steam Native game: %s", gameID)

	// Inject Goldberg DRM emulator for Steam games (Mac versions sometimes require it too)
	if err := goldberg.InjectSteamEmulator(installDir, appID); err != nil {
		a.logInfo("[DRM] Goldberg injection note: %v", err)
	}

	a.logInfo("[Setup] Steam Native setup complete for %s", gameID)
	return nil
}
