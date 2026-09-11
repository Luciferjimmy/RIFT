package rift

import (
	"os"
	"path/filepath"
	"rift/internal/cloud"
	"testing"
)

func TestLoadEnvVars(t *testing.T) {
	// First, check that if env vars are natively set, they win.
	os.Setenv("SUPABASE_URL", "native_url")
	os.Setenv("SUPABASE_ANON_KEY", "native_key")

	url, key := loadEnvVars()
	if url != "native_url" || key != "native_key" {
		t.Errorf("Expected native env vars to win, got %s / %s", url, key)
	}

	// Now unset them to test fallback
	os.Unsetenv("SUPABASE_URL")
	os.Unsetenv("SUPABASE_ANON_KEY")

	// Set up mock .env file
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	riftDir := filepath.Join(tempDir, ".rift")
	os.MkdirAll(riftDir, 0755)

	envContent := `
SUPABASE_URL="file_url"
SUPABASE_ANON_KEY="file_key"
`
	envFile := filepath.Join(riftDir, ".env")
	os.WriteFile(envFile, []byte(envContent), 0644)

	url, key = loadEnvVars()
	if url != "file_url" || key != "file_key" {
		t.Errorf("Expected file env vars, got %s / %s", url, key)
	}

	// Now remove .env file: should fall back to default production credentials
	os.Remove(envFile)
	url, key = loadEnvVars()
	if url != cloud.DefaultSupabaseURL || key != cloud.DefaultSupabaseAnonKey {
		t.Errorf("Expected default production credentials, got %s / %s", url, key)
	}
}
