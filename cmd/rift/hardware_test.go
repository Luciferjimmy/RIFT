package rift

import (
	"os"
	"testing"
)

func TestHardwareTelemetryProbes(t *testing.T) {
	// Test Hashed Device ID
	devID := GetHashedDeviceID()
	if len(devID) != 64 {
		t.Fatalf("Expected 64-character SHA256 hex string for device ID, got %q (len %d)", devID, len(devID))
	}
	t.Logf("Anonymous Device ID: %s", devID)

	// Test Thermal State
	therm := GetThermalState()
	validTherms := map[string]bool{"nominal": true, "fair": true, "serious": true, "critical": true}
	if !validTherms[therm] {
		t.Errorf("Unexpected thermal state: %q", therm)
	}
	t.Logf("Thermal State: %s", therm)

	// Test Battery vs AC
	onBatt := IsOnBattery()
	t.Logf("Is On Battery: %v", onBatt)

	// Test Rosetta 2
	isRosetta := IsRosettaTranslated()
	t.Logf("Is Rosetta Translated: %v", isRosetta)

	// Test Swap Used
	swapMB := GetSwapUsedMB()
	t.Logf("Swap Used: %.2f MB", swapMB)

	// Test Low Power Mode
	lowPower := IsLowPowerMode()
	t.Logf("Is Low Power Mode: %v", lowPower)

	// Test GPU Cores and Metal Support
	gpuCores := GetGPUCores()
	t.Logf("GPU Cores: %d", gpuCores)
	metalSupport := GetMetalSupport()
	t.Logf("Metal Support: %s", metalSupport)

	// Test GetMacHardwareSpecs
	specs := GetMacHardwareSpecs()
	if specs["Chip"] == "" {
		t.Errorf("Expected non-empty Chip")
	}
	if specs["DeviceID"] != devID {
		t.Errorf("Expected matching DeviceID")
	}
	if specs["GPUCores"] == nil || specs["GPUCores"].(int) <= 0 {
		if os.Getenv("CI") != "" {
			t.Logf("Warning: GPUCores <= 0 in headless CI runner: %v", specs["GPUCores"])
		} else {
			t.Errorf("Expected positive GPUCores on Apple Silicon, got %v", specs["GPUCores"])
		}
	}
	if specs["MetalSupport"] == "" {
		t.Errorf("Expected non-empty MetalSupport")
	}
	t.Logf("Full Hardware Specs: %+v", specs)
}
