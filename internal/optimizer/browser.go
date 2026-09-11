package optimizer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// BrowserEnricher performs targeted web scraping for games where local data
// is insufficient (< 0.4 confidence). It is designed to be conservative:
// timeouts are short, and failures degrade gracefully.
type BrowserEnricher struct {
	client *http.Client
}

// NewBrowserEnricher creates an enricher with sensible timeouts.
func NewBrowserEnricher() *BrowserEnricher {
	return &BrowserEnricher{
		client: &http.Client{Timeout: 8 * time.Second},
	}
}

// EnrichResult holds any scraping results together with a before/after
// confidence delta.
type EnrichResult struct {
	ProtonDBHint    string // latest user report extracted from ProtonDB page
	RawSuggestions  string // raw text from Reddit / search
	ConfidenceDelta float64
	Log             []string
}

// Enrich attempts to scrape the web for the given game. It returns nil
// when scraping fails or when nothing useful was found — the pipeline
// should continue with its existing config.
func (e *BrowserEnricher) Enrich(ctx context.Context, ident *GameIdentity) *EnrichResult {
	res := &EnrichResult{}

	// Scrape ProtonDB live page for newer reports.
	if ident.SteamAppID != "" {
		hint, err := e.scrapeProtonDB(ctx, ident.SteamAppID)
		if err == nil && hint != "" {
			res.ProtonDBHint = hint
			res.ConfidenceDelta += 0.15
			res.Log = append(res.Log, fmt.Sprintf("Browser: ProtonDB live scrape returned data for %s", ident.SteamAppID))
		}
	}

	return res
}

// scrapeProtonDB fetches the game page on ProtonDB and extracts the
// top-level compatibility summary.
func (e *BrowserEnricher) scrapeProtonDB(ctx context.Context, steamID string) (string, error) {
	url := fmt.Sprintf("https://www.protondb.com/app/%s", steamID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) RIFT/1.0")
	req.Header.Set("Accept", "text/html")

	resp, err := e.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Extract summary from the page. ProtonDB pages contain JSON-LD or
	// meta tags with the rating. We look for the "summary" class or
	// JSON-LD script tags containing "protonLabel".
	html := string(body)

	// Try JSON-LD first.
	if idx := strings.Index(html, `"protonLabel"`); idx > 0 {
		start := idx + 14 // len(`"protonLabel":"`)
		end := strings.Index(html[start:], `"`)
		if end > 0 {
			return "ProtonDB rating: " + html[start:start+end], nil
		}
	}

	// Fallback: search for the rating text in the page title or meta.
	if strings.Contains(html, "Native") && !strings.Contains(html, "not Native") {
		return "ProtonDB: Native macOS support", nil
	}
	if strings.Contains(html, "Platinum") {
		return "ProtonDB: Platinum", nil
	}
	if strings.Contains(html, "Gold") {
		return "ProtonDB: Gold", nil
	}

	return "", fmt.Errorf("no rating found on page")
}
