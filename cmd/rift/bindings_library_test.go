package rift

import (
	"testing"
)

func TestIsAllDigits(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"730", true},
		{"11020", true},
		{"1091500", true},
		{"", false},
		{"steam-730", false},
		{"41869934302e4b8cafac2d3c0e7c293d", false},
		{"Sugar", false},
		{"12a34", false},
	}

	for _, tc := range tests {
		got := isAllDigits(tc.input)
		if got != tc.expected {
			t.Errorf("isAllDigits(%q) = %v; want %v", tc.input, got, tc.expected)
		}
	}
}

func TestIs32Hex(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"41869934302e4b8cafac2d3c0e7c293d", true},
		{"0055e45ce7654c55aade646467349e83", true},
		{"3e02273b543f4ff0a1c24d3b534a9ac3", true},
		{"41869934302E4B8CAFAC2D3C0E7C293D", true},
		{"730", false},
		{"", false},
		{"epic-41869934302e4b8cafac2d3c0e7c293d", false}, // has prefix
		{"Sugar", false},
		{"41869934302e4b8cafac2d3c0e7c293z", false},      // 'z' is not hex
		{"41869934302e4b8cafac2d3c0e7c293", false},       // 31 chars
	}

	for _, tc := range tests {
		got := is32Hex(tc.input)
		if got != tc.expected {
			t.Errorf("is32Hex(%q) = %v; want %v", tc.input, got, tc.expected)
		}
	}
}

func TestDetectCapsulePlatformAndAppID(t *testing.T) {
	tests := []struct {
		gameID         string
		expectedPlat   string
		expectedAppID  string
	}{
		{"steam-730", "Steam", "730"},
		{"steam-11020", "Steam", "11020"},
		{"730", "Steam", "730"},
		{"1091500", "Steam", "1091500"},
		{"epic-Sugar", "Epic", "Sugar"},
		{"41869934302e4b8cafac2d3c0e7c293d", "Epic", "41869934302e4b8cafac2d3c0e7c293d"},
		{"0055e45ce7654c55aade646467349e83", "Epic", "0055e45ce7654c55aade646467349e83"},
		{"Sugar", "Epic", "Sugar"},
	}

	for _, tc := range tests {
		plat, cleanID := detectCapsulePlatformAndAppID(tc.gameID)
		if plat != tc.expectedPlat || cleanID != tc.expectedAppID {
			t.Errorf("detectCapsulePlatformAndAppID(%q) = (%q, %q); want (%q, %q)",
				tc.gameID, plat, cleanID, tc.expectedPlat, tc.expectedAppID)
		}
	}
}
