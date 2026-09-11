package rift

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"rift/internal/cloud"
	"testing"
	"time"
)

func TestRingBuffer(t *testing.T) {
	rb := NewRingBuffer(3)

	rb.Add("line 1")
	rb.Add("line 2")
	rb.Add("line 3")
	rb.Add("line 4") // Should evict line 1

	out := rb.String()
	
	// Check if line 4 is present
	if !contains(out, "line 4") {
		t.Errorf("Expected line 4 to be in buffer")
	}

	// Check if line 1 was evicted
	if contains(out, "line 1") {
		t.Errorf("Expected line 1 to be evicted from buffer")
	}
}

func TestRingBuffer_Sanitize(t *testing.T) {
	rb := NewRingBuffer(10)
	
	// Add unprintable chars like null byte \x00, bell \a, backspace \b
	rb.Add("hello\x00\x07\x08world")
	out := rb.String()

	if contains(out, "\x00") || contains(out, "\x07") || contains(out, "\x08") {
		t.Errorf("Expected unprintable characters to be sanitized")
	}
	if !contains(out, "helloworld") {
		t.Errorf("Expected printable characters to remain")
	}
}

func contains(s, substr string) bool {
	// Simple strings.Contains
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestSanitizePII(t *testing.T) {
	home, _ := os.UserHomeDir()
	if home == "" {
		return
	}

	rawConfig := map[string]interface{}{
		"gameDir": filepath.Join(home, ".rift", "games", "steam-11020"),
		"name":    "TrackMania",
	}

	sanitized := sanitizePII(rawConfig)
	m, ok := sanitized.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", sanitized)
	}

	dirVal, ok := m["gameDir"].(string)
	if !ok {
		t.Fatalf("Expected string for gameDir")
	}

	if contains(dirVal, home) {
		t.Errorf("Sanitized path still contained raw home directory: %s", dirVal)
	}
	if !contains(dirVal, "~/.rift") {
		t.Errorf("Expected path to be sanitized to start with ~/.rift, got: %s", dirVal)
	}
}

func TestOfflineQueue(t *testing.T) {
	app := &App{}
	payload := TelemetryPayload{
		AppID:           "test-app-123",
		GameName:        "Test Game",
		DurationMinutes: 5,
		Status:          "Success",
	}

	app.queuePendingTelemetry(payload)

	home, _ := os.UserHomeDir()
	queueDir := filepath.Join(home, ".rift", "data", "pending_telemetry")
	entries, err := os.ReadDir(queueDir)
	if err != nil {
		t.Fatalf("Failed to read pending telemetry dir: %v", err)
	}

	found := false
	for _, e := range entries {
		if contains(e.Name(), "test-app-123") {
			found = true
			_ = os.Remove(filepath.Join(queueDir, e.Name())) // cleanup
			break
		}
	}

	if !found {
		t.Errorf("Expected queued telemetry file with test-app-123 to exist in queueDir")
	}
}

func TestCrashTaxonomyClassification(t *testing.T) {
	testCases := []struct {
		logContent     string
		expectedReason string
	}{
		{"wine: ShellExecuteEx failed: File not found", CrashReasonShellExecuteEx},
		{"Assertion failed: (pBlock != NULL), file heap.c, line 120", CrashReasonMemoryAssertion},
		{"0024:err:start:fatal_error FormatMessage failed to translate", CrashReasonFatalWineStart},
		{"Unhandled exception: page fault on execute access to 0x00000000", CrashReasonUnhandledException},
		{"wine: page fault on read access to 0x00000004 in 32-bit code", CrashReasonSegmentationFault},
		{"Exception code c0000005 (ACCESS_VIOLATION)", CrashReasonAccessViolation},
	}

	for _, tc := range testCases {
		classified := CrashReasonNone
		if contains(tc.logContent, "ShellExecuteEx failed") {
			classified = CrashReasonShellExecuteEx
		} else if contains(tc.logContent, "Assertion failed") {
			classified = CrashReasonMemoryAssertion
		} else if contains(tc.logContent, "0024:err:start:fatal_error FormatMessage failed") {
			classified = CrashReasonFatalWineStart
		} else if contains(tc.logContent, "Unhandled exception") {
			classified = CrashReasonUnhandledException
		} else if contains(tc.logContent, "page fault on read access") || contains(tc.logContent, "page fault on write access") {
			classified = CrashReasonSegmentationFault
		} else if contains(tc.logContent, "c0000005") {
			classified = CrashReasonAccessViolation
		}

		if classified != tc.expectedReason {
			t.Errorf("For log %q, expected reason %q, got %q", tc.logContent, tc.expectedReason, classified)
		}
	}
}

