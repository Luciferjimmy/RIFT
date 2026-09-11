package rift

import (
	"context"
	"os/exec"
	"sync"
	"time"
)

// Game status constants — frontend uses these to decide which buttons to show.
const (
	GameStatusNotDownloaded = "not_downloaded"
	GameStatusDownloading   = "downloading"
	GameStatusReady         = "ready"
	GameStatusRunning       = "running"
)

// downloadTask tracks an in-flight legendary (or Steam) install for a game.
type downloadTask struct {
	GameID    string
	AppID     string
	Platform  string
	Cancel    context.CancelFunc
	Cmd       *exec.Cmd
	StartedAt time.Time
}

// DownloadManager holds the set of active downloads safe for concurrent access.
type DownloadManager struct {
	mu   sync.Mutex
	jobs map[string]*downloadTask
}

// NewDownloadManager creates an empty download manager.
func NewDownloadManager() *DownloadManager {
	return &DownloadManager{jobs: make(map[string]*downloadTask)}
}

// Store registers a new download task.
func (dm *DownloadManager) Store(t *downloadTask) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.jobs[t.GameID] = t
}

// Get retrieves a download task by game ID.
func (dm *DownloadManager) Get(gameID string) *downloadTask {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	return dm.jobs[gameID]
}

// Remove deletes a finished/cancelled download task.
func (dm *DownloadManager) Remove(gameID string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	delete(dm.jobs, gameID)
}

// Cancel cancels the download for the given game ID. Returns true if found.
func (dm *DownloadManager) Cancel(gameID string) bool {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	t, ok := dm.jobs[gameID]
	if ok && t.Cancel != nil {
		t.Cancel()
		delete(dm.jobs, gameID)
	}
	return ok
}

// Snapshot returns a copy of all active download tasks (safe to iterate).
func (dm *DownloadManager) Snapshot() []downloadTask {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	snap := make([]downloadTask, 0, len(dm.jobs))
	for _, t := range dm.jobs {
		snap = append(snap, *t)
	}
	return snap
}

// Elapsed returns the duration since the task started.
func (t *downloadTask) Elapsed() time.Duration {
	if t.StartedAt.IsZero() {
		return 0
	}
	return time.Since(t.StartedAt)
}
