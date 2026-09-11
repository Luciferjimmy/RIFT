package core

import (
	"debug/pe"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Result holds detected game properties.
type Result struct {
	Engine string // "unreal-engine-4", "unreal-engine-5", "unity", "unknown"
	DX     string // "dx9", "dx10", "dx11", "dx12", "vulkan", "opengl", "unknown"
}

// Analyze reads the game directory and .exe import table to detect engine and DX version.
func Analyze(exePath, gameDir string) *Result {
	r := &Result{Engine: "unknown", DX: "unknown"}

	// 1. Check for engine-specific files in the game directory
	detectEngine(r, gameDir)
	detectByImportTable(r, exePath)

	// 2. If engine is unknown, check the directory name for clues
	if r.Engine == "unknown" {
		detectEngineByFiles(r, gameDir)
	}

	return r
}

func detectEngine(r *Result, gameDir string) {
	entries, err := os.ReadDir(gameDir)
	if err != nil {
		return
	}

	for _, e := range entries {
		name := strings.ToLower(e.Name())

		// Unreal Engine detection
		if strings.Contains(name, "engine") || strings.Contains(name, "ue4") || strings.Contains(name, "ue5") {
			// Check for UE4/5 specific files
			if hasUEFiles(gameDir) {
				// Determine UE4 vs UE5 by looking at specific DLLs
				r.Engine = detectUEVersion(gameDir)
			}
		}

		// Unity detection
		if strings.Contains(name, "unity") || strings.Contains(name, "_data") && e.IsDir() {
			if hasUnityFiles(gameDir) {
				r.Engine = "unity"
			}
		}

		// Godot detection
		if strings.HasSuffix(name, ".pck") || strings.HasSuffix(name, "godot") {
			r.Engine = "godot"
		}
	}
}

func hasUEFiles(gameDir string) bool {
	// Look for UE4/5 specific DLLs in the game directory
	ueDLLs := []string{"ue4game", "ue5game", "unrealengine"}

	filepath.Walk(gameDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		lower := strings.ToLower(info.Name())
		for _, dll := range ueDLLs {
			if strings.Contains(lower, dll) {
				return os.ErrExist // signal found
			}
		}
		return nil
	})

	// Also check for Engine/ directory at root
	engineDir := filepath.Join(gameDir, "Engine")
	if info, err := os.Stat(engineDir); err == nil && info.IsDir() {
		return true
	}

	// Check for .uproject file (UE project file)
	entries, _ := os.ReadDir(gameDir)
	for _, e := range entries {
		if strings.HasSuffix(strings.ToLower(e.Name()), ".uproject") {
			return true
		}
		// Sometimes inside a "GameName" folder
		if e.IsDir() {
			subEntries, _ := os.ReadDir(filepath.Join(gameDir, e.Name()))
			for _, se := range subEntries {
				if strings.HasSuffix(strings.ToLower(se.Name()), ".uproject") {
					return true
				}
			}
		}
	}

	return false
}

func detectUEVersion(gameDir string) string {
	// Check for specific DLLs that differentiate UE4 vs UE5
	ue5DLLs := []string{"ue5game", "ue5", "unrealengine5"}

	filepath.Walk(gameDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		lower := strings.ToLower(info.Name())
		for _, dll := range ue5DLLs {
			if strings.Contains(lower, dll) {
				return os.ErrExist
			}
		}
		return nil
	})

	// Fallback: check for UE5-specific file patterns
	if hasUE5Files(gameDir) {
		return "unreal-engine-5"
	}

	return "unreal-engine-4"
}

func hasUE5Files(gameDir string) bool {
	// UE5 games often have specific metadata
	filepath.Walk(gameDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		name := strings.ToLower(info.Name())
		// Lumen, Nanite, etc. specific to UE5
		if strings.Contains(name, "lumen") || strings.Contains(name, "nanite") {
			return os.ErrExist
		}
		return nil
	})
	return false
}

func hasUnityFiles(gameDir string) bool {
	// Look for Unity player DLL
	filepath.Walk(gameDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		lower := strings.ToLower(info.Name())
		if strings.Contains(lower, "unityplayer.dll") {
			return os.ErrExist
		}
		if strings.Contains(lower, "mono.dll") || strings.Contains(lower, "mono-2.0") {
			return os.ErrExist
		}
		return nil
	})

	// Check for typical Unity folder structure: GameName_Data/
	entries, _ := os.ReadDir(gameDir)
	for _, e := range entries {
		if e.IsDir() && strings.HasSuffix(e.Name(), "_Data") {
			// Check for managed assemblies inside
			managedDir := filepath.Join(gameDir, e.Name(), "Managed")
			if info, err := os.Stat(managedDir); err == nil && info.IsDir() {
				return true
			}
		}
	}

	return false
}

