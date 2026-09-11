package runners

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"rift/internal/engine"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

// LegendaryErrorType categorizes Legendary CLI failures for proper handling upstream.
type LegendaryErrorType int

const (
	LegendaryErrGeneric LegendaryErrorType = iota
	LegendaryErrAuthExpired
	LegendaryErrNetwork
	LegendaryErrDiskSpace
	LegendaryErrCancelled
	LegendaryErrManifestNotFound
)

// LegendaryError is a structured error returned by Legendary operations.
type LegendaryError struct {
	Type    LegendaryErrorType
	Message string
	RawLog  string
}

func (e *LegendaryError) Error() string {
	return e.Message
}

// ClassifyLegendaryOutput analyzes raw Legendary output and returns a structured error.
func ClassifyLegendaryOutput(output string) *LegendaryError {
	lower := strings.ToLower(output)

	// Auth / session failures
	if strings.Contains(lower, "403 client error") ||
		strings.Contains(lower, "no saved credentials") ||
		strings.Contains(lower, "log in failed") ||
		strings.Contains(lower, "login failed") ||
		strings.Contains(lower, "invalid credentials") ||
		strings.Contains(lower, "session expired") ||
		strings.Contains(lower, "token is expired") ||
		strings.Contains(lower, "refresh token is invalid") ||
		strings.Contains(lower, "could not refresh") ||
		strings.Contains(output, "HTTP request for login failed") {
		return &LegendaryError{
			Type:    LegendaryErrAuthExpired,
			Message: "Your Epic Games login session has expired. Please re-login in Settings.",
			RawLog:  output,
		}
	}

	// Network / DNS / Host unreachable
	if strings.Contains(lower, "connectionerror") ||
		strings.Contains(lower, "nameresolutionerror") ||
		strings.Contains(lower, "failed to resolve") ||
		strings.Contains(lower, "max retries exceeded") ||
		strings.Contains(lower, "network unreachable") ||
		strings.Contains(lower, "connection timed out") ||
		strings.Contains(lower, "could not connect to host") ||
		strings.Contains(lower, "nodename nor servname provided") {
		return &LegendaryError{
			Type:    LegendaryErrNetwork,
			Message: "Unable to connect to Epic Games servers. Please check your internet connection.",
			RawLog:  output,
		}
	}

	// Disk space
	if strings.Contains(lower, "no space left on device") ||
		strings.Contains(lower, "not enough disk space") ||
		strings.Contains(lower, "disk is full") ||
		strings.Contains(lower, "errno 28") {
		return &LegendaryError{
			Type:    LegendaryErrDiskSpace,
			Message: "Not enough disk space to download this game. Please free up some storage.",
			RawLog:  output,
		}
	}

	// Manifest / App not found
	if strings.Contains(lower, "manifest not found") ||
		strings.Contains(lower, "app not found") ||
		strings.Contains(lower, "does not have a manifest") {
		return &LegendaryError{
			Type:    LegendaryErrManifestNotFound,
			Message: "Game manifest was not found on Epic Games servers.",
			RawLog:  output,
		}
	}

	// Cancellation
	if strings.Contains(lower, "signal: killed") ||
		strings.Contains(lower, "context canceled") ||
		strings.Contains(lower, "sigkill") {
		return &LegendaryError{
			Type:    LegendaryErrCancelled,
			Message: "Download was cancelled.",
			RawLog:  output,
		}
	}

	return &LegendaryError{
		Type:    LegendaryErrGeneric,
		Message: output,
		RawLog:  output,
	}
}

// LegendaryRunner handles executing commands using a specific legendary binary.
type LegendaryRunner struct {
	BinaryPath    string
	GlobalAuthDir string // Shared globally so the user only logs in once
}

// NewLegendaryRunner creates a runner for the legendary CLI.
func NewLegendaryRunner(binaryPath string, globalAuthDir string) *LegendaryRunner {
	return &LegendaryRunner{
		BinaryPath:    binaryPath,
		GlobalAuthDir: globalAuthDir,
	}
}

