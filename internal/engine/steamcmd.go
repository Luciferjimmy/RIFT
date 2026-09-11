package engine

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	SteamCMDDownloadURL = "https://client-update.steamstatic.com/installer/steamcmd_osx.tar.gz"
)

type AuthResult struct {
	Success         bool   `json:"success"`
	NeedsSteamGuard bool   `json:"needsGuard"`
	NeedsEmailCode  bool   `json:"needsEmailCode"`
	ErrorMessage    string `json:"error"`
	Username        string `json:"username"`
}

type DownloadProgress struct {
	Percent    float64 `json:"percent"`
	BytesDone  int64   `json:"bytesDone"`
	BytesTotal int64   `json:"bytesTotal"`
	SpeedMBps  float64 `json:"speed"`
	Stage      string  `json:"stage"`
}

// SteamCMDErrorType categorizes SteamCMD failures for proper handling upstream.
type SteamCMDErrorType int

const (
	SteamErrGeneric        SteamCMDErrorType = iota
	SteamErrLoginExpired                     // Session expired, token invalid, need re-auth
	SteamErrNoSubscription                   // User doesn't own the game
	SteamErrDiskSpace                        // Not enough disk space
	SteamErrNetwork                          // Network/connection failure
	SteamErrRateLimit                        // Too many requests
	SteamErrCancelled                        // User cancelled
)

// SteamCMDError is a structured error returned by SteamCMD operations.
type SteamCMDError struct {
	Type    SteamCMDErrorType
	Message string // User-friendly message
	RawLog  string // Raw SteamCMD output for debugging
}

func (e *SteamCMDError) Error() string {
	return e.Message
}

// classifySteamCMDOutput analyzes raw SteamCMD output and returns a structured error.
// Returns nil if the output indicates success.
func classifySteamCMDOutput(output string) *SteamCMDError {
	lower := strings.ToLower(output)

	// Login / auth failures
	if strings.Contains(lower, "login failure") ||
		strings.Contains(lower, "invalid password") ||
		strings.Contains(lower, "failed login") ||
		strings.Contains(lower, "expired") ||
		strings.Contains(lower, "access denied") ||
		strings.Contains(lower, "logon session has expired") ||
		strings.Contains(lower, "not logged on") ||
		strings.Contains(output, "FAILED (Invalid Password)") ||
		strings.Contains(output, "FAILED (Expired Login)") {
		return &SteamCMDError{
			Type:    SteamErrLoginExpired,
			Message: "Your Steam session has expired. Please log in to Steam again.",
			RawLog:  output,
		}
	}

	// No subscription / ownership
	if strings.Contains(output, "No subscription") ||
		strings.Contains(lower, "no subscription") ||
		strings.Contains(output, "No licenses") {
		return &SteamCMDError{
			Type:    SteamErrNoSubscription,
			Message: "You do not own this game. Please purchase it or add it to your Steam library first.",
			RawLog:  output,
		}
	}

	// Disk space
	if strings.Contains(lower, "not enough disk space") ||
		strings.Contains(lower, "disk full") ||
		strings.Contains(lower, "no space left") {
		return &SteamCMDError{
			Type:    SteamErrDiskSpace,
			Message: "Not enough disk space to download this game. Please free up some storage.",
			RawLog:  output,
		}
	}

	// Network errors
	if strings.Contains(lower, "connection timed out") ||
		strings.Contains(lower, "could not connect") ||
		strings.Contains(lower, "no connection") ||
		strings.Contains(lower, "network unreachable") ||
		strings.Contains(lower, "content servers unreachable") {
		return &SteamCMDError{
			Type:    SteamErrNetwork,
			Message: "Could not connect to Steam servers. Please check your internet connection and try again.",
			RawLog:  output,
		}
	}

	// Rate limiting
	if strings.Contains(output, "Rate Limit Exceeded") ||
		strings.Contains(lower, "rate limit") ||
		strings.Contains(lower, "too many requests") {
		return &SteamCMDError{
			Type:    SteamErrRateLimit,
			Message: "Steam is rate-limiting requests. Please wait a few minutes and try again.",
			RawLog:  output,
		}
	}

	return nil
}

// IsSteamCMDReady checks if the native macOS steamcmd binary exists.
func IsSteamCMDReady(engineDir string) bool {
	binaryPath := filepath.Join(engineDir, "steamcmd.sh")
	if fi, err := os.Stat(binaryPath); err == nil && !fi.IsDir() {
		return true
	}
	binaryPath = filepath.Join(engineDir, "steamcmd")
	if fi, err := os.Stat(binaryPath); err == nil && !fi.IsDir() {
		return true
	}
	return false
}

