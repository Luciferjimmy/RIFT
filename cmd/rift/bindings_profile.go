package rift

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"hash/fnv"
	"os/exec"
	"strings"
)

// ProfileData represents the combined data sent to the frontend
type ProfileData struct {
	Alias           string         `json:"alias"` // Sourced from Auth table
	Clearance       string         `json:"clearance"`
	RiftID          string         `json:"rift_id"` // Unique deterministic 6-digit ID
	OSTarget        string         `json:"os_target"` // Fetched via sw_vers
	Chip            string         `json:"chip"` // Real detected chip from sysctl
	PlaytimeMinutes int            `json:"playtime_minutes"`
	GamesOwned      int            `json:"games_owned"`
	GamesPlayed     int            `json:"games_played"`
	CurrentStreak   int            `json:"current_streak"`
	JoinDate        string         `json:"join_date"`
	Heatmap         map[string]int `json:"heatmap"`
	Achievements    []Achievement  `json:"achievements"`
	AllAchievements []Achievement  `json:"all_achievements"`
	Transmissions   []Transmission `json:"transmissions"`
}

type Transmission struct {
	ID         string `json:"id"`
	VersionTag string `json:"version_tag"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Date       string `json:"date"`
}

type Achievement struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IconTag     string `json:"icon_tag"`
	Rarity      string `json:"rarity"`
	ColorHex    string `json:"color_hex"`
	UnlockedAt  string `json:"unlocked_at"`
	IsUnlocked  bool   `json:"is_unlocked"`
}

// GetUserProfileData fetches the user's extended profile and achievements from Supabase
func (a *App) GetUserProfileData() string {
	a.logInfo("[Profile] Fetching extended profile data...")

	if !a.Supabase.IsLoggedIn() {
		return `{"error": "Not logged in"}`
	}

	client := &http.Client{}

	// 1. Fetch Profile Data
	profileReq, _ := http.NewRequest("GET", a.Supabase.ProjectURL+"/rest/v1/user_profiles?select=clearance,playtime_minutes,games_owned,games_played,current_streak,created_at,activity_heatmap&user_id=eq."+a.Supabase.UserID, nil)
	profileReq.Header.Set("apikey", a.Supabase.AnonKey)
	profileReq.Header.Set("Authorization", "Bearer "+a.Supabase.AccessToken)

	profileResp, err := client.Do(profileReq)
	if err != nil {
		a.logError("[Profile] Failed to fetch profile: %v", err)
		return `{"error": "Network error"}`
	}

	// Automatic Token Refresh on Expiry
	if profileResp.StatusCode == 401 {
		profileResp.Body.Close()
		a.logInfo("[Profile] JWT Expired. Refreshing session...")
		if refreshErr := a.Supabase.RefreshSession(); refreshErr == nil {
			profileReq.Header.Set("Authorization", "Bearer "+a.Supabase.AccessToken)
			profileResp, err = client.Do(profileReq)
			if err != nil {
				return `{"error": "Network error on retry"}`
			}
		} else {
			a.logError("[Profile] Failed to refresh token: %v", refreshErr)
			return `{"error": "Session expired. Please log in again."}`
		}
	}
	defer profileResp.Body.Close()

	body, _ := io.ReadAll(profileResp.Body)
	
	// Dynamically fetch actual macOS version and chip
	osTarget := "macOS"
	if out, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
		osTarget = "macOS " + strings.TrimSpace(string(out))
	}

	chip := "Apple Silicon"
	if out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
		chip = strings.TrimSpace(string(out))
	}

	data := ProfileData{
		Alias:           a.Supabase.UserName,
		Clearance:       "Operator / Tier 3",
		RiftID:          "",
		OSTarget:        osTarget,
		Chip:            chip,
		PlaytimeMinutes: 0,
		GamesOwned:      0,
		GamesPlayed:     0,
		CurrentStreak:   0,
		JoinDate:        "Unknown",
		Heatmap:         make(map[string]int),
		Achievements:    []Achievement{},
	}

	if profileResp.StatusCode != 200 {
		// Output the Supabase error directly to the UI so we can debug it
		data.Clearance = fmt.Sprintf("API Error: %d - %s", profileResp.StatusCode, string(body))
	} else {
		var profileRows []map[string]interface{}
		json.Unmarshal(body, &profileRows)

		if len(profileRows) > 0 {
			row := profileRows[0]
			if val, ok := row["clearance"].(string); ok {
				data.Clearance = val
			}
			if val, ok := row["playtime_minutes"].(float64); ok {
				data.PlaytimeMinutes = int(val)
			}
			if val, ok := row["games_owned"].(float64); ok {
				data.GamesOwned = int(val)
			}
			if val, ok := row["games_played"].(float64); ok {
				data.GamesPlayed = int(val)
			}
			if val, ok := row["current_streak"].(float64); ok {
				data.CurrentStreak = int(val)
			}
			if val, ok := row["created_at"].(string); ok && len(val) >= 10 {
				data.JoinDate = val[:10]
			}
			if val, ok := row["activity_heatmap"].(map[string]interface{}); ok {
				for dateKey, countVal := range val {
					if fCount, ok := countVal.(float64); ok {
						data.Heatmap[dateKey] = int(fCount)
					}
				}
			}
		} else {
		    data.Clearance = "Error: 0 Rows found for ID: " + a.Supabase.UserID
		}
	}

	// 2. Dynamic ID Generation based on Clearance Prefix
	prefix := "OP"
	clearanceLower := strings.ToLower(data.Clearance)
	if strings.Contains(clearanceLower, "founder") {
		prefix = "F"
	} else if strings.Contains(clearanceLower, "dev") {
		prefix = "DEV"
	}

	h := fnv.New32a()
	h.Write([]byte(a.Supabase.UserID))
	deterministicID := (h.Sum32() % 900000) + 100000

	if strings.HasPrefix(strings.ToLower(a.Supabase.UserName), "abhinaw") {
		data.RiftID = fmt.Sprintf("RIFT-%s-000000", prefix)
	} else {
		data.RiftID = fmt.Sprintf("RIFT-%s-%d", prefix, deterministicID)
	}

	// 3. Fetch Unlocked Achievements
	achReq, _ := http.NewRequest("GET", a.Supabase.ProjectURL+"/rest/v1/user_achievements?select=unlocked_at,achievements_dict(id,title,description,icon_tag,rarity,color_hex)&user_id=eq."+a.Supabase.UserID, nil)
	achReq.Header.Set("apikey", a.Supabase.AnonKey)
	achReq.Header.Set("Authorization", "Bearer "+a.Supabase.AccessToken)

	achResp, err := client.Do(achReq)
	if err == nil {
		defer achResp.Body.Close()
		var achRows []map[string]interface{}
		achBody, _ := io.ReadAll(achResp.Body)
		json.Unmarshal(achBody, &achRows)

		for _, row := range achRows {
			dictObj, ok := row["achievements_dict"].(map[string]interface{})
			if !ok {
				continue
			}
			ach := Achievement{
				UnlockedAt: fmt.Sprintf("%v", row["unlocked_at"]),
				ID:         fmt.Sprintf("%v", dictObj["id"]),
				Title:      fmt.Sprintf("%v", dictObj["title"]),
				Description: fmt.Sprintf("%v", dictObj["description"]),
				IconTag:    fmt.Sprintf("%v", dictObj["icon_tag"]),
				Rarity:     fmt.Sprintf("%v", dictObj["rarity"]),
				ColorHex:   fmt.Sprintf("%v", dictObj["color_hex"]),
				IsUnlocked: true,
			}
			data.Achievements = append(data.Achievements, ach)
		}
	}

	// 4. Fetch All Achievements from Dictionary
	dictReq, _ := http.NewRequest("GET", a.Supabase.ProjectURL+"/rest/v1/achievements_dict?select=id,title,description,icon_tag,rarity,color_hex", nil)
	dictReq.Header.Set("apikey", a.Supabase.AnonKey)
	dictReq.Header.Set("Authorization", "Bearer "+a.Supabase.AccessToken)

	dictResp, err := client.Do(dictReq)
	if err == nil {
		defer dictResp.Body.Close()
		var dictRows []map[string]interface{}
		dictBody, _ := io.ReadAll(dictResp.Body)
		json.Unmarshal(dictBody, &dictRows)

		for _, row := range dictRows {
			achID := fmt.Sprintf("%v", row["id"])
			isUnlocked := false
			var unlockedAt string
			for _, unlocked := range data.Achievements {
				if unlocked.ID == achID {
					isUnlocked = true
					unlockedAt = unlocked.UnlockedAt
					break
				}
			}

			ach := Achievement{
				ID:          achID,
				Title:       fmt.Sprintf("%v", row["title"]),
				Description: fmt.Sprintf("%v", row["description"]),
				IconTag:     fmt.Sprintf("%v", row["icon_tag"]),
				Rarity:      fmt.Sprintf("%v", row["rarity"]),
				ColorHex:    fmt.Sprintf("%v", row["color_hex"]),
				UnlockedAt:  unlockedAt,
				IsUnlocked:  isUnlocked,
			}
			data.AllAchievements = append(data.AllAchievements, ach)
		}
	}

	// 5. Fetch Transmissions (Top 15)
	transReq, _ := http.NewRequest("GET", a.Supabase.ProjectURL+"/rest/v1/system_transmissions?select=*&order=created_at.desc&limit=15", nil)
	transReq.Header.Set("apikey", a.Supabase.AnonKey)
	transReq.Header.Set("Authorization", "Bearer "+a.Supabase.AccessToken)

	transResp, err := client.Do(transReq)
	if err == nil {
		defer transResp.Body.Close()
		transBody, _ := io.ReadAll(transResp.Body)
		
		if transResp.StatusCode == 200 {
			var transRows []map[string]interface{}
			json.Unmarshal(transBody, &transRows)

			for _, row := range transRows {
				dateStr := "Unknown"
				if val, ok := row["created_at"].(string); ok && len(val) >= 10 {
					dateStr = val[:10]
				}
				t := Transmission{
					ID:         fmt.Sprintf("%v", row["id"]),
					VersionTag: fmt.Sprintf("%v", row["version_tag"]),
					Title:      fmt.Sprintf("%v", row["title"]),
					Content:    fmt.Sprintf("%v", row["content"]),
					Date:       dateStr,
				}
				data.Transmissions = append(data.Transmissions, t)
			}
		}
	}

	out, _ := json.Marshal(data)
	return string(out)
}

// LogGameActivity hits a Supabase RPC to increment activity counters for today using Real Dates
func (a *App) LogGameActivity() string {
	if !a.Supabase.IsLoggedIn() {
		return `{"error": "Not logged in"}`
	}
	
	client := &http.Client{}
	req, _ := http.NewRequest("POST", a.Supabase.ProjectURL+"/rest/v1/rpc/log_game_activity", nil)
	req.Header.Set("apikey", a.Supabase.AnonKey)
	req.Header.Set("Authorization", "Bearer "+a.Supabase.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return `{"success": false}`
	}
	
	if resp.StatusCode == 401 {
		resp.Body.Close()
		if refreshErr := a.Supabase.RefreshSession(); refreshErr == nil {
			req.Header.Set("Authorization", "Bearer "+a.Supabase.AccessToken)
			resp, err = client.Do(req)
			if err != nil {
				return `{"success": false}`
			}
		}
	}
	
	if resp.StatusCode >= 400 {
		return `{"success": false}`
	}
	return `{"success": true}`
}
