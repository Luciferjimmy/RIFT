package rift

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// UnlockAchievement securely inserts a new achievement record for the user.
func (a *App) UnlockAchievement(achievementID string) string {
	if !a.Supabase.IsLoggedIn() {
		return `{"error": "Not logged in"}`
	}

	payload := map[string]string{
		"target_achievement_id": achievementID,
	}
	bodyData, _ := json.Marshal(payload)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", a.Supabase.ProjectURL+"/rest/v1/rpc/unlock_achievement", bytes.NewBuffer(bodyData))
	req.Header.Set("apikey", a.Supabase.AnonKey)
	req.Header.Set("Authorization", "Bearer "+a.Supabase.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return `{"success": false, "error": "Network failure"}`
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		if refreshErr := a.Supabase.RefreshSession(); refreshErr == nil {
			req.Header.Set("Authorization", "Bearer "+a.Supabase.AccessToken)
			resp, _ = client.Do(req)
			if resp != nil {
				defer resp.Body.Close()
			}
		}
	}

	if resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		a.logInfo("[Achievements] Successfully unlocked: %s", achievementID)
		return `{"success": true}`
	}
	
	a.logError("[Achievements] Failed to unlock %s (Status: %d)", achievementID, resp.StatusCode)
	return `{"success": false}`
}
