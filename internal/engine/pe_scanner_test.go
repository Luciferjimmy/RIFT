package engine

import (
	"os"
	"testing"
)

func TestDetectGameGraphicsEngine(t *testing.T) {
	// Test on empty path
	_, err := DetectGameGraphicsEngine("")
	if err == nil {
		t.Errorf("Expected error on empty path, got nil")
	}

	// Test on non-existent path
	_, err = DetectGameGraphicsEngine("/non/existent/path.exe")
	if err == nil {
		t.Errorf("Expected error on non-existent path, got nil")
	}

	shippingExe := "/Users/abhinawsingh/.rift/games/0055e45ce7654c55aade646467349e83/game_files/MortalShell/Dungeonhaven/Binaries/Win64/Dungeonhaven-Win64-Shipping.exe"
	if _, err := os.Stat(shippingExe); err == nil {
		result, err := DetectGameGraphicsEngine(shippingExe)
		if err != nil {
			t.Fatalf("Failed to detect engine for Mortal Shell Shipping: %v", err)
		}
		t.Logf("Mortal Shell Shipping PE Analysis: Engine=%s, 32Bit=%v, API=%s, Total DLLs=%d, DLLs=%v",
			result.Engine, result.Is32Bit, result.DirectXAPI, len(result.ImportedDLLs), result.ImportedDLLs)
	}
}
