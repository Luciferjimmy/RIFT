package rift

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"rift/internal/sysmon"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// RingBuffer stores a maximum number of string lines.
type RingBuffer struct {
	lines []string
	max   int
	mu    sync.Mutex
}

func NewRingBuffer(maxLines int) *RingBuffer {
	return &RingBuffer{
		lines: make([]string, 0, maxLines),
		max:   maxLines,
	}
}

func (rb *RingBuffer) Add(line string) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if len(rb.lines) >= rb.max {
		// Remove oldest
		rb.lines = rb.lines[1:]
	}
	rb.lines = append(rb.lines, fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), line))
}

func (rb *RingBuffer) String() string {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	// Cap the payload string to ~5MB safely by ensuring we don't output too many characters
	content := strings.Join(rb.lines, "\n")
	if len(content) > 5*1024*1024 {
		content = content[len(content)-5*1024*1024:]
	}
	// Sanitize unprintable characters
	var clean strings.Builder
	clean.Grow(len(content))
	for _, r := range content {
		if strconv.IsPrint(r) || r == '\n' || r == '\t' || r == '\r' {
			clean.WriteRune(r)
		}
	}
	return clean.String()
}

// Structured Crash Taxonomy Enums for AI curation and query analytics
const (
	CrashReasonNone               = ""
	CrashReasonShellExecuteEx     = "ShellExecuteEx_NotFound"
	CrashReasonMemoryAssertion    = "MemoryAssertion_32Bit"
	CrashReasonAccessViolation    = "AccessViolation_0xC0000005"
	CrashReasonFatalWineStart     = "FatalWineStart_BadDLL"
	CrashReasonUnhandledException = "UnhandledException_Internal"
	CrashReasonSegmentationFault  = "SegmentationFault"
	CrashReasonNonZeroExit        = "NonZeroExitCode"
)

// TelemetryPayload defines the comprehensive, anonymized data sent to the cloud.
type TelemetryPayload struct {
	AppID            string      `json:"app_id"`
	GameName         string      `json:"game_name"`
	DurationMinutes  int         `json:"duration_minutes"`
	LaunchDurationMs int         `json:"launch_duration_ms"`
	AvgFPS           *int        `json:"avg_fps"` // null when unmeasured, integer when measured
	AvgCPU           int         `json:"avg_cpu,omitempty"`
	MaxCPU           int         `json:"max_cpu,omitempty"`
	AvgRAM           int         `json:"avg_ram,omitempty"`
	MaxRAM           int         `json:"max_ram,omitempty"`
	ThermalState     string      `json:"thermal_state,omitempty"`
	OnBattery        bool        `json:"on_battery"`
	SwapUsedMB       float64     `json:"swap_used_mb,omitempty"`
	Status           string      `json:"status"` // 'Success', 'Crashed', 'TerminatedEarly'
	CrashReason      string      `json:"crash_reason,omitempty"`
	CrashLog         string      `json:"crash_log,omitempty"`
	TerminalLog      string      `json:"terminal_log,omitempty"`
	ExitCode         int         `json:"exit_code"`
	DiskSizeMB       float64     `json:"disk_size_mb"`
	IsSetupComplete  bool        `json:"is_setup_complete"`
	ConfigUsed       interface{} `json:"config_used,omitempty"`
	HardwareSpecs    interface{} `json:"hardware_specs,omitempty"`
}

func getDirSizeMB(dirPath string) float64 {
	var totalBytes int64
	_ = filepath.Walk(dirPath, func(_ string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			totalBytes += info.Size()
		}
		return nil
	})
	return float64(totalBytes) / (1024 * 1024)
}

