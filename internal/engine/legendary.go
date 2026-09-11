package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// EngineManager handles the setup and installation of isolated engine tools per game.
type EngineManager struct {
	Downloader *Downloader
}

// NewEngineManager creates a new EngineManager.
func NewEngineManager() *EngineManager {
	return &EngineManager{
		Downloader: NewDownloader(),
	}
}

// EnsureLegendary ensures that the legendary binary exists.
// It checks the system PATH first. If missing, it creates a Python 3 virtual environment (venv)
// and installs legendary-gl into the specified targetDir.
func (m *EngineManager) EnsureLegendary(targetDir string) (string, error) {
	// 1. Check if legendary is already installed on the system PATH
	if path, err := exec.LookPath("legendary"); err == nil {
		fmt.Printf("[RIFT Engine] Found system legendary at %s\n", path)
		return path, nil
	}

	// 2. Check if already installed inside our isolated engine folder
	localPath := filepath.Join(targetDir, "venv", "bin", "legendary")
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	fmt.Printf("[RIFT Engine] Legendary not found. Preparing virtual environment at %s...\n", targetDir)

	// Ensure the parent directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create legendary directory: %w", err)
	}

	// 3. Create virtual environment
	venvDir := filepath.Join(targetDir, "venv")
	cmd := exec.Command("python3", "-m", "venv", venvDir)
	if err := cmd.Run(); err != nil {
		// Fallback to checking if python exists
		return "", fmt.Errorf("failed to create python venv (ensure python3 is installed): %w", err)
	}

	// 4. Install legendary-gl via pip
	pipPath := filepath.Join(venvDir, "bin", "pip")
	fmt.Println("[RIFT Engine] Installing legendary-gl inside virtual environment...")
	cmd = exec.Command(pipPath, "install", "--upgrade", "pip", "legendary-gl")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to install legendary-gl: %w", err)
	}

	if _, err := os.Stat(localPath); err != nil {
		return "", fmt.Errorf("legendary binary not found after pip install: %w", err)
	}

	fmt.Println("[RIFT Engine] Successfully installed isolated Legendary binary.")
	return localPath, nil
}
