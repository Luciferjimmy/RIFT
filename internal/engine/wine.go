package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// Game Porting Toolkit prebuilt cask from Gcenx
	GPTKCaskName = "gcenx/wine/game-porting-toolkit"

	// Common paths where D3DMetal.framework might be installed
	GPTKBrewPath    = "/Applications/Game Porting Toolkit.app/Contents/Resources/wine/lib/external/D3DMetal.framework/Versions/A/D3DMetal"
	GPTKBrewAltPath = "/opt/homebrew/Caskroom/game-porting-toolkit"
)

// engineAsset represents a downloadable engine component: the live GitHub
// release asset name pattern plus a verified working fallback URL used when
// the GitHub API is unavailable (rate-limited, offline, etc.).
type engineAsset struct {
	repo           string
	assetSuffixes  []string // asset names containing any of these are candidates
	preferSubstr   []string // among candidates, prefer those containing these (e.g. "staging")
	rejectSubstr   []string // skip candidates containing these (e.g. "builtin")
	fallbackURL    string   // known-good release asset, used when the API cannot be reached
	verifyRelPath  string   // file that must exist after extraction (fail-proof check)
}

// fullFallbackURLs are latest-known-good release URLs, verified on 2026-08-03.
const (
	fallbackWineStaging = "https://github.com/Gcenx/macOS_Wine_builds/releases/download/11.10/wine-staging-11.10-osx64.tar.xz"
	fallbackDXVK        = "https://github.com/Gcenx/DXVK-macOS/releases/download/v1.10.3-20230507-repack/dxvk-macOS-async-v1.10.3-20230507-repack.tar.gz"
	fallbackGPTK        = "https://github.com/Gcenx/game-porting-toolkit/releases/download/Game-Porting-Toolkit-3.0-3/game-porting-toolkit-3.0-3.tar.xz"
)

func wineEngineAsset() engineAsset {
	return engineAsset{
		repo:           "Gcenx/macOS_Wine_builds",
		assetSuffixes:  []string{"wine-staging-", "wine-devel-", "osx64.tar.xz"},
		preferSubstr:   []string{"staging"},
		rejectSubstr:   nil,
		fallbackURL:    fallbackWineStaging,
		verifyRelPath:  "bin/wine",
	}
}

func dxvkEngineAsset() engineAsset {
	return engineAsset{
		repo:           "Gcenx/DXVK-macOS",
		assetSuffixes:  []string{"-repack.tar.gz", "macOS-async"},
		preferSubstr:   nil,
		rejectSubstr:   []string{"builtin", "built-in"},
		fallbackURL:    fallbackDXVK,
		verifyRelPath:  "x64/d3d11.dll",
	}
}

func gptkEngineAsset() engineAsset {
	return engineAsset{
		repo:           "Gcenx/game-porting-toolkit",
		assetSuffixes:  []string{".tar.xz"},
		preferSubstr:   nil,
		rejectSubstr:   nil,
		fallbackURL:    fallbackGPTK,
		verifyRelPath:  "bin/wine64",
	}
}

// resolveURL fetches the newest release for the repo and returns the most
// suitable asset URL. If the API request fails (rate limit, offline,
// malformed response) it returns a known-good fallback URL.
func (a engineAsset) resolveURL() string {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", a.repo)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return a.fallbackURL
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return a.fallbackURL
	}
	defer resp.Body.Close()

	var result struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return a.fallbackURL
	}

	// 1) Prefer an asset that matches a preferred substring.
	if len(a.preferSubstr) > 0 {
		for _, asset := range result.Assets {
			if a.rejects(asset.Name) {
				continue
			}
			for _, pref := range a.preferSubstr {
				if strings.Contains(asset.Name, pref) && a.matchesSuffix(asset.Name) {
					return asset.BrowserDownloadURL
				}
			}
		}
	}
	// 2) Fall back to any asset matching a suffix.
	for _, asset := range result.Assets {
		if a.rejects(asset.Name) {
			continue
		}
		if a.matchesSuffix(asset.Name) {
			return asset.BrowserDownloadURL
		}
	}
	return a.fallbackURL
}