// MonitorGame watches a game process and tracks telemetry
func (a *App) MonitorGame(gameID, appID, gameName, exePath string, cmd *exec.Cmd, sysProfile interface{}, gameConfig interface{}, launchDurationMs int) {
	defer func() {
		if r := recover(); r != nil {
			a.logWarn("[Crash] Panic recovered in MonitorGame: %v", r)
		}
	}()
	startTime := time.Now()

	// Ring buffer for logs (~30,000 lines should be well under 5MB)
	logBuffer := NewRingBuffer(30000)

	// Intercept stdout/stderr — only create pipes if not already assigned
	if cmd.Stdout == nil {
		stdout, err := cmd.StdoutPipe()
		if err == nil {
			go func() {
				buf := make([]byte, 1024)
				for {
					n, err := stdout.Read(buf)
					if n > 0 {
						logBuffer.Add(string(buf[:n]))
					}
					if err != nil {
						break
					}
				}
			}()
		}
	}

	if cmd.Stderr == nil {
		stderr, err := cmd.StderrPipe()
		if err == nil {
			go func() {
				buf := make([]byte, 1024)
				for {
					n, err := stderr.Read(buf)
					if n > 0 {
						logBuffer.Add(string(buf[:n]))
					}
					if err != nil {
						break
					}
				}
			}()
		}
	}

	// Start the game process
	err := cmd.Start()
	if err != nil {
		a.logError("[Launch] Failed to start game process for %s: %v", gameID, err)
		a.emitEvent("launch_failed", map[string]interface{}{
			"gameId": gameID,
			"error":  fmt.Sprintf("Failed to start game: %v. Is the translation engine installed? Go to Settings → Engine Setup.", err),
		})
		return
	}

	wrapperPID := cmd.Process.Pid
	sysmon.SetActiveGame(int32(wrapperPID), gameID)
	a.emitEvent("launch_complete", gameID)

	// Activate RIFT Game Mode if enabled in settings
	gameModeActive := false
	cfg := a.GetFullConfig()
	if override, exists := cfg.GameOverrides[gameID]; exists && override.GameMode {
		a.ActivateGameMode(wrapperPID)
		gameModeActive = true
	}

	// Wrapper wait (runs in background so we don't prematurely kill telemetry if wrapper exits)
	waitErrCh := make(chan error, 1)
	go func() {
		waitErrCh <- cmd.Wait()
	}()

	var totalCPU, totalRAM, pollCount int
	var maxCPU, maxRAM int

	// Polling Loop
	ticker := time.NewTicker(3 * time.Second)
	var waitErr error
	var exited bool
	for !exited {
		select {
		case err := <-waitErrCh:
			waitErr = err
			exited = true
		case <-ticker.C:
			cpu, ram, realPID := getGameMetrics(wrapperPID)

			// If CPU and RAM are both 0, check if wrapper or child processes are still alive
			if cpu == 0 && ram == 0 {
				if !isProcessAlive(wrapperPID) {
					exited = true
					break
				}
			}

			sysmon.UpdateActiveGamePID(int32(realPID))

			// Only accumulate active samples to prevent dragging down CPU/RAM averages
			if cpu > 0 || ram > 0 {
				totalCPU += cpu
				totalRAM += ram
				pollCount++
				if cpu > maxCPU {
					maxCPU = cpu
				}
				if ram > maxRAM {
					maxRAM = ram
				}
			}

			// Emit to frontend graph
			runtime.EventsEmit(a.ctx, "TelemetryUpdate", map[string]interface{}{
				"cpu": cpu,
				"ram": ram,
			})
		}
	}
	ticker.Stop()

	// Parse log buffer for crashes and map to structured taxonomy enum
	rawLogs := logBuffer.String()
	crashReasonText := ""
	crashReasonEnum := CrashReasonNone

	if strings.Contains(rawLogs, "ShellExecuteEx failed") {
		crashReasonText = "ShellExecuteEx failed (Wine path error or executable missing)"
		crashReasonEnum = CrashReasonShellExecuteEx
	} else if strings.Contains(rawLogs, "Assertion failed") {
		crashReasonText = "Memory assertion failed (likely a 32-bit bug in this Wine engine)"
		crashReasonEnum = CrashReasonMemoryAssertion
	} else if strings.Contains(rawLogs, "0024:err:start:fatal_error FormatMessage failed") {
		crashReasonText = "Wine fatal start error (invalid DLL overrides or missing translation libraries)"
		crashReasonEnum = CrashReasonFatalWineStart
	} else if strings.Contains(rawLogs, "Unhandled exception") {
		crashReasonText = "Unhandled exception (Game crashed internally)"
		crashReasonEnum = CrashReasonUnhandledException
	} else if strings.Contains(rawLogs, "page fault on read access") || strings.Contains(rawLogs, "page fault on write access") {
		crashReasonText = "Segmentation fault (Wine memory access violation)"
		crashReasonEnum = CrashReasonSegmentationFault
	} else if strings.Contains(rawLogs, "c0000005") {
		crashReasonText = "Access Violation (0xC0000005)"
		crashReasonEnum = CrashReasonAccessViolation
	} else if waitErr != nil {
		crashReasonText = fmt.Sprintf("Exited with non-zero status: %v", waitErr)
		crashReasonEnum = CrashReasonNonZeroExit
	}

	if crashReasonText != "" {
		a.logError("[Crash] Game %s crashed: %s", gameID, crashReasonText)
		a.emitEvent("game_crashed", map[string]interface{}{"gameId": gameID, "error": crashReasonText})
	}

	// Deactivate RIFT Game Mode if it was activated
	if gameModeActive {
		a.DeactivateGameMode()
	}

	sessionSeconds := int(time.Since(startTime).Seconds())
	duration := int(time.Since(startTime).Minutes())

	avgCPU := 0
	avgRAM := 0
	if pollCount > 0 {
		avgCPU = totalCPU / pollCount
		avgRAM = totalRAM / pollCount
	}

	capsuleDir := a.ResolveGameDir(gameID)
	diskSizeMB := getDirSizeMB(capsuleDir)
	setupMarker := filepath.Join(capsuleDir, ".rift_setup_complete")
	_, setupErr := os.Stat(setupMarker)
	isSetupComplete := setupErr == nil

	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	} else if waitErr != nil {
		if exitError, ok := waitErr.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	// Accurate Telemetry Status Determination
	status := "Success"
	var crashLog string
	if waitErr != nil || crashReasonEnum != CrashReasonNone {
		status = "Crashed"
		crashLog = rawLogs
		if len(crashLog) > 1000 {
			a.logWarn("[Crash] Wine crash captured! Last 1000 bytes: %s", crashLog[len(crashLog)-1000:])
		} else {
			a.logWarn("[Crash] Wine crash captured! Output: %s", crashLog)
		}
	} else if !isSetupComplete && duration == 0 {
		status = "TerminatedEarly"
		crashReasonEnum = "IncompleteDownload_MissingAssets"
		crashReasonText = "Terminated early: Game assets incomplete on disk"
	}

	if appID == "" {
		appID = "unknown_app"
	}
	if gameName == "" || gameName == gameID || gameName == appID {
		if resolved := queryGameNameFromDB(appID); resolved != "" {
			gameName = resolved
		} else if resolved := queryGameNameFromDB(gameID); resolved != "" {
			gameName = resolved
		} else {
			home, _ := os.UserHomeDir()
			runtimePath := filepath.Join(home, ".rift", "games", gameID, "game_runtime.json")
			if data, err := os.ReadFile(runtimePath); err == nil {
				var rconf map[string]interface{}
				if json.Unmarshal(data, &rconf) == nil {
					if gn, ok := rconf["gameName"].(string); ok && gn != "" {
						gameName = gn
					}
				}
			}
		}
	}
	if gameName == "" {
		gameName = "Unknown Game"
	}

	// Scrub PII paths from config before sending to cloud
	safeConfig := sanitizePII(gameConfig)

	// Cap crash log to the most informative last 64KB and scrub PII paths
	var safeCrashLog string
	if crashLog != "" {
		truncated := crashLog
		if len(truncated) > 64*1024 {
			truncated = "... [Earlier logs truncated for cloud upload] ...\n" + truncated[len(truncated)-64*1024:]
		}
		safeCrashLog = sanitizeStringPII(truncated)
	}

	// Capture full terminal output (up to 256KB for complete forensics on every run)
	truncatedTermLog := rawLogs
	if len(truncatedTermLog) > 256*1024 {
		truncatedTermLog = "... [Earlier logs truncated for cloud upload] ...\n" + truncatedTermLog[len(truncatedTermLog)-256*1024:]
	}
	safeTerminalLog := sanitizeStringPII(truncatedTermLog)

	// Capture live environmental metrics
	thermalState := GetThermalState()
	onBattery := IsOnBattery()
	swapUsed := GetSwapUsedMB()

	payload := TelemetryPayload{
		AppID:            appID,
		GameName:         gameName,
		DurationMinutes:  duration,
		LaunchDurationMs: launchDurationMs,
		AvgCPU:           avgCPU,
		MaxCPU:           maxCPU,
		AvgRAM:           avgRAM,
		MaxRAM:           maxRAM,
		AvgFPS:           nil, // Unmeasured: nil maps cleanly to null in JSON/Postgres
		ThermalState:     thermalState,
		OnBattery:        onBattery,
		SwapUsedMB:       swapUsed,
		Status:           status,
		CrashReason:      crashReasonEnum,
		CrashLog:         safeCrashLog,
		TerminalLog:      safeTerminalLog,
		ExitCode:         exitCode,
		DiskSizeMB:       diskSizeMB,
		IsSetupComplete:  isSetupComplete,
		ConfigUsed:       safeConfig,
		HardwareSpecs:    sysProfile,
	}

	// Persist full telemetry session payload locally for transparency and user inspection
	home, _ := os.UserHomeDir()
	summaryPath := filepath.Join(home, ".rift", "logs", "last_session_summary.json")
	if summaryBytes, err := json.MarshalIndent(payload, "", "  "); err == nil {
		_ = os.WriteFile(summaryPath, summaryBytes, 0644)
	}

	a.uploadTelemetry(payload)
	a.SyncUserPlaytime(gameID, gameName, sessionSeconds)

	platform := "epic"
	if strings.HasPrefix(gameID, "steam-") {
		platform = "steam"
	}

	a.TrackUserEvent("game_exited", gameID, appID, platform, map[string]interface{}{
		"duration_seconds": sessionSeconds,
		"exit_code":        exitCode,
		"status":           status,
		"crash_reason":     crashReasonEnum,
	})
}

