package rift

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

// Application version and release metadata
const (
	AppVersion          = "1.0.0-beta.2"
	AppBuildStage       = "Beta 2"
	MinSupportedVersion = "1.0.0"
	GitHubRepo          = "Luciferjimmy/RIFT"
)

// UpdateInfo holds details about available application updates
type UpdateInfo struct {
	UpdateAvailable bool   `json:"updateAvailable"`
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	ReleaseNotes    string `json:"releaseNotes"`
	DownloadURL     string `json:"downloadUrl"`
	PublishedAt     string `json:"publishedAt"`
}

// GetAppVersion returns the current client version string (e.g. "1.0.0")
func (a *App) GetAppVersion() string {
	return AppVersion
}

// GetAppBuildInfo returns comprehensive build and platform information
func (a *App) GetAppBuildInfo() map[string]string {
	return map[string]string{
		"version":      AppVersion,
		"stage":        AppBuildStage,
		"platform":     "darwin",
		"architecture": "arm64",
		"engine":       "Consolidated Translation Architecture",
	}
}

// CheckForUpdates queries GitHub Releases to determine if a newer version of RIFT is available
func (a *App) CheckForUpdates() UpdateInfo {
	info := UpdateInfo{
		UpdateAvailable: false,
		CurrentVersion:  AppVersion,
		LatestVersion:   AppVersion,
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GitHubRepo)
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return info
	}
	req.Header.Set("User-Agent", "RIFT-Desktop/"+AppVersion)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return info
	}
	defer resp.Body.Close()

	var release struct {
		TagName     string `json:"tag_name"`
		Body        string `json:"body"`
		PublishedAt string `json:"published_at"`
		HTMLURL     string `json:"html_url"`
		Assets      []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return info
	}

	latestClean := strings.TrimPrefix(strings.TrimSpace(release.TagName), "v")
	info.LatestVersion = latestClean
	info.ReleaseNotes = release.Body
	info.PublishedAt = release.PublishedAt
	info.DownloadURL = release.HTMLURL

	// Find direct DMG or ZIP asset if available
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, ".dmg") || strings.HasSuffix(asset.Name, ".zip") {
			info.DownloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if isVersionNewer(latestClean, AppVersion) {
		info.UpdateAvailable = true
	}

	return info
}

// DownloadAndApplyUpdate downloads an update asset with progress tracking and opens it for installation
func (a *App) DownloadAndApplyUpdate(downloadURL string) error {
	if downloadURL == "" {
		return fmt.Errorf("empty download URL")
	}

	// If not a direct binary asset (.dmg / .zip / .pkg), open in default browser
	lower := strings.ToLower(downloadURL)
	if !strings.HasSuffix(lower, ".dmg") && !strings.HasSuffix(lower, ".zip") && !strings.HasSuffix(lower, ".pkg") {
		_ = exec.Command("open", downloadURL).Start()
		return nil
	}

	go func() {
		home, err := os.UserHomeDir()
		if err != nil {
			a.emitEvent("update_failed", "Cannot find home directory")
			return
		}
		filename := filepath.Base(downloadURL)
		if idx := strings.Index(filename, "?"); idx != -1 {
			filename = filename[:idx]
		}
		destPath := filepath.Join(home, "Downloads", filename)

		resp, err := http.Get(downloadURL)
		if err != nil {
			a.emitEvent("update_failed", err.Error())
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			a.emitEvent("update_failed", fmt.Sprintf("Download failed: HTTP %d", resp.StatusCode))
			return
		}

		out, err := os.Create(destPath)
		if err != nil {
			a.emitEvent("update_failed", err.Error())
			return
		}
		defer out.Close()

		totalBytes := resp.ContentLength
		var downloaded int64
		buf := make([]byte, 32*1024)

		for {
			n, readErr := resp.Body.Read(buf)
			if n > 0 {
				_, writeErr := out.Write(buf[:n])
				if writeErr != nil {
					a.emitEvent("update_failed", writeErr.Error())
					return
				}
				downloaded += int64(n)
				percent := 0.0
				if totalBytes > 0 {
					percent = float64(downloaded) / float64(totalBytes) * 100
				}
				a.emitEvent("update_progress", map[string]interface{}{
					"percent":    percent,
					"downloaded": downloaded,
					"total":      totalBytes,
				})
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				a.emitEvent("update_failed", readErr.Error())
				return
			}
		}

		a.emitEvent("update_completed", destPath)
		// Automatically open installer on macOS
		_ = exec.Command("open", destPath).Start()
	}()

	return nil
}

// isVersionNewer compares two semver strings (e.g. "1.1.0" > "1.0.0", handles "1.0.0-beta.2" vs "1.0.0-beta.1")
func isVersionNewer(latest, current string) bool {
	if latest == current {
		return false
	}

	splitPre := func(v string) (string, string) {
		if idx := strings.Index(v, "-"); idx != -1 {
			return v[:idx], v[idx+1:]
		}
		return v, ""
	}

	lBase, lPre := splitPre(latest)
	cBase, cPre := splitPre(current)

	lParts := strings.Split(lBase, ".")
	cParts := strings.Split(cBase, ".")

	for i := 0; i < len(lParts) && i < len(cParts); i++ {
		lNum, _ := strconv.Atoi(lParts[i])
		cNum, _ := strconv.Atoi(cParts[i])
		if lNum > cNum {
			return true
		}
		if lNum < cNum {
			return false
		}
	}
	if len(lParts) != len(cParts) {
		return len(lParts) > len(cParts)
	}

	// Base versions are identical:
	// A release version (empty pre) is newer than a pre-release version
	if lPre == "" && cPre != "" {
		return true
	}
	if lPre != "" && cPre == "" {
		return false
	}
	return comparePrereleases(lPre, cPre) > 0
}

// comparePrereleases compares two semver prerelease strings (e.g. "beta.2" vs "beta.10")
// Returns 1 if lPre > cPre, -1 if lPre < cPre, 0 if equal.
func comparePrereleases(lPre, cPre string) int {
	if lPre == cPre {
		return 0
	}
	lParts := strings.Split(lPre, ".")
	cParts := strings.Split(cPre, ".")
	for i := 0; i < len(lParts) && i < len(cParts); i++ {
		lp, cp := lParts[i], cParts[i]
		if lp == cp {
			continue
		}
		lNum, lErr := strconv.Atoi(lp)
		cNum, cErr := strconv.Atoi(cp)
		if lErr == nil && cErr == nil {
			if lNum > cNum {
				return 1
			}
			if lNum < cNum {
				return -1
			}
		} else if lErr == nil && cErr != nil {
			return -1
		} else if lErr != nil && cErr == nil {
			return 1
		} else {
			if lp > cp {
				return 1
			}
			if lp < cp {
				return -1
			}
		}
	}
	if len(lParts) > len(cParts) {
		return 1
	}
	if len(lParts) < len(cParts) {
		return -1
	}
	return 0
}