func (a engineAsset) matchesSuffix(name string) bool {
	for _, s := range a.assetSuffixes {
		if strings.Contains(name, s) {
			return true
		}
	}
	return false
}

func (a engineAsset) rejects(name string) bool {
	for _, r := range a.rejectSubstr {
		if strings.Contains(name, r) {
			return true
		}
	}
	return false
}

// VerifyEngineInstall confirms a required file exists inside the extracted
// engine directory. If it does not, the install is incomplete or the wrong
// asset was downloaded — this is the fail-proof check before any launch.
func VerifyEngineInstall(targetDir, relPath string) error {
	if targetDir == "" || relPath == "" {
		return fmt.Errorf("invalid verify args (dir=%q rel=%q)", targetDir, relPath)
	}
	checkPath := filepath.Join(targetDir, relPath)
	if fi, err := os.Stat(checkPath); err != nil {
		return fmt.Errorf("engine verification failed: %s missing (%v)", checkPath, err)
	} else if fi.IsDir() {
		return fmt.Errorf("engine verification failed: %s is a directory, expected a file", checkPath)
	}
	return nil
}

// EnsureWine makes sure that the central Wine engine is downloaded and
// extracted. It specifically prefers Wine Staging assets.
func (m *EngineManager) EnsureWine(targetDir string) error {
	return m.EnsureWineProgress(targetDir, nil)
}

// EnsureWineProgress is EnsureWine with a download progress callback.
// onProgress receives (downloadedBytes, totalBytes); totalBytes is -1 when
// the server does not report Content-Length.
func (m *EngineManager) EnsureWineProgress(targetDir string, onProgress func(downloaded, total int64)) error {
	asset := wineEngineAsset()
	if err := VerifyEngineInstall(targetDir, asset.verifyRelPath); err == nil {
		return nil
	}

	// Clean up any corrupt/partial install before downloading fresh.
	os.RemoveAll(targetDir)

	fmt.Printf("[EnsureWine] Resolving latest Wine asset...\n")
	url := asset.resolveURL()
	fmt.Printf("[EnsureWine] Downloading %s to %s...\n", url, targetDir)
	if err := m.downloadAndExtract(url, targetDir, onProgress); err != nil {
		return fmt.Errorf("failed to download Wine: %w", err)
	}
	if err := VerifyEngineInstall(targetDir, asset.verifyRelPath); err != nil {
		return fmt.Errorf("Wine extraction incomplete: %w", err)
	}
	fmt.Printf("[EnsureWine] Wine Staging installed to %s\n", targetDir)
	return nil
}

// EnsureDXVK makes sure that the central DXVK bundle is downloaded and extracted.
func (m *EngineManager) EnsureDXVK(targetDir string) error {
	return m.EnsureDXVKProgress(targetDir, nil)
}

// EnsureDXVKProgress is EnsureDXVK with a download progress callback.
func (m *EngineManager) EnsureDXVKProgress(targetDir string, onProgress func(downloaded, total int64)) error {
	asset := dxvkEngineAsset()
	if err := VerifyEngineInstall(targetDir, asset.verifyRelPath); err == nil {
		return nil
	}

	os.RemoveAll(targetDir)

	url := asset.resolveURL()
	fmt.Printf("[EnsureDXVK] Downloading %s to %s...\n", url, targetDir)
	if err := m.downloadAndExtract(url, targetDir, onProgress); err != nil {
		return fmt.Errorf("failed to download DXVK: %w", err)
	}
	if err := VerifyEngineInstall(targetDir, asset.verifyRelPath); err != nil {
		return fmt.Errorf("DXVK extraction incomplete: %w", err)
	}
	fmt.Printf("[EnsureDXVK] DXVK ready in %s\n", targetDir)
	return nil
}