// getProcessMetrics shells out to `ps` to get CPU and RAM (in MB) for a PID.
func getProcessMetrics(pid int) (int, int) {
	out, err := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", "%cpu,rss").Output()
	if err != nil {
		return 0, 0
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return 0, 0
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 2 {
		return 0, 0
	}

	var cpu float64
	var ramKB int
	fmt.Sscanf(fields[0], "%f", &cpu)
	fmt.Sscanf(fields[1], "%d", &ramKB)

	return int(cpu), ramKB / 1024
}

// collectAllProcessPIDs discovers all processes in the game and Wine translation tree
func collectAllProcessPIDs(rootPID int) []int {
	pids := []int{rootPID}
	visited := map[int]bool{rootPID: true}
	queue := []int{rootPID}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		out, err := exec.Command("pgrep", "-P", fmt.Sprintf("%d", curr)).Output()
		if err == nil {
			for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				var p int
				if _, err := fmt.Sscanf(line, "%d", &p); err == nil && p > 0 && !visited[p] {
					visited[p] = true
					pids = append(pids, p)
					queue = append(queue, p)
				}
			}
		}
	}

	// Also query the process group to ensure any detached daemons are caught
	pgidOut, err := exec.Command("ps", "-o", "pgid=", "-p", fmt.Sprintf("%d", rootPID)).Output()
	if err == nil {
		pgid := strings.TrimSpace(string(pgidOut))
		if pgid != "" {
			if out, err := exec.Command("pgrep", "-g", pgid).Output(); err == nil {
				for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
					var p int
					if _, err := fmt.Sscanf(line, "%d", &p); err == nil && p > 0 && !visited[p] {
						visited[p] = true
						pids = append(pids, p)
					}
				}
			}
		}
	}

	return pids
}

