package types

import (
	"encoding/json"
	"testing"
)

func TestSystemProfileJSON(t *testing.T) {
	sp := SystemProfile{
		Chip:           "M2 Max",
		RAMGB:          32,
		MacOSVersion:   "14.4.1",
		GPUMetalFamily: "Apple8",
		HasGPTK:        true,
	}

	data, err := json.Marshal(sp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var sp2 SystemProfile
	if err := json.Unmarshal(data, &sp2); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if sp.Chip != sp2.Chip || sp.RAMGB != sp2.RAMGB || sp.MacOSVersion != sp2.MacOSVersion {
		t.Errorf("Mismatch after unmarshal. Expected %+v, got %+v", sp, sp2)
	}
}

func TestRuntimeConfigMerge(t *testing.T) {
	rc := RuntimeConfig{
		Engine:     "d3dmetal",
		LaunchArgs: []string{"-vulkan"},
		EnvVars: map[string]string{
			"WINEDLLOVERRIDES": "dxgi=n,b",
		},
	}

	if rc.Engine != "d3dmetal" {
		t.Errorf("Expected engine d3dmetal, got %s", rc.Engine)
	}
	if rc.EnvVars["WINEDLLOVERRIDES"] != "dxgi=n,b" {
		t.Errorf("Expected WINEDLLOVERRIDES to be dxgi=n,b")
	}
}

func TestCloudGameConfigDefaults(t *testing.T) {
	cgc := CloudGameConfig{
		Engine:     "dxvk",
		Confidence: 0.9,
	}
	if cgc.Confidence != 0.9 {
		t.Errorf("Expected confidence 0.9, got %f", cgc.Confidence)
	}
}