// InstallDXVKToPrefix copies DXVK DLLs from the DXVK-macOS bundle into the Wine prefix.
// Gcenx DXVK-macOS ships d3d11.dll + d3d10core.dll (for D3D11/D3D10 support via Vulkan+MoltenVK).
func (m *EngineManager) InstallDXVKToPrefix(prefixPath string, dxvkDir string) error {
	// DLLs are in dxvkDir/x64/ and dxvkDir/x32/
	sys32 := filepath.Join(prefixPath, "drive_c", "windows", "system32")
	sys64 := filepath.Join(prefixPath, "drive_c", "windows", "syswow64")

	// Ensure target directories exist
	os.MkdirAll(sys32, 0755)
	os.MkdirAll(sys64, 0755)

	// Gcenx DXVK-macOS ships d3d11.dll, d3d10core.dll, dxgi.dll, and d3d9.dll
	dlls := []string{"d3d11.dll", "d3d10core.dll", "dxgi.dll", "d3d9.dll"}

	// Copy x64 files to system32
	x64Src := filepath.Join(dxvkDir, "x64")
	for _, dll := range dlls {
		srcFile := filepath.Join(x64Src, dll)
		destFile := filepath.Join(sys32, dll)
		if _, err := os.Stat(srcFile); err == nil {
			if err := copyFile(srcFile, destFile); err != nil {
				return fmt.Errorf("failed to copy %s to system32: %w", dll, err)
			}
		}
	}

	// Copy x32 files to syswow64
	x32Src := filepath.Join(dxvkDir, "x32")
	for _, dll := range dlls {
		srcFile := filepath.Join(x32Src, dll)
		destFile := filepath.Join(sys64, dll)
		if _, err := os.Stat(srcFile); err == nil {
			if err := copyFile(srcFile, destFile); err != nil {
				return fmt.Errorf("failed to copy %s to syswow64: %w", dll, err)
			}
		}
	}

	return nil
}

// copyFile is a helper to copy file contents from source to destination.
func copyFile(src, dst string) error {
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

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	if err := out.Sync(); err != nil {
		return err
	}
	return nil
}

// downloadAndExtract downloads a tarball and extracts it into the target directory.
func (m *EngineManager) downloadAndExtract(url string, targetDir string, onProgress func(downloaded, total int64)) error {
	// Skip if already exists and contains contents
	if info, err := os.Stat(targetDir); err == nil && info.IsDir() {
		files, errRead := os.ReadDir(targetDir)
		if errRead == nil && len(files) > 1 {
			return nil
		}
	}

	fmt.Printf("[RIFT Engine] Downloading %s to %s...\n", url, targetDir)

	// Determine file extension to name the tmp file correctly
	ext := ".tar.gz"
	if filepath.Ext(url) == ".xz" || (len(url) > 7 && url[len(url)-7:] == ".tar.xz") {
		ext = ".tar.xz"
	}

	// Make sure the parent dir of targetDir exists
	parentDir := filepath.Dir(targetDir)
	os.MkdirAll(parentDir, 0755)

	tmpTar := targetDir + ext
	if err := m.Downloader.DownloadFileProgress(url, tmpTar, false, onProgress); err != nil {
		return err
	}
	defer os.Remove(tmpTar) // Clean up tarball after extraction

	// Extract the tarball into targetDir
	os.MkdirAll(targetDir, 0755)
	fmt.Printf("[RIFT Engine] Extracting %s...\n", tmpTar)
	cmd := exec.Command("tar", "-xf", tmpTar, "-C", targetDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to extract tarball (output: %q): %w", string(out), err)
	}

	// Flatten nested directory if the tarball extracted into a single subdirectory or a macOS app wrapper
	if err := flattenDir(targetDir); err != nil {
		return fmt.Errorf("failed to flatten extracted directory: %w", err)
	}

	// Strip macOS Gatekeeper quarantine attributes from all extracted binaries and frameworks
	// so macOS does not block Wine, DXVK, or D3DMetal execution with "unverified developer" dialogs.
	_ = exec.Command("xattr", "-cr", targetDir).Run()

	return nil
}