// getGameMetrics sums the CPU and RAM of all processes matching the Wine translation tree and wrapper
func getGameMetrics(wrapperPID int) (int, int, int) {
	pids := collectAllProcessPIDs(wrapperPID)
	var totalCPU int
	var totalRAM int
	realPID := wrapperPID
	maxRAM := 0

	for _, pid := range pids {
		c, r := getProcessMetrics(pid)
		totalCPU += c
		totalRAM += r

		// The process using the most memory is usually the actual game .exe
		if r > maxRAM {
			maxRAM = r
			realPID = pid
		}
	}

	// Fallback to wrapper process directly if tree returned nothing
	if totalCPU == 0 && totalRAM == 0 {
		c, r := getProcessMetrics(wrapperPID)
		totalCPU = c
		totalRAM = r
	}

	// Normalize CPU across logical cores so it represents true 0% to 100% total system CPU
	numCores := goruntime.NumCPU()
	if numCores <= 0 {
		numCores = 1
	}
	normalizedCPU := (totalCPU + numCores/2) / numCores
	if normalizedCPU > 100 {
		normalizedCPU = 100
	}

	return normalizedCPU, totalRAM, realPID
}

func isProcessAlive(pid int) bool {
	err := exec.Command("ps", "-p", fmt.Sprintf("%d", pid)).Run()
	return err == nil
}

