package engine

import (
	"debug/pe"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PEDetectionResult contains the analyzed properties of a Windows executable.
type PEDetectionResult struct {
	Engine        string   // "d3dmetal", "dxvk", or "wine"
	Is32Bit       bool     // true if 32-bit (x86), false if 64-bit (x86_64)
	DirectXAPI    string   // "DirectX 12", "DirectX 11", "DirectX 10", "DirectX 9", "OpenGL", "Vulkan", or "Unknown"
	ImportedDLLs  []string // List of all imported DLLs found
	RequiresWoW64 bool     // true if 32-bit WoW64 execution is needed
}

// DetectGameGraphicsEngine analyzes a Windows PE (.exe) binary to determine the optimal graphics engine.
// It checks machine architecture and parses imported DLLs to distinguish DX9/10/Legacy from DX11/12.
func DetectGameGraphicsEngine(exePath string) (*PEDetectionResult, error) {
	if exePath == "" {
		return nil, fmt.Errorf("empty executable path")
	}

	f, err := os.Open(exePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open executable %s: %w", exePath, err)
	}
	defer f.Close()

	peFile, err := pe.NewFile(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PE headers for %s: %w", exePath, err)
	}
	defer peFile.Close()

	result := &PEDetectionResult{
		Engine:     "d3dmetal", // Default modern fallback
		DirectXAPI: "Unknown",
	}

	// 1. Architecture Check
	switch peFile.Machine {
	case pe.IMAGE_FILE_MACHINE_I386:
		result.Is32Bit = true
		result.RequiresWoW64 = true
	case pe.IMAGE_FILE_MACHINE_AMD64:
		result.Is32Bit = false
		result.RequiresWoW64 = false
	default:
		// Default to 64-bit assumption if unknown
		result.Is32Bit = false
	}

	// 2. Read Imported Libraries
	importedLibs, err := peFile.ImportedLibraries()
	if err != nil {
		// Even if imported libraries reading fails, return architecture result
		if result.Is32Bit {
			result.Engine = "dxvk"
		} else {
			result.Engine = "d3dmetal"
		}
		return result, nil
	}

	result.ImportedDLLs = importedLibs

	// 3. Score graphics APIs based on imports
	hasDX12 := false
	hasDX11 := false
	hasDX10 := false
	hasDX9 := false
	hasVulkan := false
	hasOpenGL := false

	if len(importedLibs) > 0 {
		for _, lib := range importedLibs {
			lower := strings.ToLower(filepath.Base(lib))
			switch {
			case lower == "d3d12.dll" || strings.HasPrefix(lower, "d3d12"):
				hasDX12 = true
			case lower == "d3d11.dll" || strings.HasPrefix(lower, "d3d11"):
				hasDX11 = true
			case lower == "d3d10.dll" || lower == "d3d10core.dll" || lower == "d3d10_1.dll" || strings.HasPrefix(lower, "d3d10"):
				hasDX10 = true
			case lower == "d3d9.dll" || lower == "d3d8.dll" || lower == "ddraw.dll" || strings.HasPrefix(lower, "d3dx9"):
				hasDX9 = true
			case lower == "vulkan-1.dll" || strings.HasPrefix(lower, "vulkan"):
				hasVulkan = true
			case lower == "opengl32.dll":
				hasOpenGL = true
			}
		}
	} else {
		// If static imports were empty (common with delay-load or custom linkers), scan binary data
		buf := make([]byte, 10*1024*1024) // Read up to first 10MB
		f.Seek(0, 0)
		n, _ := f.Read(buf)
		content := strings.ToLower(string(buf[:n]))

		if strings.Contains(content, "d3d12.dll") {
			hasDX12 = true
			result.ImportedDLLs = append(result.ImportedDLLs, "d3d12.dll")
		}
		if strings.Contains(content, "d3d11.dll") {
			hasDX11 = true
			result.ImportedDLLs = append(result.ImportedDLLs, "d3d11.dll")
		}
		if strings.Contains(content, "d3d10.dll") || strings.Contains(content, "d3d10core.dll") {
			hasDX10 = true
			result.ImportedDLLs = append(result.ImportedDLLs, "d3d10.dll")
		}
		if strings.Contains(content, "d3d9.dll") || strings.Contains(content, "d3dx9") {
			hasDX9 = true
			result.ImportedDLLs = append(result.ImportedDLLs, "d3d9.dll")
		}
		if strings.Contains(content, "vulkan-1.dll") {
			hasVulkan = true
			result.ImportedDLLs = append(result.ImportedDLLs, "vulkan-1.dll")
		}
		if strings.Contains(content, "opengl32.dll") {
			hasOpenGL = true
			result.ImportedDLLs = append(result.ImportedDLLs, "opengl32.dll")
		}
	}

	// Priority scoring: DX12 > DX11 > DX10 > DX9 > Vulkan > OpenGL
	if hasDX12 {
		result.Engine = "d3dmetal"
		result.DirectXAPI = "DirectX 12"
	} else if hasDX11 {
		result.Engine = "d3dmetal"
		result.DirectXAPI = "DirectX 11"
	} else if hasDX10 {
		result.Engine = "dxvk"
		result.DirectXAPI = "DirectX 10"
	} else if hasDX9 {
		result.Engine = "dxvk"
		result.DirectXAPI = "DirectX 9"
	} else if hasVulkan {
		result.Engine = "dxvk"
		result.DirectXAPI = "Vulkan"
	} else if hasOpenGL {
		result.Engine = "wine"
		result.DirectXAPI = "OpenGL"
	} else {
		// Inconclusive fallback: 32-bit legacy titles use DXVK, 64-bit titles use D3DMetal
		if result.Is32Bit {
			result.Engine = "dxvk"
			result.DirectXAPI = "DirectX 9/Legacy (32-bit Auto)"
		} else {
			result.Engine = "d3dmetal"
			result.DirectXAPI = "DirectX 11/12 (64-bit Auto)"
		}
	}

	return result, nil
}

// IsLargeAddressAware checks if a PE binary has the IMAGE_FILE_LARGE_ADDRESS_AWARE flag (0x0020) set.
func IsLargeAddressAware(exePath string) bool {
	if exePath == "" {
		return false
	}
	f, err := pe.Open(exePath)
	if err != nil {
		return false
	}
	defer f.Close()
	return (f.Characteristics & 0x0020) != 0
}
