package scraper

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"rift/internal/sysmon"
	"strings"
	"sync"
	"time"

	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"io"
	"net/http"
	"net/url"
)

// BackgroundScraper initializes a headless browser and runs a periodic task
// to scrape gaming wikis for configuration data, populating the local scraper_cache.db.
// It automatically pauses its work if a game is actively running.
type BackgroundScraper struct {
	// Optional http client if needed in future, but we'll just use http.Get
	stopChan  chan struct{}
	queue     []ScrapeJob
	db        *sql.DB
	mu        sync.Mutex
	isRunning bool
	wailsCtx  context.Context
}

type ScrapeJob struct {
	AppID    string
	Platform string
	Name     string
}

// NewBackgroundScraper creates a new BackgroundScraper instance.
func NewBackgroundScraper() *BackgroundScraper {
	home, _ := os.UserHomeDir()
	dbDir := filepath.Join(home, ".rift", "data")
	os.MkdirAll(dbDir, 0755)

	db, err := sql.Open("sqlite3", filepath.Join(dbDir, "scraper_cache.db"))
	if err == nil {
		db.Exec(`CREATE TABLE IF NOT EXISTS translation_profiles (
			app_id TEXT PRIMARY KEY,
			cipher TEXT,
			medium TEXT,
			edition TEXT,
			cadence TEXT,
			depth TEXT
		);`)
	}

	return &BackgroundScraper{
		stopChan: make(chan struct{}),
		queue:    make([]ScrapeJob, 0),
		db:       db,
	}
}

// QueueGame adds a game to the scraper queue.
func (s *BackgroundScraper) QueueGame(appID, platform, name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Skip if already cached
	if s.db != nil {
		var exists int
		s.db.QueryRow("SELECT 1 FROM translation_profiles WHERE app_id = ?", appID).Scan(&exists)
		if exists == 1 {
			return
		}
	}

	// Avoid duplicates in queue
	for _, job := range s.queue {
		if job.AppID == appID {
			return
		}
	}

	s.queue = append(s.queue, ScrapeJob{AppID: appID, Platform: platform, Name: name})
}

// Start begins the background scraping worker loop.
func (s *BackgroundScraper) Start(ctx context.Context) {
	s.wailsCtx = ctx
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = true
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			s.isRunning = false
			s.mu.Unlock()
		}()

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stopChan:
				return
			case <-ticker.C:
				if sysmon.IsGameRunning() {
					// Pause operations while a game is actively running
					continue
				}

				// Example background job: scrape AppleGamingWiki for a game in queue
				s.performScrapeTask()
			}
		}
	}()
}

// Stop gracefully shuts down the scraper.
func (s *BackgroundScraper) Stop() {
	if s.db != nil {
		s.db.Close()
	}
	close(s.stopChan)
}

func (s *BackgroundScraper) performScrapeTask() {
	s.mu.Lock()
	if len(s.queue) == 0 {
		s.mu.Unlock()
		return
	}
	job := s.queue[0]
	s.queue = s.queue[1:]
	s.mu.Unlock()

	// Fallback assignment
	cipher := "Wine + DXVK (DX9/10/11) - Pending"

	// 1. Search for exact PCGamingWiki title using MediaWiki API
	searchURL := fmt.Sprintf("https://www.pcgamingwiki.com/w/api.php?action=query&list=search&srsearch=%s&utf8=&format=json", url.QueryEscape(job.Name))
	resp, err := http.Get(searchURL)
	if err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		var searchRes struct {
			Query struct {
				Search []struct {
					Title string `json:"title"`
				} `json:"search"`
			} `json:"query"`
		}
		json.Unmarshal(body, &searchRes)

		if len(searchRes.Query.Search) > 0 {
			title := searchRes.Query.Search[0].Title

			// 2. Fetch the raw wikitext (bypasses Cloudflare HTML checks)
			wikiURL := fmt.Sprintf("https://www.pcgamingwiki.com/w/api.php?action=query&prop=revisions&rvprop=content&rvslots=main&titles=%s&format=json", url.QueryEscape(title))
			resp2, err2 := http.Get(wikiURL)
			if err2 == nil {
				defer resp2.Body.Close()
				body2, _ := io.ReadAll(resp2.Body)
				textLower := strings.ToLower(string(body2))

				if strings.Contains(textLower, "direct3d 12") || strings.Contains(textLower, "dx12") || strings.Contains(textLower, "vulkan") || strings.Contains(textLower, "d3dmetal") {
					cipher = "Wine + D3DMetal (DX11/12) - Pending"
				}
			}
		}
	}

	fmt.Printf("[Scraper] Processed %s: assigned %s\n", job.Name, cipher)

	// Update SQLite DB
	if s.db != nil {
		s.db.Exec(`INSERT INTO translation_profiles (app_id, cipher, medium, edition, cadence, depth) 
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(app_id) DO UPDATE SET cipher=excluded.cipher`,
			job.AppID, cipher, "Wine-Staging 11.10", "Win10 64-bit", "60 FPS", "—")
	}

	if s.wailsCtx != nil {
		runtime.EventsEmit(s.wailsCtx, "scraper_update")
	}
}