// EnsureSteamCMD downloads, extracts, and bootstraps native macOS SteamCMD.
func EnsureSteamCMD(engineDir string) error {
	if IsSteamCMDReady(engineDir) {
		return nil
	}

	if err := os.MkdirAll(engineDir, 0755); err != nil {
		return fmt.Errorf("failed to create steamcmd dir: %w", err)
	}

	tarballPath := filepath.Join(engineDir, "steamcmd_osx.tar.gz")
	
	// Download with proper timeout and retry
	client := &http.Client{
		Timeout: 45 * time.Second,
	}
	var resp *http.Response
	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		resp, err = client.Get(SteamCMDDownloadURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(time.Duration(attempt) * time.Second)
	}
	if err != nil || resp == nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			return fmt.Errorf("steamcmd download failed with status %d: %v", resp.StatusCode, err)
		}
		return fmt.Errorf("failed to download steamcmd after 3 attempts: %v", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(tarballPath)
	if err != nil {
		return fmt.Errorf("failed to create tarball file: %w", err)
	}
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		return fmt.Errorf("failed to write steamcmd tarball: %w", err)
	}

	// Extract
	f, err := os.Open(tarballPath)
	if err != nil {
		return fmt.Errorf("failed to open tarball: %w", err)
	}
	defer f.Close()
	defer os.Remove(tarballPath)

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar reading error: %w", err)
		}

		target := filepath.Join(engineDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			baseDir := filepath.Dir(target)
			if err := os.MkdirAll(baseDir, 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, header.FileInfo().Mode())
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
			os.Chmod(target, 0755)
		}
	}

	// Strip macOS Gatekeeper quarantine attributes from extracted steamcmd binaries
	_ = exec.Command("xattr", "-cr", engineDir).Run()

	// Bootstrap run to allow SteamCMD to self-update once
	cmdScript := filepath.Join(engineDir, "steamcmd.sh")
	if _, err := os.Stat(cmdScript); err == nil {
		os.Chmod(cmdScript, 0755)
		cmd := exec.Command(cmdScript, "+quit")
		cmd.Dir = engineDir
		_ = cmd.Run() // exit code can be 7 or 0 during initial update, ignore error
	}

	return nil
}

// parseSteamAuthOutput parses SteamCMD output and returns a clean, user-friendly AuthResult.
// Raw terminal logs are never returned directly to the user interface.
func parseSteamAuthOutput(output string, username string) *AuthResult {
	lower := strings.ToLower(output)

	// 1. Success check
	if strings.Contains(output, "Logged in OK") ||
		strings.Contains(output, "Waiting for user info...OK") ||
		(strings.Contains(lower, "logged in ok") && !strings.Contains(lower, "error") && !strings.Contains(lower, "failed")) {
		return &AuthResult{
			Success:  true,
			Username: username,
		}
	}

	// 2. Steam Guard / 2FA required (First prompt)
	if (strings.Contains(lower, "steam guard") && !strings.Contains(lower, "steam guard code provided")) ||
		strings.Contains(lower, "two-factor") ||
		strings.Contains(lower, "2fa") ||
		strings.Contains(lower, "account logon denied") ||
		strings.Contains(output, "FAILED (Account Logon Denied)") ||
		strings.Contains(output, "ERROR (Account Logon Denied)") {
		return &AuthResult{
			Success:         false,
			NeedsSteamGuard: true,
			Username:        username,
			ErrorMessage:    "Steam Guard code required. Please enter the code from your Steam Mobile Authenticator or email.",
		}
	}

	// 3. Steam Guard code was provided, but is wrong or expired
	if strings.Contains(lower, "invalid login auth code") ||
		strings.Contains(output, "FAILED (Invalid Login Auth Code)") ||
		strings.Contains(output, "ERROR (Invalid Login Auth Code)") {
		return &AuthResult{
			Success:         false,
			NeedsSteamGuard: true,
			Username:        username,
			ErrorMessage:    "The Steam Guard code is invalid or expired. Please check your app or email for a fresh code.",
		}
	}

	// 4. Invalid Password / Account Name
	if strings.Contains(lower, "invalid password") ||
		strings.Contains(output, "ERROR (Invalid Password)") ||
		strings.Contains(output, "FAILED (Invalid Password)") ||
		strings.Contains(lower, "login failure") {
		return &AuthResult{
			Success:      false,
			Username:     username,
			ErrorMessage: "Incorrect Steam account name or password. Please verify your credentials and try again.",
		}
	}

	// 5. Rate Limiting
	if strings.Contains(lower, "rate limit") ||
		strings.Contains(output, "Rate Limit Exceeded") ||
		strings.Contains(output, "FAILED (Rate Limit Exceeded)") ||
		strings.Contains(output, "ERROR (Rate Limit Exceeded)") {
		return &AuthResult{
			Success:      false,
			Username:     username,
			ErrorMessage: "Steam is temporarily rate-limiting sign-in attempts. Please wait a few minutes before trying again.",
		}
	}

	// 6. Account Disabled / Suspended
	if strings.Contains(lower, "account disabled") || strings.Contains(lower, "account suspended") {
		return &AuthResult{
			Success:      false,
			Username:     username,
			ErrorMessage: "This Steam account has been disabled or locked by Valve.",
		}
	}

	// 7. Network / Connection errors
	if strings.Contains(lower, "connection timed out") ||
		strings.Contains(lower, "could not connect") ||
		strings.Contains(lower, "network unreachable") ||
		strings.Contains(lower, "content servers unreachable") {
		return &AuthResult{
			Success:      false,
			Username:     username,
			ErrorMessage: "Could not connect to Steam authentication servers. Please check your internet connection and try again.",
		}
	}

	// 8. Fallback: Clean user-friendly message (NEVER expose raw stdout/stderr with file paths or PIDs)
	return &AuthResult{
		Success:      false,
		Username:     username,
		ErrorMessage: "Steam authentication failed. Please verify your account name and password, or check if Steam Guard is required.",
	}
}