// execute runs a legendary command and returns stdout.
func (r *LegendaryRunner) execute(args ...string) (string, error) {
	cmd := exec.Command(r.BinaryPath, args...)
	cmd.Env = append(engine.SafeEnviron(), "LEGENDARY_CONFIG_PATH="+r.GlobalAuthDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("legendary command failed: %w\nStderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// executeCombined runs a command and returns both stdout and stderr merged.
// Essential for 'status' command which outputs entirely to stderr.
func (r *LegendaryRunner) executeCombined(args ...string) (string, error) {
	cmd := exec.Command(r.BinaryPath, args...)
	cmd.Env = append(engine.SafeEnviron(), "LEGENDARY_CONFIG_PATH="+r.GlobalAuthDir)

	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("legendary command failed: %w\nOutput: %s", err, combined.String())
	}

	return combined.String(), nil
}

// AuthWithCode securely logs in using an Epic Games authorization code.
func (r *LegendaryRunner) AuthWithCode(code string) (string, error) {
	return r.execute("auth", "--code", code)
}

// AuthWithSID securely logs in using the captured Epic Games session ID.
func (r *LegendaryRunner) AuthWithSID(sid string) (string, error) {
	return r.execute("auth", "--sid", sid)
}

// GetUsername queries legendary status and extracts the Epic Games display name.
func (r *LegendaryRunner) GetUsername() (string, error) {
	out, err := r.executeCombined("status")
	if err != nil {
		return "", err
	}

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Epic account: ") {
			userPart := strings.TrimPrefix(line, "Epic account: ")
			if strings.Contains(userPart, "<not logged in>") {
				return "", fmt.Errorf("not logged in")
			}

			// Extract just the username before the (Account ID)
			parts := strings.Split(userPart, " (Account ID:")
			if len(parts) > 0 {
				return strings.TrimSpace(parts[0]), nil
			}
		}
	}
	return "", fmt.Errorf("could not parse username from legendary status")
}

// GetUsernameFast reads the user.json file directly from the config directory.
func (r *LegendaryRunner) GetUsernameFast() (string, error) {
	userFile := filepath.Join(r.GlobalAuthDir, "user.json")
	data, err := os.ReadFile(userFile)
	if err != nil {
		return "", fmt.Errorf("no user.json found: %w", err)
	}

	var userData map[string]interface{}
	if err := json.Unmarshal(data, &userData); err != nil {
		return "", fmt.Errorf("failed to parse user.json: %w", err)
	}

	if displayName, ok := userData["displayName"].(string); ok && displayName != "" {
		return displayName, nil
	}

	if displayName, ok := userData["display_name"].(string); ok && displayName != "" {
		return displayName, nil
	}

	return "", fmt.Errorf("no display name found in user.json")
}

// ImportExistingAuth copies the user's existing Legendary credentials.
func (r *LegendaryRunner) ImportExistingAuth() error {
	home, _ := os.UserHomeDir()
	defaultConfigDir := filepath.Join(home, ".config", "legendary")
	srcUserJSON := filepath.Join(defaultConfigDir, "user.json")

	if _, err := os.Stat(srcUserJSON); err != nil {
		return fmt.Errorf("no existing Legendary credentials found at %s", srcUserJSON)
	}

	data, err := os.ReadFile(srcUserJSON)
	if err != nil {
		return fmt.Errorf("failed to read existing user.json: %w", err)
	}

	if err := os.MkdirAll(r.GlobalAuthDir, 0755); err != nil {
		return fmt.Errorf("failed to create RIFT auth directory: %w", err)
	}

	destUserJSON := filepath.Join(r.GlobalAuthDir, "user.json")
	if err := os.WriteFile(destUserJSON, data, 0600); err != nil {
		return fmt.Errorf("failed to write user.json to RIFT config: %w", err)
	}

	srcConfig := filepath.Join(defaultConfigDir, "config.ini")
	if data, err := os.ReadFile(srcConfig); err == nil {
		os.WriteFile(filepath.Join(r.GlobalAuthDir, "config.ini"), data, 0600)
	}

	fmt.Println("[Legendary] Successfully imported existing Epic Games credentials.")
	return nil
}