// flattenDir checks targetDir and flattens it if it's wrapped in single nested dirs.
func flattenDir(targetDir string) error {
	if targetDir == "" || targetDir == "/" {
		return fmt.Errorf("invalid target directory for flattening: %s", targetDir)
	}

	// Gcenx Crossover builds extract with a root directory structure like:
	// "Wine Crossover.app/Contents/Resources/wine" or a similar folder.
	// Let's recursively walk or locate the actual "bin" directory of Wine to extract it up.
	binDir := ""
	filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() && info.Name() == "bin" {
			// Check if this bin directory has "wine"
			if _, errWine := os.Stat(filepath.Join(path, "wine")); errWine == nil {
				binDir = path
				return filepath.SkipDir
			}
		}
		return nil
	})

	if binDir != "" {
		// The parent of bin is the real Wine prefix
		realWineRoot := filepath.Dir(binDir)
		if realWineRoot == targetDir {
			return nil
		}

		// Move everything from realWineRoot up to targetDir
		_, err := os.ReadDir(realWineRoot)
		if err != nil {
			return err
		}

		// Move to a temporary folder first to avoid collisions
		tempMoveDir := targetDir + "_temp_move"
		os.RemoveAll(tempMoveDir)
		if err := os.Rename(realWineRoot, tempMoveDir); err != nil {
			return err
		}

		// Clear targetDir
		os.RemoveAll(targetDir)
		os.MkdirAll(targetDir, 0755)

		// Move entries from tempMoveDir into targetDir
		tempEntries, _ := os.ReadDir(tempMoveDir)
		for _, entry := range tempEntries {
			oldPath := filepath.Join(tempMoveDir, entry.Name())
			newPath := filepath.Join(targetDir, entry.Name())
			os.Rename(oldPath, newPath)
		}

		os.RemoveAll(tempMoveDir)
		return nil
	}

	// Standard single subdirectory flattening fallback
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return err
	}

	if len(entries) == 1 && entries[0].IsDir() {
		subDirName := entries[0].Name()
		subDirPath := filepath.Join(targetDir, subDirName)

		subEntries, err := os.ReadDir(subDirPath)
		if err != nil {
			return err
		}

		for _, entry := range subEntries {
			oldPath := filepath.Join(subDirPath, entry.Name())
			newPath := filepath.Join(targetDir, entry.Name())
			if err := os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("failed to move %s to %s: %w", oldPath, newPath, err)
			}
		}

		if err := os.RemoveAll(subDirPath); err != nil {
			return fmt.Errorf("failed to remove empty subdirectory %s: %w", subDirPath, err)
		}
	}

	return nil
}

// GPTKVersion returns the Game Porting Toolkit version of the engine at
// targetDir (e.g. "3.0"), read from the D3DMetal.framework Info.plist.
//
// The wine64 --version banner is deliberately NOT used: Apple's wine fork
// hardcodes "wine-7.7 (Game Porting Toolkit 1.1)" in all releases, so the
// banner is meaningless for GPTK versioning. D3DMetal.framework ships a
// correct CFBundleShortVersionString (e.g. "3.0").
func GPTKVersion(targetDir string) (string, error) {
	if targetDir == "" {
		return "", fmt.Errorf("GPTKVersion: empty target dir")
	}
	plistPath := filepath.Join(targetDir, "lib", "external", "D3DMetal.framework", "Versions", "A", "Resources", "Info.plist")
	if _, err := os.Stat(plistPath); err != nil {
		return "", fmt.Errorf("GPTKVersion: no D3DMetal framework at %s: %w", plistPath, err)
	}
	out, err := exec.Command("/usr/bin/plutil", "-extract", "CFBundleShortVersionString", "raw", plistPath).Output()
	if err != nil {
		return "", fmt.Errorf("GPTKVersion: plutil failed on %s: %w", plistPath, err)
	}
	ver := strings.TrimSpace(string(out))
	if ver == "" {
		return "", fmt.Errorf("GPTKVersion: empty CFBundleShortVersionString in %s", plistPath)
	}
	return ver, nil
}

// GPTKVersionAtLeast reports whether the GPTK engine at targetDir is at
// least the given major.minor (e.g. 3.0). Missing or unparseable engines
// return false.
func GPTKVersionAtLeast(targetDir string, minMajor float64) bool {
	ver, err := GPTKVersion(targetDir)
	if err != nil {
		return false
	}
	// Version token like "3.0-3" or "1.1"; compare the leading numeric part.
	parts := strings.SplitN(ver, "-", 2)
	majorMinor, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return false
	}
	return majorMinor >= minMajor
}