// sanitizeStringPII replaces user home paths and usernames with [REDACTED] or ~
func sanitizeStringPII(s string) string {
	if s == "" {
		return ""
	}
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		s = strings.ReplaceAll(s, home, "~")
		winHome := strings.ReplaceAll(home, "/", "\\")
		s = strings.ReplaceAll(s, winHome, "~")
		username := filepath.Base(home)
		if username != "" && len(username) > 2 {
			s = strings.ReplaceAll(s, "/Users/"+username, "/Users/[REDACTED]")
			s = strings.ReplaceAll(s, "\\Users\\"+username, "\\Users\\[REDACTED]")
			s = strings.ReplaceAll(s, "\\users\\"+username, "\\users\\[REDACTED]")
		}
	}
	return s
}

// sanitizePII replaces user home paths with ~ to prevent leaking usernames to telemetry
func sanitizePII(input interface{}) interface{} {
	if input == nil {
		return nil
	}
	b, err := json.Marshal(input)
	if err != nil {
		return input
	}
	s := sanitizeStringPII(string(b))
	var out interface{}
	if err := json.Unmarshal([]byte(s), &out); err == nil {
		return out
	}
	return input
}

func (a *App) uploadTelemetry(payload TelemetryPayload) {
	if a.GetFullConfig().DisableTelemetry {
		a.logInfo("[Telemetry] Skipped upload: Telemetry disabled by user in Settings.")
		return
	}

	// Flush any previously queued offline telemetry sessions
	go a.FlushPendingTelemetry()

	if !a.sendTelemetryPayload(payload) {
		a.queuePendingTelemetry(payload)
	}
}

func (a *App) sendTelemetryPayload(payload TelemetryPayload) bool {
	if a.Supabase == nil || a.Supabase.ProjectURL == "" {
		a.logInfo("[Telemetry] Supabase client not initialized; queueing telemetry.")
		return false
	}

	token := a.Supabase.AccessToken
	if token == "" {
		a.logInfo("[Telemetry] User unauthenticated; queueing telemetry session locally.")
		return false
	}

	url := fmt.Sprintf("%s/functions/v1/upload-telemetry", a.Supabase.ProjectURL)
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return false
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		a.logError("[Telemetry] Failed to create request: %v", err)
		return false
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("apikey", a.Supabase.AnonKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-RIFT-Version", AppVersion)
	req.Header.Set("X-RIFT-Client", "desktop-mac")

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		a.logInfo("[Telemetry] Offline / Network unreachable; queued telemetry payload locally.")
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		a.logWarn("[Telemetry] Supabase rejected upload (Status: %d). Error: %s", resp.StatusCode, string(body))
		return false
	}

	a.logInfo("[Telemetry] Successfully uploaded game session for %s to RIFT Cloud!", payload.GameName)
	return true
}

