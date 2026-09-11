package engine

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// EnsureShaderCacheDir creates and returns the persistent shader cache directory
// for a given game. This directory survives reboots and game updates.
func EnsureShaderCacheDir(gameDir string) string {
	dir := filepath.Join(gameDir, "shader_cache")
	os.MkdirAll(dir, 0755)
	return dir
}

// OptimalCompilerThreads returns the ideal number of DXVK pipeline compiler
// threads for the current Apple Silicon chip. Leaves 2 cores free for the
// game's own render and logic threads.
func OptimalCompilerThreads() int {
	cores := runtime.NumCPU()
	threads := cores - 2
	if threads < 4 {
		threads = 4
	}
	if threads > 16 {
		threads = 16
	}
	return threads
}

// GenerateDXVKConf writes an optimized dxvk.conf file into the game's
// executable directory. DXVK reads this file automatically on startup.
//
// Settings are tuned specifically for Apple Silicon + MoltenVK:
//   - State cache enabled and persistent
//   - Multi-threaded pipeline compilation
//   - Graphics Pipeline Library set to Auto for modern DXVK
//   - Verbose logging suppressed for performance
func GenerateDXVKConf(exeDir string, cpuCores int) error {
	if exeDir == "" {
		return fmt.Errorf("empty exe directory")
	}

	threads := cpuCores - 2
	if threads < 4 {
		threads = 4
	}
	if threads > 16 {
		threads = 16
	}

	conf := fmt.Sprintf(`# RIFT Auto-Generated DXVK Configuration
# Optimized for Apple Silicon + MoltenVK
# Do not edit manually — regenerated on each launch.

# Shader State Cache
dxvk.enableStateCache = True

# Pipeline Compilation Threads (detected %d cores, using %d threads)
dxvk.numCompilerThreads = %d

# Graphics Pipeline Library — let DXVK decide based on driver support
dxvk.enableGraphicsPipelineLibrary = Auto

# Suppress verbose logging for performance
dxvk.logLevel = none

# Memory optimizations for MoltenVK
dxvk.shrinkNvidiaHvvHeap = True
`, cpuCores, threads, threads)

	confPath := filepath.Join(exeDir, "dxvk.conf")
	return os.WriteFile(confPath, []byte(conf), 0644)
}

// darwinUserCacheDir returns the Darwin user cache directory by calling
// getconf DARWIN_USER_CACHE_DIR. This is where macOS stores volatile
// per-process caches including D3DMetal shader translations.
func darwinUserCacheDir() (string, error) {
	out, err := exec.Command("getconf", "DARWIN_USER_CACHE_DIR").Output()
	if err != nil {
		return "", fmt.Errorf("getconf DARWIN_USER_CACHE_DIR failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// BackupD3DMetalCache copies the volatile D3DMetal shaders.cache from the
// Darwin user cache directory into the game's persistent shader_cache/d3dm_backup/.
// Called after game exit to preserve compiled Metal pipelines before macOS
// cache cleanup can purge them.
func BackupD3DMetalCache(gameDir, exeName string) error {
	if exeName == "" || gameDir == "" {
		return nil
	}

	cacheBase, err := darwinUserCacheDir()
	if err != nil {
		return err
	}

	// D3DMetal stores caches at: $(DARWIN_USER_CACHE_DIR)/d3dm/<ExeName>/
	baseName := strings.TrimSuffix(filepath.Base(exeName), filepath.Ext(exeName))
	srcDir := filepath.Join(cacheBase, "d3dm", baseName)

	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil // No D3DMetal cache exists yet, nothing to back up
	}

	backupDir := filepath.Join(gameDir, "shader_cache", "d3dm_backup", baseName)
	os.MkdirAll(backupDir, 0755)

	// Copy all files from the D3DMetal cache directory
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to read D3DMetal cache dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		srcFile := filepath.Join(srcDir, entry.Name())
		dstFile := filepath.Join(backupDir, entry.Name())
		if err := shaderCopyFile(srcFile, dstFile); err != nil {
			// Log but don't fail — partial backup is better than none
			continue
		}
	}

	return nil
}

// RestoreD3DMetalCache copies the backed-up D3DMetal shader cache back into
// the Darwin user cache directory before game launch. This prevents D3DMetal
// from having to re-translate all shaders after macOS purges its cache.
func RestoreD3DMetalCache(gameDir, exeName string) error {
	if exeName == "" || gameDir == "" {
		return nil
	}

	baseName := strings.TrimSuffix(filepath.Base(exeName), filepath.Ext(exeName))
	backupDir := filepath.Join(gameDir, "shader_cache", "d3dm_backup", baseName)

	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return nil // No backup exists, nothing to restore
	}

	cacheBase, err := darwinUserCacheDir()
	if err != nil {
		return err
	}

	dstDir := filepath.Join(cacheBase, "d3dm", baseName)
	os.MkdirAll(dstDir, 0755)

	// Only restore if the target cache is empty or missing
	dstEntries, _ := os.ReadDir(dstDir)
	if len(dstEntries) > 0 {
		return nil // Cache already populated, don't overwrite active cache
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("failed to read D3DMetal backup dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		srcFile := filepath.Join(backupDir, entry.Name())
		dstFile := filepath.Join(dstDir, entry.Name())
		if err := shaderCopyFile(srcFile, dstFile); err != nil {
			continue
		}
	}

	return nil
}

// shaderCopyFile copies a single file from src to dst.
// Named distinctly to avoid conflict with copyFile in wine.go.
func shaderCopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