// EpicGameEntry represents a single game from Legendary's --json output.
type EpicGameEntry struct {
	AppName  string `json:"app_name"`
	AppTitle string `json:"app_title"`
}

// ListGames calls `legendary list --platform Windows --json` and returns the parsed game list.
func (r *LegendaryRunner) ListGames() ([]EpicGameEntry, error) {
	out, err := r.execute("list", "--platform", "Windows", "--json")
	if err != nil {
		return nil, err
	}

	var games []EpicGameEntry
	if err := json.Unmarshal([]byte(out), &games); err != nil {
		return nil, fmt.Errorf("failed to parse legendary game list: %w", err)
	}

	return games, nil
}

// List returns the list of owned Epic games in JSON format.
func (r *LegendaryRunner) List() (string, error) {
	return r.execute("list", "--platform", "Windows", "--json")
}

// Install initiates the installation of a Windows Epic game.
func (r *LegendaryRunner) Install(appName string, installPath string) (string, error) {
	return r.execute("install", appName, "--base-path", installPath, "-y", "--skip-sdl", "--skip-dlcs")
}

// InstallWithContext runs legendary install with a cancellable context and real-time progress callback.
// progressFn receives (percentComplete, speedMBps, downloadedMB, totalMB). Returns the *exec.Cmd so caller can track the process.
// If configPath is non-empty it overrides the runner's GlobalAuthDir for LEGENDARY_CONFIG_PATH.
func (r *LegendaryRunner) InstallWithContext(ctx context.Context, appName, installPath, configPath string, progressFn func(percent, speed, downloaded, total float64), authErrorFn func()) (*exec.Cmd, error) {
	cfgDir := configPath
	if cfgDir == "" {
		cfgDir = r.GlobalAuthDir
	}
	cmd := exec.CommandContext(ctx, r.BinaryPath, "install", appName, "--base-path", installPath, "-y", "--skip-sdl", "--skip-dlcs", "--max-workers", "16")
	cmd.Stdin = strings.NewReader("\n\n")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process != nil {
			return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		return nil
	}
	cmd.Env = append(engine.SafeEnviron(), "LEGENDARY_CONFIG_PATH="+cfgDir)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return cmd, fmt.Errorf("stdout pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout // merge stderr into stdout for parsing

	if err := cmd.Start(); err != nil {
		return cmd, fmt.Errorf("start legendary: %w", err)
	}

	// When running without a TTY, Legendary outputs a multi-line verbose format:
	// [DLManager] INFO: = Progress: 15.30% (5/26456), Running for 00:00:05...
	// [DLManager] INFO:  - Downloaded: 10.50 MiB, Written: 4.01 MiB
	// [DLManager] INFO:  - Cache usage: 33.00 MiB, active tasks: 32
	// [DLManager] INFO:  + Download	- 2.50 MiB/s (raw) / 0.99 MiB/s (decompressed)
	// [DLManager] INFO:  + Disk	- 3.98 MiB/s (write) / 0.00 MiB/s (read)

	progressRe := regexp.MustCompile(`=\s*Progress:\s*([\d.]+)%`)
	downloadedRe := regexp.MustCompile(`-\s*Downloaded:\s*([\d.]+)\s*([KMGT]?i?B)`)
	speedRe := regexp.MustCompile(`\+\s*Download\s*-\s*([\d.]+)\s*([KMGT]?i?B/s)`)
	totalSizeRe := regexp.MustCompile(`(?i)download size:\s*([\d.]+)\s*([KMGT]?i?B)`)

	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 1024*64), 1024*64) // 64KB buffer

		var curPercent, curSpeed, curDown, curTotal float64

		for scanner.Scan() {
			line := scanner.Text()

			if strings.Contains(line, "403 Client Error") || strings.Contains(line, "No saved credentials") || strings.Contains(line, "Log in failed") {
				if authErrorFn != nil {
					authErrorFn()
				}
				if cmd.Process != nil {
					syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
				}
				continue
			}

			if match := progressRe.FindStringSubmatch(line); match != nil {
				curPercent, _ = strconv.ParseFloat(match[1], 64)
			} else if match := downloadedRe.FindStringSubmatch(line); match != nil {
				val, _ := strconv.ParseFloat(match[1], 64)
				unit := match[2]
				if strings.HasPrefix(unit, "G") {
					val *= 1024
				}
				if strings.HasPrefix(unit, "K") {
					val /= 1024
				}
				curDown = val
			} else if match := speedRe.FindStringSubmatch(line); match != nil {
				val, _ := strconv.ParseFloat(match[1], 64)
				unit := match[2]
				if strings.HasPrefix(unit, "G") {
					val *= 1024
				}
				if strings.HasPrefix(unit, "K") {
					val /= 1024
				}
				curSpeed = val

				// Speed is the 4th line in the block, emit progress here
				if progressFn != nil {
					progressFn(curPercent, curSpeed, curDown, curTotal)
				}
			} else if match := totalSizeRe.FindStringSubmatch(line); match != nil {
				val, _ := strconv.ParseFloat(match[1], 64)
				unit := match[2]
				if strings.HasPrefix(strings.ToUpper(unit), "G") {
					val *= 1024
				}
				if strings.HasPrefix(strings.ToUpper(unit), "K") {
					val /= 1024
				}
				curTotal = val
			}
		}
		// Consume remaining output to avoid pipe deadlock
		io.Copy(io.Discard, stdout)
	}()

	return cmd, nil
}

