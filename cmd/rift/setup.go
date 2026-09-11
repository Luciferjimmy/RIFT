package rift

import (
	"os"
	"os/exec"
	goruntime "runtime"
)

// CheckSystemDependencies verifies whether host prerequisites (such as Rosetta 2 on Apple Silicon) are satisfied.
// If Rosetta 2 is missing on Apple Silicon, it automatically initiates installation in the background.
// Third-party package managers (such as Homebrew) are NEVER required as RIFT engines are fully self-contained.
func (a *App) CheckSystemDependencies() bool {
	if goruntime.GOARCH == "arm64" {
		if _, err := os.Stat("/Library/Apple/usr/libexec/oah/libRosettaRuntime"); os.IsNotExist(err) {
			a.logInfo("[System] Rosetta 2 runtime missing on Apple Silicon. Auto-initiating installation in background...")
			go a.EnsureRosetta()
			return false
		}
	}
	return true
}

// CheckRosettaInstalled returns true if Rosetta 2 is present on Apple Silicon, or true if running on Intel.
func (a *App) CheckRosettaInstalled() bool {
	if goruntime.GOARCH != "arm64" {
		return true // Intel Macs do not need Rosetta
	}
	if _, err := os.Stat("/Library/Apple/usr/libexec/oah/libRosettaRuntime"); os.IsNotExist(err) {
		return false
	}
	return true
}

// EnsureRosetta ensures Rosetta 2 is installed on Apple Silicon automatically without manual user steps.
func (a *App) EnsureRosetta() bool {
	if goruntime.GOARCH != "arm64" {
		return true // Intel Macs don't need Rosetta
	}

	if _, err := os.Stat("/Library/Apple/usr/libexec/oah/libRosettaRuntime"); err == nil {
		return true // Already installed
	}

	a.logInfo("[System] Auto-installing Rosetta 2 via softwareupdate...")
	cmd := exec.Command("softwareupdate", "--install-rosetta", "--agree-to-license")
	if out, err := cmd.CombinedOutput(); err != nil {
		a.logWarn("[System] Standard Rosetta install returned: %v (%s). Trying admin execution...", err, string(out))
		script := `do shell script "softwareupdate --install-rosetta --agree-to-license" with administrator privileges`
		cmdAdmin := exec.Command("osascript", "-e", script)
		if adminOut, adminErr := cmdAdmin.CombinedOutput(); adminErr != nil {
			a.logError("[System] Failed to auto-install Rosetta 2: %v (%s)", adminErr, string(adminOut))
			return false
		}
	}

	// Verify installation
	if _, err := os.Stat("/Library/Apple/usr/libexec/oah/libRosettaRuntime"); err == nil {
		a.logInfo("[System] Rosetta 2 successfully verified and ready!")
		return true
	}
	return false
}