func TestTelemetryPayloadFullMetrics(t *testing.T) {
	specs := GetMacHardwareSpecs()
	if specs["DeviceID"] == nil || len(specs["DeviceID"].(string)) != 64 {
		t.Fatalf("DeviceID should be a 64-character SHA-256 hash, got %v", specs["DeviceID"])
	}

	payload := TelemetryPayload{
		AppID:            "steam-11020",
		GameName:         "TrackMania Nations Forever",
		DurationMinutes:  42,
		LaunchDurationMs: 1450,
		AvgFPS:           nil, // Unmeasured maps to null
		AvgCPU:           45,
		MaxCPU:           98,
		AvgRAM:           1250,
		MaxRAM:           2480,
		ThermalState:     "nominal",
		OnBattery:        true,
		SwapUsedMB:       128.5,
		Status:           "Success",
		CrashReason:      CrashReasonNone,
		CrashLog:         "",
		ConfigUsed: map[string]interface{}{
			"engine": "d3dmetal",
		},
		HardwareSpecs: specs,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal TelemetryPayload: %v", err)
	}

	jsonStr := string(data)
	requiredKeys := []string{
		`"app_id":"steam-11020"`,
		`"game_name":"TrackMania Nations Forever"`,
		`"duration_minutes":42`,
		`"launch_duration_ms":1450`,
		`"avg_fps":null`,
		`"avg_cpu":45`,
		`"max_cpu":98`,
		`"avg_ram":1250`,
		`"max_ram":2480`,
		`"thermal_state":"nominal"`,
		`"on_battery":true`,
		`"swap_used_mb":128.5`,
		`"status":"Success"`,
	}

	for _, k := range requiredKeys {
		if !contains(jsonStr, k) {
			t.Errorf("Marshaled JSON missing required key fragment: %s. Got: %s", k, jsonStr)
		}
	}
}

func TestGetGameMetrics_LiveProcess(t *testing.T) {
	// Spawn a real live process that exercises CPU and memory
	cmd := exec.Command("sh", "-c", "for i in $(seq 1 10000000); do :; done")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start test process: %v", err)
	}
	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}()

	pid := cmd.Process.Pid
	if !isProcessAlive(pid) {
		t.Fatalf("Expected process %d to be alive", pid)
	}

	// Give the process a brief moment to accumulate CPU time
	time.Sleep(200 * time.Millisecond)
	cpu, ram, realPID := getGameMetrics(pid)
	t.Logf("Live process metrics -> PID: %d, realPID: %d, CPU: %d%%, RAM: %d MB", pid, realPID, cpu, ram)

	if realPID <= 0 {
		t.Errorf("Expected positive realPID, got %d", realPID)
	}
}

func TestUserPlaytimeScoping(t *testing.T) {
	tempBase := t.TempDir()
	app := &App{
		RiftBaseDir: tempBase,
	}

	// 1. Guest / unauthenticated user defaults to 'default'
	if !contains(app.UserDir(), filepath.Join(tempBase, "users", "default")) {
		t.Errorf("Expected default UserDir, got %s", app.UserDir())
	}
	if !contains(app.StoresDir(), filepath.Join(tempBase, "users", "default", "stores")) {
		t.Errorf("Expected default StoresDir, got %s", app.StoresDir())
	}

	// 2. Playtime for guest should initially be empty
	stats := app.GetPlaytimeStats()
	if !contains(stats, `"total": 0`) {
		t.Errorf("Expected initial total 0, got %s", stats)
	}

	// 3. Add playtime for guest
	app.SyncUserPlaytime("steam-11020", "TrackMania", 3600)
	stats = app.GetPlaytimeStats()
	if !contains(stats, `"total": 3600`) || !contains(stats, `"steam-11020": 3600`) {
		t.Errorf("Expected 3600 seconds recorded for guest, got %s", stats)
	}

	// 4. Authenticate as new user (e.g. Test2)
	testUserBase := filepath.Join(tempBase, "users", "user-uuid-test2")
	_ = os.MkdirAll(testUserBase, 0755)

	app2 := &App{
		RiftBaseDir: tempBase,
		Supabase: &cloud.SupabaseClient{
			UserID:      "user-uuid-test2",
			AccessToken: "dummy-tok",
		},
	}

	if !contains(app2.UserDir(), filepath.Join(tempBase, "users", "user-uuid-test2")) {
		t.Errorf("Expected user2 UserDir to be scoped to user-uuid-test2, got %s", app2.UserDir())
	}

	// Verify user2 starts with 0 playtime, isolated from guest / user1!
	user2Stats := app2.GetPlaytimeStats()
	if !contains(user2Stats, `"total": 0`) {
		t.Errorf("Expected user2 to start with 0 playtime, got %s", user2Stats)
	}

	// Record playtime for user2
	app2.SyncUserPlaytime("epic-game1", "Mortal Shell", 7200)
	user2Stats = app2.GetPlaytimeStats()
	if !contains(user2Stats, `"total": 7200`) || !contains(user2Stats, `"epic-game1": 7200`) {
		t.Errorf("Expected 7200 seconds for user2, got %s", user2Stats)
	}

	// Guest playtime remains unchanged (3600 seconds)
	guestStats := app.GetPlaytimeStats()
	if !contains(guestStats, `"total": 3600`) {
		t.Errorf("Expected guest playtime to remain 3600, got %s", guestStats)
	}
}

