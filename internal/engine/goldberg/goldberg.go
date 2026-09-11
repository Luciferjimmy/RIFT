package goldberg

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// InjectSteamEmulator scans the given game directory for Steam DRM DLLs (steam_api.dll / steam_api64.dll).
// If found, it renames the original files to .original and creates a steam_appid.txt file.
func InjectSteamEmulator(gameDir string, appID string) error {
	return filepath.WalkDir(gameDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		name := strings.ToLower(d.Name())
		if name == "steam_api.dll" || name == "steam_api64.dll" {
			// Write steam_appid.txt in the DLL directory
			txtPath := filepath.Join(filepath.Dir(path), "steam_appid.txt")
			_ = os.WriteFile(txtPath, []byte(appID), 0644)

			// Backup original if not backed up yet
			originalPath := path + ".original"
			if _, err := os.Stat(originalPath); os.IsNotExist(err) {
				_ = os.Rename(path, originalPath)
				// Copy original back as primary if no custom emulator DLL is embedded
				_ = copyFile(originalPath, path)
			}
		} else if name == "hl.exe" || name == "steam.dll" {
			iniPath := filepath.Join(filepath.Dir(path), "SmartSteamEmu.ini")
			if _, err := os.Stat(iniPath); os.IsNotExist(err) {
				iniData := fmt.Sprintf(`[Launcher]
Target = %s
AppId = %s
PersonaName = RIFT_User
`, d.Name(), appID)
				_ = os.WriteFile(iniPath, []byte(iniData), 0644)
			}
		}

		return nil
	})
}

func copyFile(src, dst string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, in, 0755)
}
