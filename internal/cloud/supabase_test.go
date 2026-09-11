package cloud

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"rift/internal/types"
	"testing"
)

func TestSupabaseClient_GetGameConfig(t *testing.T) {
	// Setup test server with IPv4 listener to prevent sandbox bind issues
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create IPv4 listener: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/functions/v1/get-game-config" {
			// Validate Auth
			if r.Header.Get("Authorization") != "Bearer test-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			resp := EdgeFunctionResponse{
				Source: "ai_live",
				Config: types.CloudGameConfig{
					Engine:     "d3dmetal",
					Confidence: 0.99,
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	ts := httptest.NewUnstartedServer(handler)
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Temp auth dir
	tempDir := t.TempDir()

	client := NewSupabaseClient(ts.URL, "anon", tempDir)
	// Force login state
	client.AccessToken = "test-token"

	sysProfile := types.SystemProfile{
		Chip:         "M3",
		HasGPTK:      true,
		MacOSVersion: "14.4",
	}

	config, err := client.GetGameConfig("123", "Test Game", "Steam", "user1", sysProfile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if config == nil {
		t.Fatal("Expected config, got nil")
	}
	if config.Engine != "d3dmetal" {
		t.Errorf("Expected engine d3dmetal, got %s", config.Engine)
	}
	if config.Confidence != 0.99 {
		t.Errorf("Expected confidence 0.99, got %f", config.Confidence)
	}
}

func TestSupabaseClient_SessionPersistence(t *testing.T) {
	tempDir := t.TempDir()
	client := NewSupabaseClient("http://fake", "anon", tempDir)

	client.AccessToken = "access-token-123"
	client.UserID = "user-123"
	client.saveSession()

	client2 := NewSupabaseClient("http://fake", "anon", tempDir)
	if client2.AccessToken != "access-token-123" {
		t.Errorf("Expected token access-token-123, got %s", client2.AccessToken)
	}
	if client2.UserID != "user-123" {
		t.Errorf("Expected userid user-123, got %s", client2.UserID)
	}

	client2.clearSession()
	if _, err := os.Stat(filepath.Join(tempDir, "supabase_session.json")); !os.IsNotExist(err) {
		t.Errorf("Expected session file to be deleted")
	}
}
