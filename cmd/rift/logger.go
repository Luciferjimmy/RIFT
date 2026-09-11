package rift

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

// RiftLogger writes structured, timestamped logs to ~/.rift/logs/session.log
// using standard log/slog for better I/O performance.
type RiftLogger struct {
	mu      sync.Mutex
	file    *os.File
	logger  *slog.Logger
	enabled bool
}

// InitLogger creates (or returns) the singleton session logger.
func InitLogger() *RiftLogger {
	home, _ := os.UserHomeDir()
	logDir := filepath.Join(home, ".rift", "logs")
	os.MkdirAll(logDir, 0755)

	logPath := filepath.Join(logDir, "session.log")
	prevPath := filepath.Join(logDir, "session_prev.log")

	// Rotate previous session
	if _, err := os.Stat(logPath); err == nil {
		os.Rename(logPath, prevPath)
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[RiftLogger] FATAL: cannot open log file: %v\n", err)
		return &RiftLogger{enabled: false}
	}

	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewTextHandler(f, opts)
	slogger := slog.New(handler)

	l := &RiftLogger{file: f, logger: slogger, enabled: true}
	l.Info("=== RIFT Session Started ===")
	l.Info("Log file: %s", logPath)
	return l
}

func (l *RiftLogger) Info(format string, args ...interface{}) {
	if l != nil && l.enabled {
		msg := fmt.Sprintf(format, args...)
		l.logger.Info(msg)
		fmt.Println("[INFO] " + msg)
	}
}

func (l *RiftLogger) Warn(format string, args ...interface{}) {
	if l != nil && l.enabled {
		msg := fmt.Sprintf(format, args...)
		l.logger.Warn(msg)
		fmt.Println("[WARN] " + msg)
	}
}

func (l *RiftLogger) Error(format string, args ...interface{}) {
	if l != nil && l.enabled {
		msg := fmt.Sprintf(format, args...)
		l.logger.Error(msg)
		fmt.Println("[ERROR] " + msg)
	}
}

func (l *RiftLogger) Debug(format string, args ...interface{}) {
	if l != nil && l.enabled {
		msg := fmt.Sprintf(format, args...)
		l.logger.Debug(msg)
	}
}

func (l *RiftLogger) Close() {
	if l == nil || !l.enabled {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Info("=== RIFT Session Ended ===")
	l.file.Sync() // Only sync on close
	l.file.Close()
	l.enabled = false
}