// EnsureGPTK downloads and extracts the Game Porting Toolkit into targetDir.
// The Gcenx release ships as a .app bundle containing D3DMetal.framework and d3dmetal.dll.
func (m *EngineManager) EnsureGPTK(targetDir string) error {
	return m.EnsureGPTKProgress(targetDir, nil)
}

// EnsureGPTKProgress is EnsureGPTK with a download progress callback.
func (m *EngineManager) EnsureGPTKProgress(targetDir string, onProgress func(downloaded, total int64)) error {
	asset := gptkEngineAsset()
	if err := VerifyEngineInstall(targetDir, asset.verifyRelPath); err == nil {
		// Fail-proof: a file-existence check is NOT enough. GPTK 1.x
		// (Wine 7.7) passes VerifyEngineInstall but is far too old to run
		// the Steam client's Chromium 126 (webhelper crash loop). Only a
		// modern GPTK (>= 3.0) is trusted; anything older must be replaced.
		if GPTKVersionAtLeast(targetDir, 3.0) {
			return nil
		}
		fmt.Printf("[EnsureGPTK] stale GPTK install at %s (< 3.0), reinstalling\n", targetDir)
	}

	// Download and extract to a temp location first
	tmpDir := targetDir + fmt.Sprintf("_tmp_%d", time.Now().UnixNano())
	os.RemoveAll(tmpDir)
	defer os.RemoveAll(tmpDir)

	// Fetch the latest dynamic URL from GitHub (falls back to a verified URL)
	gptkURL := asset.resolveURL()
	fmt.Printf("[EnsureGPTK] Downloading %s to %s...\n", gptkURL, targetDir)

	if err := m.downloadAndExtract(gptkURL, tmpDir, onProgress); err != nil {
		return fmt.Errorf("failed to download GPTK: %w", err)
	}

	// The tarball extracts as a .app bundle. Walk to find the wine/ directory
	// which contains bin/, lib/, and share/ with D3DMetal support.
	var wineRoot string
	filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == "wine" {
			wineRoot = path
			return filepath.SkipDir
		}
		return nil
	})

	if wineRoot == "" {
		return fmt.Errorf("could not find wine/ in GPTK bundle")
	}

	// Copy EVERYTHING from wineRoot to targetDir/ (bin, lib, share)
	os.RemoveAll(targetDir)
	os.MkdirAll(targetDir, 0755)

	// We must use cp -a to copy the contents cleanly and preserve extended attributes / code signatures
	cpCmd := exec.Command("cp", "-a", wineRoot+"/", targetDir+"/")
	if err := cpCmd.Run(); err != nil {
		return fmt.Errorf("failed to copy wine dir: %w", err)
	}

	// Fail-proof: the extracted bundle must actually contain a usable wine64.
	if err := VerifyEngineInstall(targetDir, asset.verifyRelPath); err != nil {
		return fmt.Errorf("GPTK extraction incomplete: %w", err)
	}

	// Strip quarantine flags from extracted GPTK binaries and D3DMetal frameworks
	_ = exec.Command("xattr", "-cr", targetDir).Run()

	fmt.Printf("[EnsureGPTK] GPTK extracted to %s (bin/wine64 + D3DMetal.framework)\n", targetDir)
	return nil
}

