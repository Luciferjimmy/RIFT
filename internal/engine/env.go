package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SafeEnviron returns a copy of strings representing the environment,
// in the form "key=value", but strips out any sensitive API keys
// to prevent leakage to child processes.
// It also strips Wine version-specific environment variables that could
// cause version mismatches between different Wine builds (e.g. GPTK vs Gcenx).
func SafeEnviron() []string {
	var env []string
	for _, e := range os.Environ() {
		// Strip sensitive keys
		if strings.HasPrefix(e, "SUPABASE_") ||
			strings.HasPrefix(e, "GROQ_") ||
			strings.HasPrefix(e, "OPENROUTER_") ||
			strings.HasPrefix(e, "OPENAI_") {
			continue
		}
		env = append(env, e)
	}
	return env
}

// WineEnviron returns SafeEnviron plus the WINESERVER and WINELOADER env
// vars pinned to the given wineBinary directory. This prevents cross-version
// wineserver conflicts when multiple Wine builds are present (e.g. GPTK vs Gcenx).
func WineEnviron(wineBinary, winePrefix string) []string {
	base := SafeEnviron()

	// Determine the Wine binary directory to pin WINESERVER/WINELOADER
	wineDir := filepath.Dir(wineBinary)

	// Strip any host Wine env vars that would cause version mismatch
	var filtered []string
	for _, e := range base {
		key := e
		if idx := strings.Index(e, "="); idx != -1 {
			key = e[:idx]
		}
		upper := strings.ToUpper(key)
		if upper == "WINESERVER" || upper == "WINELOADER" || upper == "WINEDLLPATH" {
			continue
		}
		filtered = append(filtered, e)
	}

	// Pin WINESERVER and WINELOADER to the same Wine build
	filtered = append(filtered, "WINESERVER="+filepath.Join(wineDir, "wineserver"))
	filtered = append(filtered, "WINELOADER="+wineBinary)

	if winePrefix != "" {
		filtered = append(filtered, "WINEPREFIX="+winePrefix)
	}

	return filtered
}

// KillWineServer kills any running wineserver in the given prefix or system-wide.
// This prevents "wine client error: version mismatch" when switching between
// different Wine builds (e.g. GPTK wine64 vs Gcenx Wine staging).
func KillWineServer(wineBinary, winePrefix string) error {
	// Build the kill command: wineserver -k in the target prefix
	// We need to use the correct wineserver binary from the same Wine build
	wineDir := filepath.Dir(wineBinary)
	wsBinary := filepath.Join(wineDir, "wineserver")

	if _, err := os.Stat(wsBinary); err != nil {
		// If wineserver doesn't exist next to wine, just try system wineserver
		wsBinary = "wineserver"
	}

	killCmd := exec.Command(wsBinary, "-k")
	env := SafeEnviron()
	env = append(env, "WINEPREFIX="+winePrefix)
	// Strip any conflicting WINESERVER from host env
	var filtered []string
	for _, e := range env {
		if strings.HasPrefix(e, "WINESERVER=") {
			continue
		}
		filtered = append(filtered, e)
	}
	killCmd.Env = append(filtered, "WINESERVER="+wsBinary)

	// Ignore errors — wineserver may not be running
	output, _ := killCmd.CombinedOutput()
	if len(output) > 0 {
		fmt.Printf("[WineEnv] wineserver kill output: %s\n", string(output))
	}
	return nil
}
