package rift

import (
	"fmt"
	"os/exec"
)

// ActivateGameMode boosts the game process priority
// to maximize available CPU and memory for the game using macOS taskpolicy.
func (a *App) ActivateGameMode(pid int) {
	a.logInfo("[GameMode] Activating — ensuring PID %d is prioritized", pid)

	// Attempt to clear any background/utility state and set highest QoS via taskpolicy
	// taskpolicy -B clears background state
	cmd1 := exec.Command("taskpolicy", "-B", "-p", fmt.Sprint(pid))
	if err := cmd1.Run(); err != nil {
		a.logInfo("[GameMode] Notice: Failed to clear background state for PID %d: %v", pid, err)
	}

	// Attempt to ensure nice value is 0 (normal priority). We cannot use negative nice without sudo.
	cmd2 := exec.Command("renice", "-n", "0", "-p", fmt.Sprint(pid))
	if err := cmd2.Run(); err != nil {
		a.logInfo("[GameMode] Notice: Failed to renice PID %d: %v", pid, err)
	}

	a.logInfo("[GameMode] To activate Apple's Native Game Mode (which pauses background tasks and boosts GPU priority), please ensure the game is set to FULLSCREEN in its display settings.")
	a.logInfo("[GameMode] Activated — PID %d prioritization complete", pid)
}

// DeactivateGameMode restores normal system behavior after the game exits.
func (a *App) DeactivateGameMode() {
	a.logInfo("[GameMode] Deactivating — system returning to normal state")
}

// preFlightMemoryOptimize attempts to maximize available RAM before a game launches.
func (a *App) preFlightMemoryOptimize() {
	a.logInfo("[PerfEngine] Running pre-flight memory optimization...")

	// 1. We cannot run 'purge' without sudo on macOS.
	a.logInfo("[PerfEngine] Memory purge skipped (requires sudo). To free maximum memory, close unused background apps.")
}
