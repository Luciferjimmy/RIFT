package cloud

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"rift/internal/types"
	"strconv"
	"strings"
	"time"
)

// ClientVersion represents the active client build sent in request headers.
const ClientVersion = "1.0.0"

type SupabaseClient struct {
	ProjectURL   string
	AnonKey      string
	AccessToken  string
	RefreshToken string
	UserID       string
	UserEmail    string
	UserName     string
	AuthDir      string
}

// Default production Supabase credentials for standalone client distributions.
const (
	DefaultSupabaseURL     = "https://iuhsscbeaagbqlerumsu.supabase.co"
	DefaultSupabaseAnonKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Iml1aHNzY2JlYWFnYnFsZXJ1bXN1Iiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODQwOTcxMzksImV4cCI6MjA5OTY3MzEzOX0.N7DLNaa96d_im3zCbVQj6T3lIfSYkwVRSHof5BEFGzk"
)

func NewSupabaseClient(projectURL, anonKey, authDir string) *SupabaseClient {
	if projectURL == "" {
		projectURL = DefaultSupabaseURL
	}
	if anonKey == "" {
		anonKey = DefaultSupabaseAnonKey
	}
	c := &SupabaseClient{
		ProjectURL: projectURL,
		AnonKey:    anonKey,
		AuthDir:    authDir,
	}
	c.loadSession()
	return c
}

func (s *SupabaseClient) IsLoggedIn() bool {
	return s.AccessToken != ""
}

func (s *SupabaseClient) loadSession() {
	path := filepath.Join(s.AuthDir, "supabase_session.json")
	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, s); err != nil {
			// Corrupted session, wipe it
			s.clearSession()
		}
	}
}

func (s *SupabaseClient) saveSession() {
	os.MkdirAll(s.AuthDir, 0700)
	path := filepath.Join(s.AuthDir, "supabase_session.json")
	data, _ := json.MarshalIndent(s, "", "  ")
	os.WriteFile(path, data, 0600)
}

func (s *SupabaseClient) clearSession() {
	s.AccessToken = ""
	s.RefreshToken = ""
	s.UserID = ""
	s.UserEmail = ""
	s.UserName = ""
	path := filepath.Join(s.AuthDir, "supabase_session.json")
	os.Remove(path)
}