// CheckGPTKInstalled returns true if the D3DMetal framework is present on the system.
func CheckGPTKInstalled() bool {
	// Check brew cask location first
	if _, err := os.Stat(GPTKBrewPath); err == nil {
		return true
	}
	// Check Gcenx brew cask cellar
	home, _ := os.UserHomeDir()
	altPaths := []string{
		filepath.Join(home, ".rift", "engines", "gptk", "lib", "external", "D3DMetal.framework", "Versions", "A", "D3DMetal"),
		filepath.Join(home, "Library", "Application Support", "MacBridge", "Tools", "gptk_temp", "Game Porting Toolkit.app", "Contents", "Resources", "wine", "lib", "external", "D3DMetal.framework", "Versions", "A", "D3DMetal"),
	}
	for _, p := range altPaths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

// GetGPTKLibPath returns the path to the D3DMetal framework binary if installed.
func GetGPTKLibPath() string {
	if _, err := os.Stat(GPTKBrewPath); err == nil {
		return GPTKBrewPath
	}
	home, _ := os.UserHomeDir()
	altPaths := []string{
		filepath.Join(home, ".rift", "engines", "gptk", "lib", "external", "D3DMetal.framework", "Versions", "A", "D3DMetal"),
		filepath.Join(home, "Library", "Application Support", "MacBridge", "Tools", "gptk_temp", "Game Porting Toolkit.app", "Contents", "Resources", "wine", "lib", "external", "D3DMetal.framework", "Versions", "A", "D3DMetal"),
	}
	for _, p := range altPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// GetGPTKDLLPath returns the path to d3dmetal.dll in the GPTK bundle if available.
func GetGPTKDLLPath() string {
	home, _ := os.UserHomeDir()
	paths := []string{
		"/Applications/Game Porting Toolkit.app/Contents/Resources/wine/lib/wine/x86_64-windows/d3dmetal.dll",
		filepath.Join(home, ".rift", "engines", "gptk", "lib", "wine", "x86_64-windows", "d3dmetal.dll"),
		filepath.Join(home, "Applications", "Game Porting Toolkit.app", "Contents", "Resources", "wine", "lib", "wine", "x86_64-windows", "d3dmetal.dll"),
	}

	// Also search standard Homebrew paths using Glob
	brewPaths := []string{
		"/opt/homebrew/Caskroom/game-porting-toolkit/*/Game Porting Toolkit.app/Contents/Resources/wine/lib/wine/x86_64-windows/d3dmetal.dll",
		"/usr/local/Caskroom/game-porting-toolkit/*/Game Porting Toolkit.app/Contents/Resources/wine/lib/wine/x86_64-windows/d3dmetal.dll",
	}
	for _, bp := range brewPaths {
		if matches, err := filepath.Glob(bp); err == nil && len(matches) > 0 {
			paths = append(paths, matches[0])
		}
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// CheckDXVKInstalled returns true if DXVK is present in the engines directory.
func CheckDXVKInstalled() bool {
	home, _ := os.UserHomeDir()
	dxvkDir := filepath.Join(home, ".rift", "engines", "dxvk")
	x64file := filepath.Join(dxvkDir, "x64", "d3d11.dll")
	if _, err := os.Stat(x64file); err == nil {
		return true
	}
	return false
}

// CheckWineInstalled returns true if the wine binary is present.
func CheckWineInstalled() bool {
	home, _ := os.UserHomeDir()
	wineBin := filepath.Join(home, ".rift", "engines", "wine", "bin", "wine")
	if _, err := os.Stat(wineBin); err == nil {
		return true
	}
	return false
}

var isSonoma *bool

// IsSonomaOrLater checks if the macOS host version is 14.0 (Sonoma) or newer.
func IsSonomaOrLater() bool {
	if isSonoma != nil {
		return *isSonoma
	}
	res := false
	cmd := exec.Command("sw_vers", "-productVersion")
	out, err := cmd.Output()
	if err == nil {
		verStr := strings.TrimSpace(string(out))
		parts := strings.Split(verStr, ".")
		if len(parts) > 0 {
			var major int
			if _, errScan := fmt.Sscanf(parts[0], "%d", &major); errScan == nil && major >= 14 {
				res = true
			}
		}
	}
	isSonoma = &res
	return res
}

// EnsureWinetricks downloads the winetricks script if it doesn't exist.
func (m *EngineManager) EnsureWinetricks(sharedWineDir string) error {
	binDir := filepath.Join(sharedWineDir, "bin")
	os.MkdirAll(binDir, 0755)
	wtPath := filepath.Join(binDir, "winetricks")

	if _, err := os.Stat(wtPath); err == nil {
		return nil
	}

	fmt.Println("[EngineManager] Downloading Winetricks...")
	url := "https://raw.githubusercontent.com/Winetricks/winetricks/master/src/winetricks"
	err := NewDownloader().DownloadFile(url, wtPath, true)
	if err != nil {
		return err
	}

	return nil
}