func detectByImportTable(r *Result, exePath string) {
	// Open the PE file and read its import table using debug/pe
	f, err := pe.Open(exePath)
	if err == nil {
		defer f.Close()
		imports, err := f.ImportedLibraries()
		if err == nil {
			if checkGraphicsImports(r, imports) {
				return // PE import table gave us an authoritative result
			}
		}
	}

	// PE parsing didn't find graphics DLLs, try file heuristics
	r.DX = detectDLLFromFiles(exePath)
	if r.DX != "unknown" {
		return
	}

	// If we still don't know, use engine-based defaults
	if r.Engine != "unknown" {
		switch r.Engine {
		case "unreal-engine-4":
			r.DX = "dx11" // Most UE4 games support DX11
		case "unreal-engine-5":
			r.DX = "dx12" // UE5 defaults to DX12
		case "unity":
			r.DX = "dx11" // Most Unity games are DX11
		}
	}
}

// checkGraphicsImports analyzes the PE import list for graphics API DLLs.
// Returns true if a known graphics API was detected.
func checkGraphicsImports(r *Result, imports []string) bool {
	importSet := make(map[string]bool, len(imports))
	for _, dll := range imports {
		importSet[strings.ToLower(dll)] = true
	}

	switch {
	case importSet["d3d12.dll"]:
		r.DX = "dx12"
		return true
	case importSet["vulkan-1.dll"]:
		r.DX = "vulkan"
		return true
	case importSet["d3d11.dll"]:
		r.DX = "dx11"
		return true
	case importSet["d3d10.dll"] || importSet["d3d10core.dll"] || importSet["d3d10_1.dll"]:
		r.DX = "dx10"
		return true
	case importSet["d3d9.dll"]:
		r.DX = "dx9"
		return true
	case importSet["opengl32.dll"]:
		r.DX = "opengl"
		return true
	default:
		return false
	}
}

// detectDLLFromFiles falls back to checking for bundled DLL files in the game directory.
// This is less reliable than PE import parsing, but catches games that load DLLs dynamically.
func detectDLLFromFiles(exePath string) string {
	exeDir := filepath.Dir(exePath)

	// Check for bundled DLL files in the game directory (indicates DX preference)
	checks := []struct {
		path string
		dx   string
	}{
		{filepath.Join(exeDir, "d3d12.dll"), "dx12"},
		{filepath.Join(exeDir, "vulkan-1.dll"), "vulkan"},
		{filepath.Join(exeDir, "d3d11.dll"), "dx11"},
		{filepath.Join(exeDir, "d3d9.dll"), "dx9"},
	}

	for _, c := range checks {
		if _, err := os.Stat(c.path); err == nil {
			return c.dx
		}
	}

	return "unknown"
}

func detectEngineByFiles(r *Result, gameDir string) {
	// Quick scan for common engine indicators
	filepath.Walk(gameDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		name := strings.ToLower(info.Name())

		if strings.Contains(name, "ue4") || strings.Contains(name, "ue4game") {
			r.Engine = "unreal-engine-4"
			return filepath.SkipAll
		}
		if strings.Contains(name, "ue5") || strings.Contains(name, "ue5game") {
			r.Engine = "unreal-engine-5"
			return filepath.SkipAll
		}
		if strings.Contains(name, "unityplayer.dll") {
			r.Engine = "unity"
			return filepath.SkipAll
		}
		return nil
	})
}

// SelectBackend chooses the GPU backend based on detected DX version.
func SelectBackend(dx string) string {
	switch dx {
	case "dx12":
		return "d3dmetal"
	case "dx11":
		return "dxvk"
	case "dx10":
		return "dxvk"
	case "dx9":
		return "dxvk"
	case "vulkan":
		return "moltenvk"
	case "opengl":
		return "wined3d"
	default:
		return "dxvk" // safest fallback
	}
}

// FindGameExe searches common patterns to find the main game executable.
func FindGameExe(gameDir string) (string, error) {
	var candidates []string

	filepath.Walk(gameDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(info.Name()), ".exe") {
			// Skip launcher executables
			lower := strings.ToLower(info.Name())
			if strings.Contains(lower, "launcher") || strings.Contains(lower, "installer") ||
				strings.Contains(lower, "setup") || strings.Contains(lower, "unins") {
				return nil
			}

			// Prefer larger .exe files (the real game, not a helper)
			if info.Size() > 100*1024 { // > 100KB
				candidates = append(candidates, path)
			}
		}
		return nil
	})

	if len(candidates) == 0 {
		return "", fmt.Errorf("no game executable found in %s", gameDir)
	}

	// Return the largest .exe (most likely the actual game)
	var best string
	var bestSize int64
	for _, c := range candidates {
		info, _ := os.Stat(c)
		if info.Size() > bestSize {
			best = c
			bestSize = info.Size()
		}
	}

	return best, nil
}

// FallbackChain returns the ordered list of backends to try.
func FallbackChain(primary string) []string {
	order := []string{"d3dmetal", "dxvk", "moltenvk", "wined3d"}

	// Move primary to front
	result := []string{primary}
	for _, b := range order {
		if b != primary {
			result = append(result, b)
		}
	}

	return result
}
