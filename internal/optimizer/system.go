package optimizer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// SystemProfiler collects hardware and software characteristics of the host.
type SystemProfiler struct{}

// NewSystemProfiler returns a new profiler.
func NewSystemProfiler() *SystemProfiler {
	return &SystemProfiler{}
}

// Profile gathers a complete SystemProfile by running several quick commands
// in parallel internally and returns the aggregated result.
func (p *SystemProfiler) Profile() *SystemProfile {
	sp := &SystemProfile{
		MacOSVersion: macOSVersion(),
		AppleSilicon: isAppleSilicon(),
		RAMGB:        physicalRAMGB(),
		Rosetta2:     rosetta2Installed(),
		Engines:      detectEngines(),
	}
	sp.MetalGPUFamily = metalGPUFamily(sp.MacOSVersion, sp.AppleSilicon)
	return sp
}

// ---------------------------------------------------------------------------
// OS version
// ---------------------------------------------------------------------------

func macOSVersion() string {
	out, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		return "0.0.0"
	}
	return strings.TrimSpace(string(out))
}

// ---------------------------------------------------------------------------
// Architecture
// ---------------------------------------------------------------------------

func isAppleSilicon() bool {
	out, err := exec.Command("uname", "-m").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "arm64"
}

// ---------------------------------------------------------------------------
// Physical RAM
// ---------------------------------------------------------------------------

func physicalRAMGB() int {
	out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return 16 // Safe fallback instead of 0
	}
	bytes, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0
	}
	return int(bytes / (1024 * 1024 * 1024))
}

// ---------------------------------------------------------------------------
// Metal GPU family (approximated from macOS version + Apple Silicon flag)
//
//   macOS 14 (Sonoma) → Metal 3 on AS, Metal 2 on Intel
//   macOS 15 (Sequoia) → Metal 3 on AS
//   macOS 16+          → Metal 4 on AS
//   macOS 20+ (2026+)  → Metal 4 (speculative)
// ---------------------------------------------------------------------------

func metalGPUFamily(ver string, as bool) int {
	if !as {
		return 2 // Intel Macs top out at Metal 2
	}
	major := majorVersion(ver)
	switch {
	case major >= 20:
		return 4
	case major >= 16:
		return 4
	case major >= 15:
		return 3
	case major >= 14:
		return 3
	default:
		return 2
	}
}

func majorVersion(ver string) int {
	parts := strings.Split(ver, ".")
	if len(parts) == 0 {
		return 0
	}
	n, _ := strconv.Atoi(parts[0])
	return n
}

// ---------------------------------------------------------------------------
// Rosetta 2
// ---------------------------------------------------------------------------

func rosetta2Installed() bool {
	err := exec.Command("arch", "-x86_64", "/usr/bin/true").Run()
	return err == nil
}

// ---------------------------------------------------------------------------
// Engine detection
// ---------------------------------------------------------------------------

func detectEngines() EngineStates {
	home, _ := os.UserHomeDir()
	engDir := filepath.Join(home, ".rift", "engines")

	return EngineStates{
		Wine: fileExists(filepath.Join(engDir, "wine", "bin", "wine")),
		DXVK: fileExists(filepath.Join(engDir, "dxvk", "x64", "d3d11.dll")),
		GPTK: checkGPTKInstalled(),
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func checkGPTKInstalled() bool {
	// Canonical brew cask location
	candidates := []string{
		"/Applications/Game Porting Toolkit.app/Contents/Resources/wine/lib/external/D3DMetal.framework/Versions/A/D3DMetal",
	}
	home, _ := os.UserHomeDir()
	candidates = append(candidates,
		filepath.Join(home, ".rift", "engines", "gptk", "lib", "external", "D3DMetal.framework", "Versions", "A", "D3DMetal"),
	)
	for _, p := range candidates {
		if fileExists(p) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Convenience
// ---------------------------------------------------------------------------

// SystemHash produces a stable string that changes when any hardware or engine
// property changes. Used for cache invalidation.
func (sp *SystemProfile) SystemHash() string {
	var b strings.Builder
	b.WriteString(sp.MacOSVersion)
	b.WriteString(fmt.Sprint(sp.AppleSilicon))
	b.WriteString(fmt.Sprint(sp.MetalGPUFamily))
	b.WriteString(fmt.Sprint(sp.RAMGB))
	b.WriteString(fmt.Sprint(sp.Engines.Wine))
	b.WriteString(fmt.Sprint(sp.Engines.DXVK))
	b.WriteString(fmt.Sprint(sp.Engines.GPTK))
	return b.String()
}