// InstallMac initiates the installation of a Mac Epic game explicitly.
func (r *LegendaryRunner) InstallMac(appName string, installPath string) (string, error) {
	return r.execute("install", appName, "--platform", "Mac", "--base-path", installPath, "-y", "--skip-sdl", "--skip-dlcs")
}

// InstallMacWithContext runs legendary install with a cancellable context, explicitly setting --platform Mac, and streaming progress.
func (r *LegendaryRunner) InstallMacWithContext(ctx context.Context, appName, installPath, configPath string, progressFn func(percent, speed, downloaded, total float64), authErrorFn func()) (*exec.Cmd, error) {
	cfgDir := configPath
	if cfgDir == "" {
		cfgDir = r.GlobalAuthDir
	}
	cmd := exec.CommandContext(ctx, r.BinaryPath, "install", appName, "--platform", "Mac", "--base-path", installPath, "-y", "--skip-sdl", "--skip-dlcs", "--max-workers", "16")
	cmd.Stdin = strings.NewReader("\n\n")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Env = append(engine.SafeEnviron(), "LEGENDARY_CONFIG_PATH="+cfgDir)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return cmd, fmt.Errorf("stdout pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout // merge stderr into stdout for parsing

	if err := cmd.Start(); err != nil {
		return cmd, fmt.Errorf("start legendary: %w", err)
	}

	progressRe := regexp.MustCompile(`=\s*Progress:\s*([\d.]+)%`)
	downloadedRe := regexp.MustCompile(`-\s*Downloaded:\s*([\d.]+)\s*([KMGT]?i?B)`)
	speedRe := regexp.MustCompile(`\+\s*Download\s*-\s*([\d.]+)\s*([KMGT]?i?B/s)`)
	totalSizeRe := regexp.MustCompile(`(?i)download size:\s*([\d.]+)\s*([KMGT]?i?B)`)

	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 1024*64), 1024*64)

		var curPercent, curSpeed, curDown, curTotal float64

		for scanner.Scan() {
			line := scanner.Text()

			if strings.Contains(line, "403 Client Error") || strings.Contains(line, "No saved credentials") || strings.Contains(line, "Log in failed") {
				if authErrorFn != nil {
					authErrorFn()
				}
				if cmd.Process != nil {
					cmd.Process.Kill()
				}
				continue
			}

			if match := progressRe.FindStringSubmatch(line); match != nil {
				curPercent, _ = strconv.ParseFloat(match[1], 64)
			} else if match := downloadedRe.FindStringSubmatch(line); match != nil {
				val, _ := strconv.ParseFloat(match[1], 64)
				unit := match[2]
				if strings.HasPrefix(unit, "G") {
					val *= 1024
				}
				if strings.HasPrefix(unit, "K") {
					val /= 1024
				}
				curDown = val
			} else if match := speedRe.FindStringSubmatch(line); match != nil {
				val, _ := strconv.ParseFloat(match[1], 64)
				unit := match[2]
				if strings.HasPrefix(unit, "G") {
					val *= 1024
				}
				if strings.HasPrefix(unit, "K") {
					val /= 1024
				}
				curSpeed = val

				if progressFn != nil {
					progressFn(curPercent, curSpeed, curDown, curTotal)
				}
			} else if match := totalSizeRe.FindStringSubmatch(line); match != nil {
				val, _ := strconv.ParseFloat(match[1], 64)
				unit := match[2]
				if strings.HasPrefix(strings.ToUpper(unit), "G") {
					val *= 1024
				}
				if strings.HasPrefix(strings.ToUpper(unit), "K") {
					val /= 1024
				}
				curTotal = val
			}
		}
		io.Copy(io.Discard, stdout)
	}()

	return cmd, nil
}

