package optimizer

import (
	"testing"
)

func TestMajorVersion(t *testing.T) {
	tests := []struct {
		ver      string
		expected int
	}{
		{"14.5.1", 14},
		{"15.0", 15},
		{"16", 16},
		{"20.1", 20},
		{"malformed", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.ver, func(t *testing.T) {
			result := majorVersion(tt.ver)
			if result != tt.expected {
				t.Errorf("majorVersion(%q) = %d; want %d", tt.ver, result, tt.expected)
			}
		})
	}
}

func TestMetalGPUFamily(t *testing.T) {
	tests := []struct {
		name     string
		ver      string
		isAS     bool
		expected int
	}{
		{"Intel Mac always Metal 2", "15.0", false, 2},
		{"Sonoma Apple Silicon Metal 3", "14.5", true, 3},
		{"Sequoia Apple Silicon Metal 3", "15.1", true, 3},
		{"Future Apple Silicon Metal 4", "16.0", true, 4},
		{"Far Future Apple Silicon Metal 4", "20.0", true, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := metalGPUFamily(tt.ver, tt.isAS)
			if result != tt.expected {
				t.Errorf("metalGPUFamily(%q, %v) = %d; want %d", tt.ver, tt.isAS, result, tt.expected)
			}
		})
	}
}

func TestSystemHash(t *testing.T) {
	sp1 := &SystemProfile{
		MacOSVersion:   "14.5",
		AppleSilicon:   true,
		MetalGPUFamily: 3,
		RAMGB:          16,
		Rosetta2:       true,
		Engines: EngineStates{
			Wine: true,
			DXVK: false,
			GPTK: true,
		},
	}

	sp2 := &SystemProfile{
		MacOSVersion:   "14.5",
		AppleSilicon:   true,
		MetalGPUFamily: 3,
		RAMGB:          16,
		Rosetta2:       true,
		Engines: EngineStates{
			Wine: true,
			DXVK: true, // Changed
			GPTK: true,
		},
	}

	hash1 := sp1.SystemHash()
	hash2 := sp2.SystemHash()

	if hash1 == hash2 {
		t.Errorf("SystemHash should be different for different profiles: %s == %s", hash1, hash2)
	}
	if hash1 == "" {
		t.Errorf("SystemHash returned empty string")
	}
}
