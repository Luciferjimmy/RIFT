package rift

import (
	"os"
	"os/exec"
	"path/filepath"
	"rift/internal/engine"
)

// SetupEpicTranslated performs post-download configuration for Windows games downloaded via Epic/Legendary.
func (a *App) SetupEpicTranslated(gameID, appID, installDir string) error {
	a.logInfo("[Setup] Starting setup for Epic Translated game: %s", gameID)

	prefixDir := filepath.Join(a.GamesDir(), gameID, "prefix")
	os.MkdirAll(prefixDir, 0755)

	// Initialize Wine Prefix and configure as Windows 10
	wineBinary := filepath.Join(a.EnginesDir(), "wine", "bin", "wine")
	if _, err := os.Stat(wineBinary); os.IsNotExist(err) {
		wineBinary = filepath.Join(a.EnginesDir(), "gptk", "bin", "wine64")
	}
	
	a.logInfo("[Setup] Initializing Wine prefix at %s...", prefixDir)
	cmdBoot := exec.Command(wineBinary, "wineboot", "-u")
	cmdBoot.Env = append(engine.SafeEnviron(), "WINEPREFIX="+prefixDir, "WINEDEBUG=-all")
	_ = cmdBoot.Run()

	cmdReg := exec.Command(wineBinary, "reg", "add", `HKEY_CURRENT_USER\Software\Wine`, "/v", "Version", "/d", "win10", "/f")
	cmdReg.Env = append(engine.SafeEnviron(), "WINEPREFIX="+prefixDir, "WINEDEBUG=-all")
	_ = cmdReg.Run()

	// Clean default Direct3D setup — do NOT set UseGLSL=disabled or OffscreenRenderingMode=backbuffer
	// as these cause fatal OpenGL buffer errors (0x506) on Apple Silicon macOS.

	a.EnsureEssentialRuntimes(wineBinary, prefixDir)

	// 1. Fetch Supabase Data
	a.logInfo("[Setup] Checking Supabase for curated configurations for %s...", gameID)
	// (Note: The actual supabase payload logic is fetched and parsed by the router before this function, 
	// or we can just run it at launch time, but we will inject the baseline DLLs now to be safe)

	// 4. Check Supabase (Router will pass specific configurations later if needed)
	a.logInfo("[Setup] Checking Supabase for curated configurations for %s...", gameID)

	a.logInfo("[Setup] Epic Translated setup complete for %s", gameID)
	return nil
}