// Auth response structs
type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         struct {
		ID           string                 `json:"id"`
		Email        string                 `json:"email"`
		UserMetadata map[string]interface{} `json:"user_metadata"`
	} `json:"user"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	Message          string `json:"msg"`
}

func (s *SupabaseClient) doAuthReq(endpoint string, reqBody interface{}) (*authResponse, error) {
	url := fmt.Sprintf("%s/auth/v1/%s", s.ProjectURL, endpoint)
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("apikey", s.AnonKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-RIFT-Version", ClientVersion)
	req.Header.Set("X-RIFT-Client", "desktop-mac")

	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result authResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Error != "" || result.ErrorDescription != "" {
		msg := result.ErrorDescription
		if msg == "" {
			msg = result.Error
		}
		if msg == "" {
			msg = result.Message
		}
		return nil, errors.New(msg)
	}
	if result.Message != "" {
		return nil, errors.New(result.Message)
	}

	return &result, nil
}

func (s *SupabaseClient) SignUp(email, name, password string) error {
	reqBody := map[string]interface{}{
		"email":    email,
		"password": password,
		"data": map[string]string{
			"name": name,
		},
		"gotrue_meta_security": map[string]string{}, // Provide empty to avoid strict checks if needed
	}
	
	// Add redirect_to so it doesn't default to localhost:3000
	// Users will be redirected to the GitHub repo upon email verification
	reqBody["gotrue_meta_security"] = map[string]string{"redirect_to": "https://github.com/Luciferjimmy/RIFT#success"}
	res, err := s.doAuthReq("signup", reqBody)
	if err != nil {
		return err
	}
	if res.AccessToken != "" {
		s.AccessToken = res.AccessToken
		s.RefreshToken = res.RefreshToken
		s.UserID = res.User.ID
		s.UserEmail = res.User.Email
		if n, ok := res.User.UserMetadata["name"].(string); ok {
			s.UserName = n
		}
		s.saveSession()
	}
	return nil
}

func (s *SupabaseClient) SignIn(email, password string) error {
	res, err := s.doAuthReq("token?grant_type=password", map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return err
	}
	s.AccessToken = res.AccessToken
	s.RefreshToken = res.RefreshToken
	s.UserID = res.User.ID
	s.UserEmail = res.User.Email
	if n, ok := res.User.UserMetadata["name"].(string); ok {
		s.UserName = n
	}
	s.saveSession()
	return nil
}

// RefreshSession uses the stored refresh token to obtain a new access token
// from Supabase. This prevents 401 UNAUTHORIZED_ASYMMETRIC_JWT errors when
// the original JWT expires (default: 1 hour).
func (s *SupabaseClient) RefreshSession() error {
	if s.RefreshToken == "" {
		return errors.New("no refresh token available — user must sign in again")
	}

	res, err := s.doAuthReq("token?grant_type=refresh_token", map[string]string{
		"refresh_token": s.RefreshToken,
	})
	if err != nil {
		return fmt.Errorf("token refresh failed: %w", err)
	}

	if res.AccessToken == "" {
		return errors.New("token refresh returned empty access token")
	}

	s.AccessToken = res.AccessToken
	s.RefreshToken = res.RefreshToken
	if res.User.ID != "" {
		s.UserID = res.User.ID
	}
	if res.User.Email != "" {
		s.UserEmail = res.User.Email
	}
	if n, ok := res.User.UserMetadata["name"].(string); ok {
		s.UserName = n
	}
	s.saveSession()
	return nil
}

func (s *SupabaseClient) SignOut() error {
	if s.AccessToken == "" {
		return nil
	}
	url := fmt.Sprintf("%s/auth/v1/logout", s.ProjectURL)
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("apikey", s.AnonKey)
	req.Header.Set("Authorization", "Bearer "+s.AccessToken)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	client.Do(req)

	s.clearSession()
	return nil
}

// ResetPasswordForEmail sends a 6-digit recovery OTP code to the user's email.
func (s *SupabaseClient) ResetPasswordForEmail(email string) error {
	reqBody := map[string]interface{}{
		"email": strings.TrimSpace(email),
	}
	_, err := s.doAuthReq("recover", reqBody)
	return err
}

// VerifyResetCodeAndSetPassword verifies the 6-digit recovery code and updates the user's password in-app.
func (s *SupabaseClient) VerifyResetCodeAndSetPassword(email, code, newPassword string) error {
	// 1. Verify recovery OTP code
	verifyBody := map[string]interface{}{
		"type":  "recovery",
		"email": strings.TrimSpace(email),
		"token": strings.TrimSpace(code),
	}
	res, err := s.doAuthReq("verify", verifyBody)
	if err != nil {
		return fmt.Errorf("invalid verification code: %w", err)
	}

	if res.AccessToken == "" {
		return errors.New("verification succeeded but no recovery token returned")
	}

	// 2. Update the user password via Supabase User API
	url := fmt.Sprintf("%s/auth/v1/user", s.ProjectURL)
	updateBody, _ := json.Marshal(map[string]string{
		"password": newPassword,
	})

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(updateBody))
	if err != nil {
		return fmt.Errorf("failed to create update request: %w", err)
	}
	req.Header.Set("apikey", s.AnonKey)
	req.Header.Set("Authorization", "Bearer "+res.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("password update rejected: %s", string(body))
	}

	// 3. Save new active session
	s.AccessToken = res.AccessToken
	s.RefreshToken = res.RefreshToken
	s.UserID = res.User.ID
	s.UserEmail = res.User.Email
	if n, ok := res.User.UserMetadata["name"].(string); ok {
		s.UserName = n
	}
	s.saveSession()

	return nil
}

// UpdatePassword updates the password for the currently logged in user session.
func (s *SupabaseClient) UpdatePassword(newPassword string) error {
	if s.AccessToken == "" {
		return errors.New("user is not logged in")
	}

	// Proactively refresh session if a refresh token is present so JWT is never stale
	if s.RefreshToken != "" {
		_ = s.RefreshSession()
	}

	url := fmt.Sprintf("%s/auth/v1/user", s.ProjectURL)
	updateBody, _ := json.Marshal(map[string]string{
		"password": newPassword,
	})

	client := &http.Client{Timeout: 15 * time.Second}

	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequest("PUT", url, bytes.NewBuffer(updateBody))
		if err != nil {
			return fmt.Errorf("failed to create update request: %w", err)
		}
		req.Header.Set("apikey", s.AnonKey)
		req.Header.Set("Authorization", "Bearer "+s.AccessToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		// If token expired, refresh and retry once
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			if refreshErr := s.RefreshSession(); refreshErr == nil {
				continue
			}
		}

		return fmt.Errorf("password update rejected: %s", string(body))
	}

	return errors.New("password update failed after token refresh")
}

// EdgeFunctionResponse wraps the edge function response
type EdgeFunctionResponse struct {
	Source string                `json:"source"`
	Config types.CloudGameConfig `json:"config"`
}

func (s *SupabaseClient) GetGameConfig(appID, gameName, platform, steamID string, sysProfile types.SystemProfile) (*types.CloudGameConfig, error) {
	if !s.IsLoggedIn() {
		return nil, errors.New("not logged in")
	}

	url := fmt.Sprintf("%s/functions/v1/get-game-config", s.ProjectURL)

	// Determine base engine hint for edge function
	engineHint := "d3dmetal"
	if isMacOSVersionOlderThan(sysProfile.MacOSVersion, 14) || !sysProfile.HasGPTK {
		engineHint = "dxvk"
	}

	reqBody := map[string]interface{}{
		"app_id":      appID,
		"game_name":   gameName,
		"platform":    platform,
		"engine":      engineHint,
		"sys_profile": sysProfile,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	// Perform the request, with automatic token refresh on 401
	for attempt := 0; attempt < 2; attempt++ {
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+s.AccessToken)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-RIFT-Version", ClientVersion)
		req.Header.Set("X-RIFT-Client", "desktop-mac")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 401 && attempt == 0 {
			// Token expired — try refreshing and retry once
			resp.Body.Close()
			if refreshErr := s.RefreshSession(); refreshErr != nil {
				return nil, fmt.Errorf("edge function returned 401 and token refresh failed: %w", refreshErr)
			}
			continue
		}

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("edge function error: %d %s", resp.StatusCode, string(body))
		}

		var result EdgeFunctionResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		return &result.Config, nil
	}

	return nil, errors.New("edge function request failed after token refresh")
}

// isMacOSVersionOlderThan parses major version and checks if it is strictly less than targetMajor
func isMacOSVersionOlderThan(v string, targetMajor int) bool {
	parts := strings.Split(v, ".")
	if len(parts) > 0 {
		if major, err := strconv.Atoi(parts[0]); err == nil {
			return major < targetMajor
		}
	}
	return false
}
