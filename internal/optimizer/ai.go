package optimizer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	groqEndpoint = "https://api.groq.com/openai/v1/chat/completions"
	groqModel    = "llama-3.3-70b-versatile"
	groqTimeout  = 15 * time.Second
)

// AIReasoner uses an LLM to fill gaps that the deterministic rules engine
// could not resolve. It is designed to be resilient: if the API is
// unavailable, rate-limited, or returns nonsense the pipeline degrades
// gracefully.
type AIReasoner struct {
	apiKey string
	client *http.Client
}

// NewAIReasoner creates a reasoner that uses the given Groq API key.
func NewAIReasoner(apiKey string) *AIReasoner {
	return &AIReasoner{
		apiKey: apiKey,
		client: &http.Client{Timeout: groqTimeout},
	}
}

// AIFillResult is the structured output returned by the AI.
type AIFillResult struct {
	TranslationLayer string   `json:"translation_layer"` // D3DMetal | DXVK
	WineVersion      string   `json:"wine_version"`
	WindowsVersion   string   `json:"windows_version"`
	DLLOverrides     []string `json:"dll_overrides"`
	Winetricks       []string `json:"winetricks"`
	EnvVars          []string `json:"env_vars"` // "KEY=VALUE" pairs
	LaunchArgs       string   `json:"launch_args"`
	Performance      string   `json:"performance"` // Great | Good | Okay | Poor
	KnownIssues      string   `json:"known_issues"`
	Confidence       float64  `json:"confidence"` // 0.0–1.0 how sure the AI is
}

// Fill attempts to obtain AI suggestions for the game. If the API call fails
// the method returns nil — the caller should fall back to the rules-engine
// config.
func (r *AIReasoner) Fill(ctx context.Context, gameName, engine, steamID string,
	data *AggregatedGameData, sys *SystemProfile) *AIFillResult {

	if r.apiKey == "" {
		return nil // no key configured — skip AI
	}

	prompt := r.buildPrompt(gameName, engine, steamID, data, sys)
	result, err := r.callGroq(ctx, prompt)
	if err != nil {
		return nil // caller falls back to rules
	}
	return result
}

// ---------------------------------------------------------------------------
// Prompt construction
// ---------------------------------------------------------------------------

func (r *AIReasoner) buildPrompt(gameName, engine, steamID string,
	data *AggregatedGameData, sys *SystemProfile) string {

	var b strings.Builder
	b.WriteString(`You are a game-compatibility specialist for RIFT, a macOS game launcher that runs Windows games via Wine. Your task is to fill in the JSON template below with the best configuration for the given game on the given system.

CRITICAL HARDWARE RULE: If the user's system RAM is 8GB or less, D3DMetal (GPTK) often causes severe memory swapping/OOM crashes in DX12 games. If the game supports DX11, you MUST prefer DXVK and supply the "-dx11" launch argument to drastically reduce VRAM overhead. Only use D3DMetal for 8GB Macs if the game is strictly DX12-only.

Only fill fields you are confident about. Use "SKIP" for anything uncertain. Be concise.

`)
	// Game identity.
	fmt.Fprintf(&b, "GAME: %s\n", gameName)
	if steamID != "" {
		fmt.Fprintf(&b, "STEAM APPID: %s\n", steamID)
	}
	// Engine.
	eng := engine
	if eng == "" {
		eng = "unknown"
	}
	fmt.Fprintf(&b, "ENGINE: %s\n", eng)

	// System.
	fmt.Fprintf(&b, "SYSTEM: macOS %s, Apple Silicon=%v, Metal GPU Family=%d, RAM=%dGB, GPTK available=%v, DXVK available=%v\n",
		sys.MacOSVersion, sys.AppleSilicon, sys.MetalGPUFamily, sys.RAMGB,
		sys.Engines.GPTK, sys.Engines.DXVK)

	// ProtonDB evidence.
	if data.ProtonDBBest != nil {
		pb := data.ProtonDBBest
		fmt.Fprintf(&b, "\nPROTONDB REPORTS (top):\n  rating=%s version=%s\n  launch_args=%s env_vars=%s winetricks=%s\n  notes=%s\n",
			pb.Rating, pb.ProtonVersion, pb.LaunchArgs, pb.EnvVars, pb.Winetricks, truncate(pb.NotesRaw, 300))
	}

	// Lutris evidence.
	if data.LutrisBest != nil {
		ls := data.LutrisBest
		fmt.Fprintf(&b, "\nLUTRIS INSTALL SCRIPT:\n  runner=%s version=%s executable=%s\n",
			ls.Runner, ls.Version, ls.Executable)
		if ls.RawScript != "" {
			fmt.Fprintf(&b, "  raw_script (first 800 chars):\n%s\n", truncate(ls.RawScript, 800))
		}
	}

	// PCGamingWiki evidence.
	if data.PCGamingWiki != nil {
		pg := data.PCGamingWiki
		fmt.Fprintf(&b, "\nPCGAMINGWIKI:\n  developer=%s publisher=%s engine=%s\n",
			pg.Developer, pg.Publisher, pg.Engine)
	}

	b.WriteString(`

Respond ONLY with a JSON object in the following format (no markdown, no explanation):

{
  "translation_layer": "D3DMetal" | "DXVK" | "SKIP",
  "wine_version": "latest" | "specific version" | "SKIP",
  "windows_version": "win10" | "win8" | "win7" | "SKIP",
  "dll_overrides": ["d3d11=n,b", ...] | [],
  "winetricks": ["vcrun2022", ...] | [],
  "env_vars": ["WINEMSYNC=1", ...] | [],
  "launch_args": "-dx11" | "" | "SKIP",
  "performance": "Great" | "Good" | "Okay" | "Poor",
  "known_issues": "brief note or empty string",
  "confidence": 0.8
}

`)
	return b.String()
}

// ---------------------------------------------------------------------------
// Groq API call
// ---------------------------------------------------------------------------

func (r *AIReasoner) callGroq(ctx context.Context, prompt string) (*AIFillResult, error) {
	body := map[string]interface{}{
		"model": groqModel,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a precise game compatibility analyst. Output only valid JSON."},
			{"role": "user", "content": prompt},
		},
		"temperature":     0.3,
		"max_tokens":      600,
		"response_format": map[string]string{"type": "json_object"},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, groqEndpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limited (429)")
	}
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API status %d: %s", resp.StatusCode, truncate(string(respBody), 200))
	}

	var raw struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if len(raw.Choices) == 0 {
		return nil, fmt.Errorf("empty response choices")
	}

	content := raw.Choices[0].Message.Content
	var result AIFillResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("unmarshal AI JSON: %w", err)
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
