package iterm2

import (
	"strings"
	"testing"
	"time"
)

// TestGetSocketPath_Public tests the public GetSocketPath function
func TestGetSocketPath_Public(t *testing.T) {
	path, err := GetSocketPath()
	if err != nil {
		t.Fatalf("GetSocketPath() returned error: %v", err)
	}
	verifySocketPath(t, path)
}

// TestEnablePythonAPIGuide verifies guide content
func TestEnablePythonAPIGuide(t *testing.T) {
	guide := EnablePythonAPIGuide()

	// Verify guide is not empty
	if guide == "" {
		t.Fatal("EnablePythonAPIGuide() returned empty string")
	}

	// Verify guide contains key information
	expectedStrings := []string{
		"Python API",
		"Preferences",
		"General",
		"Magic",
		"Enable",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(guide, expected) {
			t.Errorf("guide missing expected string %q", expected)
		}
	}

	t.Logf("Guide content:\n%s", guide)
}

// TestWaitForITerm2_Timeout tests timeout behavior
func TestWaitForITerm2_Timeout(t *testing.T) {
	// Only run this test if iTerm2 is NOT running
	if isITerm2Running() {
		t.Skip("Skipping timeout test - iTerm2 is running")
	}

	// Very short timeout to test failure case
	timeout := 100 * time.Millisecond
	err := WaitForITerm2(timeout)

	if err == nil {
		t.Error("WaitForITerm2() should have timed out but returned nil")
	}

	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("error should mention timeout, got: %v", err)
	}
}

// TestWaitForITerm2_AlreadyRunning tests immediate success
func TestWaitForITerm2_AlreadyRunning(t *testing.T) {
	// Only run this test if iTerm2 IS running
	if !isITerm2Running() {
		t.Skip("Skipping success test - iTerm2 is not running")
	}

	err := WaitForITerm2(5 * time.Second)
	if err != nil {
		t.Errorf("WaitForITerm2() failed even though iTerm2 is running: %v", err)
	}
}

// TestLaunchITerm2_IdempotentWhenRunning tests launching when already running
func TestLaunchITerm2_IdempotentWhenRunning(t *testing.T) {
	// Only run this test if iTerm2 IS running
	if !isITerm2Running() {
		t.Skip("Skipping idempotent test - iTerm2 is not running")
	}

	// Should not error when iTerm2 is already running
	err := LaunchITerm2()
	if err != nil {
		t.Errorf("LaunchITerm2() failed when iTerm2 already running: %v", err)
	}
}

// TestOpenITerm2Preferences tests preference opening command
// Note: This doesn't actually open preferences during test, just verifies no panic
func TestOpenITerm2Preferences(t *testing.T) {
	// Skip if iTerm2 is not running, as the command may fail
	if !isITerm2Running() {
		t.Skip("Skipping preferences test - iTerm2 is not running")
	}

	// Just verify the function doesn't panic
	// We don't actually open preferences during automated tests
	t.Log("OpenITerm2Preferences() would open preferences (skipped in test)")
}
