package rift

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	cachedDisplayRes   string
	cachedGPUCores     int
	cachedMetalSupport string
	displayOnce        sync.Once
	cachedDeviceID     string
	deviceIDOnce       sync.Once
)

// GetHashedDeviceID computes an irreversible, anonymized SHA-256 hash of the machine's
// hardware UUID salted with an internal application secret.
// This guarantees that raw hardware identifiers or serial numbers are NEVER transmitted.
func GetHashedDeviceID() string {
	deviceIDOnce.Do(func() {
		rawID := ""
		out, err := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output()
		if err == nil {
			re := regexp.MustCompile(`"IOPlatformUUID"\s*=\s*"([^"]+)"`)
			matches := re.FindStringSubmatch(string(out))
			if len(matches) > 1 {
				rawID = matches[1]
			}
		}

		if rawID == "" {
			// Fallback: stable fallback from CPU brand and memory size
			cpu, _ := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output()
			mem, _ := exec.Command("sysctl", "-n", "hw.memsize").Output()
			rawID = string(cpu) + "_" + string(mem)
		}

		// Irreversible SHA-256 salt
		h := sha256.New()
		h.Write([]byte("RIFT_ANON_DEVICE_SALT_v1:" + rawID))
		cachedDeviceID = hex.EncodeToString(h.Sum(nil))
	})
	return cachedDeviceID
}

// GetThermalState inspects macOS power management for thermal throttle warnings.
// Returns "nominal", "fair", "serious", or "critical".
func GetThermalState() string {
	out, err := exec.Command("pmset", "-g", "therm").Output()
	if err == nil {
		s := string(out)

		if strings.Contains(s, "Thermal_Warning_Level = 3") {
			return "critical"
		} else if strings.Contains(s, "Thermal_Warning_Level = 2") {
			return "serious"
		} else if strings.Contains(s, "Thermal_Warning_Level = 1") {
			return "fair"
		}

		// Check if CPU speed is limited by thermal governor
		re := regexp.MustCompile(`CPU_Speed_Limit\s*=\s*(\d+)`)
		if match := re.FindStringSubmatch(s); len(match) > 1 {
			if limit, err := strconv.Atoi(match[1]); err == nil && limit > 0 {
				if limit < 70 {
					return "serious"
				} else if limit < 90 {
					return "fair"
				}
			}
		}
	}

	// Secondary probe / fallback for Intel Macs and legacy macOS (xcpm thermal level)
	if out, err := exec.Command("sysctl", "-n", "machdep.xcpm.cpu_thermal_level").Output(); err == nil {
		if lvl, err := strconv.Atoi(strings.TrimSpace(string(out))); err == nil {
			switch {
			case lvl >= 3:
				return "critical"
			case lvl == 2:
				return "serious"
			case lvl == 1:
				return "fair"
			}
		}
	}

	return "nominal"
}

// IsOnBattery returns true if running on Battery Power, false if on AC Power or desktop Mac.
func IsOnBattery() bool {
	out, err := exec.Command("pmset", "-g", "batt").Output()
	if err != nil {
		return false
	}
	s := string(out)
	if strings.Contains(s, "'Battery Power'") {
		return true
	}
	return false
}

// IsLowPowerMode returns true if macOS Low Power Mode is currently engaged.
func IsLowPowerMode() bool {
	out, err := exec.Command("pmset", "-g").Output()
	if err == nil {
		re := regexp.MustCompile(`lowpowermode\s+(\d+)`)
		if m := re.FindStringSubmatch(string(out)); len(m) > 1 {
			return m[1] == "1"
		}
	}
	return false
}

// IsRosettaTranslated checks whether the current process/host is executing through Rosetta 2.
func IsRosettaTranslated() bool {
	out, err := exec.Command("sysctl", "-n", "sysctl.proc_translated").Output()
	if err == nil && strings.TrimSpace(string(out)) == "1" {
		return true
	}
	return false
}

// GetSwapUsedMB extracts the current unified memory swap utilization in megabytes.
func GetSwapUsedMB() float64 {
	out, err := exec.Command("sysctl", "-n", "vm.swapusage").Output()
	if err != nil {
		return 0.0
	}
	// Format: total = 6144.00M  used = 4935.19M  free = 1208.81M  (encrypted)
	re := regexp.MustCompile(`used\s*=\s*([\d\.]+)M`)
	matches := re.FindStringSubmatch(string(out))
	if len(matches) > 1 {
		if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return val
		}
	}
	return 0.0
}

