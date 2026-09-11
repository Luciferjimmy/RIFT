package rift

import (
	"testing"
)

func TestIsVersionNewer(t *testing.T) {
	tests := []struct {
		latest   string
		current  string
		expected bool
	}{
		{"1.0.1", "1.0.0", true},
		{"1.1.0", "1.0.9", true},
		{"2.0.0", "1.9.9", true},
		{"1.0.0", "1.0.0", false},
		{"1.0.0", "1.0.1", false},
		{"0.9.9", "1.0.0", false},
		{"1.0.1-beta", "1.0.0", true},
		{"1.0.0-rc1", "1.0.0", false},
		{"1.0.0", "1.0.0-beta.2", true},
		{"1.0.0-beta.2", "1.0.0-beta.1", true},
		{"1.0.0-beta.1", "1.0.0-beta.2", false},
		{"1.0.0-beta.10", "1.0.0-beta.2", true},
		{"1.0.0-beta.2", "1.0.0-beta.10", false},
		{"1.0.0-rc.1", "1.0.0-beta.2", true},
		{"1.0.0-beta.2", "1.0.0-rc.1", false},
		{"1.0.0-beta.2", "1.0.0-beta.2", false},
		{"1.0.0-beta.2.1", "1.0.0-beta.2", true},
		{"1.0.0-beta.2", "1.0.0-beta.2.1", false},
	}

	for _, tc := range tests {
		got := isVersionNewer(tc.latest, tc.current)
		if got != tc.expected {
			t.Errorf("isVersionNewer(%q, %q) = %v; want %v", tc.latest, tc.current, got, tc.expected)
		}
	}
}

func TestGetAppVersion(t *testing.T) {
	app := &App{}
	if v := app.GetAppVersion(); v != AppVersion {
		t.Errorf("GetAppVersion() = %q; want %q", v, AppVersion)
	}
}

func TestGetAppBuildInfo(t *testing.T) {
	app := &App{}
	info := app.GetAppBuildInfo()
	if info["version"] != AppVersion {
		t.Errorf("GetAppBuildInfo()[version] = %q; want %q", info["version"], AppVersion)
	}
	if info["platform"] != "darwin" {
		t.Errorf("GetAppBuildInfo()[platform] = %q; want 'darwin'", info["platform"])
	}
}