// Authenticate attempts to log into SteamCMD non-interactively with username and password.
func Authenticate(steamcmdDir, username, password string) (*AuthResult, error) {
	cmdScript := filepath.Join(steamcmdDir, "steamcmd.sh")
	cmd := exec.Command(cmdScript, "+login", username, password, "+quit")
	cmd.Dir = steamcmdDir

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	_ = cmd.Run()
	output := outBuf.String() + "\n" + errBuf.String()

	return parseSteamAuthOutput(output, username), nil
}

// AuthenticateWith2FA attempts to log into SteamCMD using a 2FA Steam Guard code.
func AuthenticateWith2FA(steamcmdDir, username, password, guardCode string) (*AuthResult, error) {
	cmdScript := filepath.Join(steamcmdDir, "steamcmd.sh")
	// Passing guardCode via both +set_steam_guard_code and +login username password guardCode ensures maximum Valve compatibility
	cmd := exec.Command(cmdScript, "+set_steam_guard_code", guardCode, "+login", username, password, guardCode, "+quit")
	cmd.Dir = steamcmdDir

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	_ = cmd.Run()
	output := outBuf.String() + "\n" + errBuf.String()

	return parseSteamAuthOutput(output, username), nil
}

// CheckOwnership checks if the logged in Steam user owns a given appID.
func CheckOwnership(steamcmdDir, username, appID string) (bool, error) {
	cmdScript := filepath.Join(steamcmdDir, "steamcmd.sh")
	cmd := exec.Command(cmdScript, "+login", username, "+app_info_print", appID, "+quit")
	cmd.Dir = steamcmdDir

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf

	err := cmd.Run()
	output := outBuf.String()

	if strings.Contains(output, "No subscription") || strings.Contains(output, "ERROR! Failed to request app info") {
		// Attempt to request a free license
		licCmd := exec.Command(cmdScript, "+login", username, "+app_license_request", appID, "+quit")
		licCmd.Dir = steamcmdDir
		
		var licOutBuf bytes.Buffer
		licCmd.Stdout = &licOutBuf
		licCmd.Run()
		licOutput := licOutBuf.String()
		
		if strings.Contains(licOutput, "License Request OK") || strings.Contains(licOutput, "already own") || strings.Contains(licOutput, "OK") {
			return true, nil
		}
		
		return false, nil
	}

	if strings.Contains(output, fmt.Sprintf(`"%s"`, appID)) || strings.Contains(output, "appid") || err == nil {
		return true, nil
	}

	return false, nil
}