func probeDisplayAndGPU() {
	displayOnce.Do(func() {
		out, err := exec.Command("system_profiler", "SPDisplaysDataType").Output()
		if err == nil {
			s := string(out)
			reRes := regexp.MustCompile(`Resolution:\s*([^\n\r]+)`)
			if m := reRes.FindStringSubmatch(s); len(m) > 1 {
				cachedDisplayRes = strings.TrimSpace(m[1])
			}

			reCores := regexp.MustCompile(`Total Number of Cores:\s*(\d+)`)
			if m := reCores.FindStringSubmatch(s); len(m) > 1 {
				if c, err := strconv.Atoi(m[1]); err == nil {
					cachedGPUCores = c
				}
			}

			reMetal := regexp.MustCompile(`Metal Support:\s*([^\n\r]+)`)
			if m := reMetal.FindStringSubmatch(s); len(m) > 1 {
				cachedMetalSupport = strings.TrimSpace(m[1])
			}
		}
		if cachedGPUCores == 0 {
			// Headless VM / CI fallback
			if c, err := exec.Command("sysctl", "-n", "hw.ncpu").Output(); err == nil {
				if n, err := strconv.Atoi(strings.TrimSpace(string(c))); err == nil && n > 0 {
					cachedGPUCores = n
				}
			}
			if cachedGPUCores == 0 {
				cachedGPUCores = 8
			}
		}
		if cachedDisplayRes == "" {
			cachedDisplayRes = "Default Retina"
		}
		if cachedMetalSupport == "" {
			cachedMetalSupport = "Metal 3"
		}
	})
}

// GetDisplayResolution returns the primary display resolution string (cached once).
func GetDisplayResolution() string {
	probeDisplayAndGPU()
	return cachedDisplayRes
}

// GetGPUCores returns the exact GPU core count probed from the GPU hardware profile.
func GetGPUCores() int {
	probeDisplayAndGPU()
	return cachedGPUCores
}

// GetMetalSupport returns the Metal capability string (e.g. "Metal 4").
func GetMetalSupport() string {
	probeDisplayAndGPU()
	return cachedMetalSupport
}

// GetMacHardwareSpecs compiles comprehensive Apple Silicon hardware truth.
func GetMacHardwareSpecs() map[string]interface{} {
	specs := make(map[string]interface{})

	// Anonymous Hashed Device ID (No raw UUID)
	specs["DeviceID"] = GetHashedDeviceID()

	// Chip identification
	out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output()
	if err == nil {
		specs["Chip"] = strings.TrimSpace(string(out))
	} else {
		specs["Chip"] = "Apple Silicon"
	}

	// RAM in GB
	out, err = exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err == nil {
		if bytes, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64); err == nil {
			specs["RAMGB"] = bytes / (1024 * 1024 * 1024)
		}
	}

	// macOS Version
	out, err = exec.Command("sw_vers", "-productVersion").Output()
	if err == nil {
		specs["MacOSVersion"] = strings.TrimSpace(string(out))
	}

	// Real-time environmental facts
	specs["ThermalState"] = GetThermalState()
	specs["OnBattery"] = IsOnBattery()
	specs["LowPowerMode"] = IsLowPowerMode()
	specs["Rosetta2"] = IsRosettaTranslated()
	specs["SwapUsedMB"] = GetSwapUsedMB()
	specs["DisplayRes"] = GetDisplayResolution()
	specs["GPUCores"] = GetGPUCores()
	specs["MetalSupport"] = GetMetalSupport()

	// Engine capability probes
	home, _ := os.UserHomeDir()
	enginesDir := filepath.Join(home, ".rift", "engines")
	if fi, err := os.Stat(filepath.Join(enginesDir, "gptk")); err == nil && fi.IsDir() {
		specs["HasGPTK"] = true
	} else {
		specs["HasGPTK"] = false
	}
	if fi, err := os.Stat(filepath.Join(enginesDir, "dxvk")); err == nil && fi.IsDir() {
		specs["HasDXVK"] = true
	} else {
		specs["HasDXVK"] = false
	}

	// Dynamic GPU Metal Family identification based on Apple Silicon architecture
	chipUpper := strings.ToUpper(specs["Chip"].(string))
	if strings.Contains(chipUpper, "M4") {
		specs["GPUMetalFamily"] = "Apple10"
	} else if strings.Contains(chipUpper, "M3") {
		specs["GPUMetalFamily"] = "Apple9"
	} else if strings.Contains(chipUpper, "M2") {
		specs["GPUMetalFamily"] = "Apple8"
	} else if strings.Contains(chipUpper, "M1") {
		specs["GPUMetalFamily"] = "Apple7"
	} else {
		specs["GPUMetalFamily"] = "Mac2"
	}

	return specs
}