// SelectivePack represents an optional package, DLC, or HD texture pack in an Epic game.
type SelectivePack struct {
	Tag         string `json:"tag"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsRequired  bool   `json:"isRequired"`
}

// GetSelectivePacks queries Epic/Legendary for optional selective download packs (e.g. HD Textures, DLCs).
func (r *LegendaryRunner) GetSelectivePacks(ctx context.Context, appName, configPath string) ([]SelectivePack, error) {
	cfgDir := configPath
	if cfgDir == "" {
		cfgDir = r.GlobalAuthDir
	}

	pyScript := fmt.Sprintf(`
import json, sys
try:
    from legendary.core import LegendaryCore
    from legendary.cli import get_sdl_appname
    core = LegendaryCore()
    core.login()
    sdl_name = get_sdl_appname('%s')
    sdl_data = core.get_sdl_data(sdl_name, platform='Windows') if sdl_name else None
    packs = []
    if sdl_data:
        for tag, info in sdl_data.items():
            if tag != '__required':
                packs.append({'tag': tag, 'name': info.get('name', tag), 'description': info.get('description', '')})
    print("__JSON_START__" + json.dumps(packs))
except Exception as e:
    print("__JSON_START__[]")
`, appName)

	cmd := exec.CommandContext(ctx, "python3", "-c", pyScript)
	cmd.Env = append(engine.SafeEnviron(), "LEGENDARY_CONFIG_PATH="+cfgDir)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, nil // non-fatal fallback
	}

	outStr := string(out)
	idx := strings.Index(outStr, "__JSON_START__")
	if idx == -1 {
		return nil, nil
	}

	jsonStr := strings.TrimSpace(outStr[idx+len("__JSON_START__"):])
	var packs []SelectivePack
	if err := json.Unmarshal([]byte(jsonStr), &packs); err != nil {
		return nil, nil
	}

	return packs, nil
}

// DownloadEpicApp runs legendary install synchronously with full error classification, progress streaming, and output capture.
func (r *LegendaryRunner) DownloadEpicApp(ctx context.Context, appName, installPath, configPath string, isMac bool, installTags []string, progressFn func(percent, speed, downloaded, total float64)) error {
	cfgDir := configPath
	if cfgDir == "" {
		cfgDir = r.GlobalAuthDir
	}
	args := []string{"install", appName, "--base-path", installPath, "-y", "--max-workers", "16"}
	if isMac {
		args = []string{"install", appName, "--platform", "Mac", "--base-path", installPath, "-y", "--max-workers", "16"}
	}

	if len(installTags) > 0 {
		for _, tag := range installTags {
			trimmed := strings.TrimSpace(tag)
			if trimmed != "" {
				args = append(args, "--install-tag", trimmed)
			}
		}
	} else {
		args = append(args, "--skip-sdl", "--skip-dlcs")
	}

	cmd := exec.CommandContext(ctx, r.BinaryPath, args...)
	cmd.Stdin = strings.NewReader("\n\n")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process != nil {
			return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		return nil
	}
	cmd.Env = append(engine.SafeEnviron(), "LEGENDARY_CONFIG_PATH="+cfgDir)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout // merge stderr into stdout for parsing

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start legendary: %w", err)
	}

	progressRe := regexp.MustCompile(`=\s*Progress:\s*([\d.]+)%`)
	downloadedRe := regexp.MustCompile(`-\s*Downloaded:\s*([\d.]+)\s*([KMGT]?i?B)`)
	speedRe := regexp.MustCompile(`\+\s*Download\s*-\s*([\d.]+)\s*([KMGT]?i?B/s)`)
	totalSizeRe := regexp.MustCompile(`(?i)download size:\s*([\d.]+)\s*([KMGT]?i?B)`)

	var logLines []string
	var logMu sync.Mutex

	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 1024*64), 1024*64)

		var curPercent, curSpeed, curDown, curTotal float64

		for scanner.Scan() {
			line := scanner.Text()
			logMu.Lock()
			if len(logLines) > 100 {
				logLines = logLines[1:]
			}
			logLines = append(logLines, line)
			logMu.Unlock()

			if match := progressRe.FindStringSubmatch(line); match != nil {
				curPercent, _ = strconv.ParseFloat(match[1], 64)
			} else if match := downloadedRe.FindStringSubmatch(line); match != nil {
				val, _ := strconv.ParseFloat(match[1], 64)
				unit := match[2]
				if strings.HasPrefix(unit, "G") {
					val *= 1024
				}
				if strings.HasPrefix(unit, "K") {
					val /= 1024
				}
				curDown = val
			} else if match := speedRe.FindStringSubmatch(line); match != nil {
				val, _ := strconv.ParseFloat(match[1], 64)
				unit := match[2]
				if strings.HasPrefix(unit, "G") {
					val *= 1024
				}
				if strings.HasPrefix(unit, "K") {
					val /= 1024
				}
				curSpeed = val

				if progressFn != nil {
					progressFn(curPercent, curSpeed, curDown, curTotal)
				}
			} else if match := totalSizeRe.FindStringSubmatch(line); match != nil {
				val, _ := strconv.ParseFloat(match[1], 64)
				unit := match[2]
				if strings.HasPrefix(strings.ToUpper(unit), "G") {
					val *= 1024
				}
				if strings.HasPrefix(strings.ToUpper(unit), "K") {
					val /= 1024
				}
				curTotal = val
			}
		}
		io.Copy(io.Discard, stdout)
	}()

	runErr := cmd.Wait()
	if runErr != nil {
		if ctx.Err() != nil {
			return &LegendaryError{Type: LegendaryErrCancelled, Message: "Download was cancelled."}
		}
		logMu.Lock()
		fullLog := strings.Join(logLines, "\n")
		logMu.Unlock()
		return ClassifyLegendaryOutput(fullLog)
	}

	return nil
}