func (a *App) queuePendingTelemetry(payload TelemetryPayload) {
	home, _ := os.UserHomeDir()
	queueDir := filepath.Join(home, ".rift", "data", "pending_telemetry")
	_ = os.MkdirAll(queueDir, 0755)

	safeAppID := strings.ReplaceAll(payload.AppID, "/", "_")
	filename := fmt.Sprintf("%d_%s.json", time.Now().UnixMilli(), safeAppID)
	filePath := filepath.Join(queueDir, filename)
	if data, err := json.MarshalIndent(payload, "", "  "); err == nil {
		_ = os.WriteFile(filePath, data, 0644)
		a.logInfo("[Telemetry] Queued pending telemetry session: %s", filename)
	}
}

// FlushPendingTelemetry uploads any locally stored telemetry sessions that occurred while offline
func (a *App) FlushPendingTelemetry() {
	if a.GetFullConfig().DisableTelemetry || a.Supabase == nil || a.Supabase.AccessToken == "" {
		if a.GetFullConfig().DisableTelemetry {
			home, _ := os.UserHomeDir()
			_ = os.RemoveAll(filepath.Join(home, ".rift", "data", "pending_telemetry"))
		}
		return
	}
	home, _ := os.UserHomeDir()
	queueDir := filepath.Join(home, ".rift", "data", "pending_telemetry")
	entries, err := os.ReadDir(queueDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		fPath := filepath.Join(queueDir, entry.Name())
		data, err := os.ReadFile(fPath)
		if err != nil {
			continue
		}
		var payload TelemetryPayload
		if err := json.Unmarshal(data, &payload); err == nil {
			if a.sendTelemetryPayload(payload) {
				_ = os.Remove(fPath)
			}
		} else {
			_ = os.Remove(fPath)
		}
	}
}

// TrackUserEvent records significant user journey milestones (setup finished, store connected, game searched, download started/completed).
func (a *App) TrackUserEvent(eventName, gameID, appID, platform string, metadata map[string]interface{}) {
	if a.GetFullConfig().DisableTelemetry || a.Supabase == nil || !a.Supabase.IsLoggedIn() {
		return
	}

	payload := map[string]interface{}{
		"user_id":    a.Supabase.UserID,
		"event_name": eventName,
		"game_id":    gameID,
		"app_id":     appID,
		"platform":   platform,
		"metadata":   metadata,
	}

	go func() {
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			return
		}
		url := fmt.Sprintf("%s/rest/v1/user_events", a.Supabase.ProjectURL)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return
		}
		req.Header.Set("Authorization", "Bearer "+a.Supabase.AccessToken)
		req.Header.Set("apikey", a.Supabase.AnonKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Prefer", "return=minimal")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				a.logInfo("[Telemetry] Tracked event: %s", eventName)
			} else {
				a.logWarn("[Telemetry] TrackUserEvent %s returned status %d", eventName, resp.StatusCode)
			}
		} else {
			a.logWarn("[Telemetry] TrackUserEvent %s error: %v", eventName, err)
		}
	}()
}

// LocalPlaytime represents user-scoped playtime tracking structure.
type LocalPlaytime struct {
	Total   int64            `json:"total"`
	PerGame map[string]int64 `json:"per_game"`
}