// DownloadSteamApp downloads a Steam app using native macOS SteamCMD into targetDir.
// If the first attempt fails with "No subscription", it will automatically try to
// request a free license and retry once.
// Returns a *SteamCMDError with a classified error type, or nil on success.
func DownloadSteamApp(steamcmdDir, installDir, username, appID string, forceWindows bool, ctx context.Context) error {
	cmdScript := filepath.Join(steamcmdDir, "steamcmd.sh")

	args := []string{}
	if forceWindows {
		args = append(args, "+@sSteamCmdForcePlatformType", "windows")
	}
	args = append(args,
		"+force_install_dir", installDir,
		"+login", username,
		"+app_update", appID, "validate",
		"+quit",
	)

	cmd := exec.CommandContext(ctx, cmdScript, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process != nil {
			return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		return nil
	}
	cmd.Dir = steamcmdDir

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	output := outBuf.String() + "\n" + errBuf.String()

	// Even if err == nil, check for errors in stdout (SteamCMD sometimes exits 0 on failure)
	if classified := classifySteamCMDOutput(output); classified != nil {
		// Special handling: if it's a "No subscription" error, try to request a free license and retry once
		if classified.Type == SteamErrNoSubscription {
			licArgs := []string{"+login", username, "+app_license_request", appID, "+quit"}
			licCmd := exec.CommandContext(ctx, cmdScript, licArgs...)
			licCmd.Dir = steamcmdDir
			licCmd.Run() // Best-effort

			// Retry the download
			retryArgs := []string{}
			if forceWindows {
				retryArgs = append(retryArgs, "+@sSteamCmdForcePlatformType", "windows")
			}
			retryArgs = append(retryArgs,
				"+force_install_dir", installDir,
				"+login", username,
				"+app_update", appID, "validate",
				"+quit",
			)

			retryCmd := exec.CommandContext(ctx, cmdScript, retryArgs...)
			retryCmd.Dir = steamcmdDir

			var retryOut, retryErr bytes.Buffer
			retryCmd.Stdout = &retryOut
			retryCmd.Stderr = &retryErr

			retryRunErr := retryCmd.Run()
			retryOutput := retryOut.String() + "\n" + retryErr.String()

			if retryClassified := classifySteamCMDOutput(retryOutput); retryClassified != nil {
				return retryClassified
			}
			if retryRunErr != nil {
				if ctx.Err() != nil {
					return &SteamCMDError{Type: SteamErrCancelled, Message: "Download was cancelled.", RawLog: retryOutput}
				}
				return &SteamCMDError{Type: SteamErrGeneric, Message: "Download failed after retrying. Please try again.", RawLog: retryOutput}
			}
			return nil // Retry succeeded
		}

		// For all other classified errors, return immediately
		return classified
	}

	if err != nil {
		if ctx.Err() != nil {
			return &SteamCMDError{Type: SteamErrCancelled, Message: "Download was cancelled.", RawLog: output}
		}
		return &SteamCMDError{Type: SteamErrGeneric, Message: fmt.Sprintf("SteamCMD failed unexpectedly: %v", err), RawLog: output}
	}

	return nil
}

// TailProgress monitors logs/console_log.txt in the steamcmdDir and streams download progress.
func TailProgress(steamcmdDir string, ctx context.Context) <-chan DownloadProgress {
	ch := make(chan DownloadProgress, 10)

	logPath := filepath.Join(steamcmdDir, "logs", "console_log.txt")
	if home, err := os.UserHomeDir(); err == nil {
		// Fallback check for ~/Library/Application Support/Steam/logs/console_log.txt or ~/Steam/logs/console_log.txt
		altPath := filepath.Join(home, "Steam", "logs", "console_log.txt")
		if _, err := os.Stat(altPath); err == nil {
			logPath = altPath
		}
	}

	go func() {
		defer close(ch)

		ticker := time.NewTicker(1500 * time.Millisecond)
		defer ticker.Stop()

		var lastBytes int64
		var lastTime time.Time = time.Now()

		reProgress := regexp.MustCompile(`progress:\s*([0-9\.]+)\s*\(([0-9]+)\s*/\s*([0-9]+)\)`)

		var filePos int64 = 0

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				file, err := os.Open(logPath)
				if err != nil {
					continue
				}

				stat, err := file.Stat()
				if err != nil {
					file.Close()
					continue
				}

				if stat.Size() < filePos {
					filePos = 0 // file truncated
				}

				file.Seek(filePos, io.SeekStart)
				content, err := io.ReadAll(file)
				filePos, _ = file.Seek(0, io.SeekCurrent)
				file.Close()

				if len(content) == 0 {
					continue
				}

				var latestPct float64
				var latestDone int64
				var latestTotal int64
				foundUpdate := false

				lines := strings.Split(string(content), "\n")
				for _, line := range lines {
					matches := reProgress.FindStringSubmatch(line)
					if len(matches) == 4 {
						latestPct, _ = strconv.ParseFloat(matches[1], 64)
						latestDone, _ = strconv.ParseInt(matches[2], 10, 64)
						latestTotal, _ = strconv.ParseInt(matches[3], 10, 64)
						foundUpdate = true
					}
				}

				if foundUpdate {
					now := time.Now()
					elapsed := now.Sub(lastTime).Seconds()
					speed := 0.0
					
					// Only calculate speed and update tracking if enough time has passed (prevents jitter/0 values)
					if elapsed >= 1.0 {
						if latestDone > lastBytes {
							speed = float64(latestDone-lastBytes) / (1024 * 1024 * elapsed)
						}
						lastBytes = latestDone
						lastTime = now
					}

					ch <- DownloadProgress{
						Percent:    latestPct,
						BytesDone:  latestDone,
						BytesTotal: latestTotal,
						SpeedMBps:  speed,
						Stage:      "Downloading...",
					}
				}
			}
		}
	}()

	return ch
}

var steamcmdMutex sync.Mutex