// SyncUserPlaytime updates the user-scoped playtime JSON locally and upserts to Supabase public.user_playtime.
func (a *App) SyncUserPlaytime(gameID, gameName string, sessionSeconds int) {
	userDir := a.UserDir()
	_ = os.MkdirAll(userDir, 0755)
	playtimeFile := filepath.Join(userDir, "playtime.json")

	var pt LocalPlaytime
	pt.PerGame = make(map[string]int64)

	if data, err := os.ReadFile(playtimeFile); err == nil {
		_ = json.Unmarshal(data, &pt)
	}
	if pt.PerGame == nil {
		pt.PerGame = make(map[string]int64)
	}

	if sessionSeconds > 0 && gameID != "" {
		pt.Total += int64(sessionSeconds)
		pt.PerGame[gameID] += int64(sessionSeconds)
	}

	if data, err := json.MarshalIndent(pt, "", "  "); err == nil {
		_ = os.WriteFile(playtimeFile, data, 0644)
	}

	totalHours := math.Round((float64(pt.Total)/3600.0)*100) / 100.0
	gamesPlayed := len(pt.PerGame)

	a.emitEvent("playtime_updated", map[string]interface{}{
		"total":       pt.Total,
		"totalHours":  totalHours,
		"gamesPlayed": gamesPlayed,
		"lastGame":    gameName,
		"perGame":     pt.PerGame,
	})

	if a.Supabase == nil || !a.Supabase.IsLoggedIn() || a.Supabase.UserID == "" {
		return
	}

	go func() {
		cloudPayload := map[string]interface{}{
			"user_id":          a.Supabase.UserID,
			"total_seconds":    pt.Total,
			"total_hours":      totalHours,
			"per_game":         pt.PerGame,
			"games_played":     gamesPlayed,
			"last_played_game": gameName,
			"last_played_at":   time.Now().UTC().Format(time.RFC3339),
			"updated_at":       time.Now().UTC().Format(time.RFC3339),
		}

		bodyBytes, err := json.Marshal(cloudPayload)
		if err != nil {
			return
		}

		url := fmt.Sprintf("%s/rest/v1/user_playtime?on_conflict=user_id", a.Supabase.ProjectURL)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return
		}
		req.Header.Set("Authorization", "Bearer "+a.Supabase.AccessToken)
		req.Header.Set("apikey", a.Supabase.AnonKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Prefer", "resolution=merge-duplicates,return=minimal")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				a.logInfo("[Playtime] Synced cloud playtime to Supabase: %.2f hours, %d games", totalHours, gamesPlayed)
			} else {
				a.logWarn("[Playtime] Supabase user_playtime upsert returned status %d", resp.StatusCode)
			}
		} else {
			a.logWarn("[Playtime] Failed to sync user_playtime to Supabase: %v", err)
		}
	}()
}

// SyncCloudPlaytimeOnLogin pulls cloud playtime from Supabase if local cache is empty,
// or pushes local playtime if Supabase has no record yet.
func (a *App) SyncCloudPlaytimeOnLogin() {
	if a.Supabase == nil || !a.Supabase.IsLoggedIn() || a.Supabase.UserID == "" {
		return
	}

	userDir := a.UserDir()
	_ = os.MkdirAll(userDir, 0755)
	playtimeFile := filepath.Join(userDir, "playtime.json")

	url := fmt.Sprintf("%s/rest/v1/user_playtime?user_id=eq.%s&select=*", a.Supabase.ProjectURL, a.Supabase.UserID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+a.Supabase.AccessToken)
	req.Header.Set("apikey", a.Supabase.AnonKey)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return
	}

	var rows []struct {
		TotalSeconds   int64            `json:"total_seconds"`
		TotalHours     float64          `json:"total_hours"`
		PerGame        map[string]int64 `json:"per_game"`
		GamesPlayed    int              `json:"games_played"`
		LastPlayedGame string           `json:"last_played_game"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rows); err == nil && len(rows) > 0 {
		row := rows[0]
		pt := LocalPlaytime{
			Total:   row.TotalSeconds,
			PerGame: row.PerGame,
		}
		if pt.PerGame == nil {
			pt.PerGame = make(map[string]int64)
		}
		if data, err := json.MarshalIndent(pt, "", "  "); err == nil {
			_ = os.WriteFile(playtimeFile, data, 0644)
			a.logInfo("[Playtime] Restored %d seconds (%.2f hours) from Supabase cloud for user %s", pt.Total, row.TotalHours, a.Supabase.UserID)
		}
		a.emitEvent("playtime_updated", map[string]interface{}{
			"total":       pt.Total,
			"totalHours":  row.TotalHours,
			"gamesPlayed": row.GamesPlayed,
			"lastGame":    row.LastPlayedGame,
			"perGame":     pt.PerGame,
		})
	} else {
		// If cloud has 0 rows, check if local has playtime to push
		if localData, err := os.ReadFile(playtimeFile); err == nil {
			var pt LocalPlaytime
			if json.Unmarshal(localData, &pt) == nil && pt.Total > 0 {
				a.SyncUserPlaytime("", "", 0)
			}
		}
	}
}
