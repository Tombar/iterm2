package iterm2

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// verifySocketPath is a test helper that verifies a socket path has the correct format.
func verifySocketPath(t *testing.T, path string) {
	t.Helper()

	// Verify path contains expected components
	expectedComponents := []string{
		"Library",
		"Application Support",
		"iTerm2",
		"private",
		"socket",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(path, component) {
			t.Errorf("socket path %q missing expected component %q", path, component)
		}
	}

	// Verify path is absolute
	if !strings.HasPrefix(path, "/") {
		t.Errorf("socket path %q is not absolute", path)
	}
}

// TestSentinelErrors verifies that sentinel errors can be detected with errors.Is()
func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name     string
		sentinel error
		wrapped  error
	}{
		{
			name:     "iTerm2 not running",
			sentinel: ErrITerm2NotRunning,
			wrapped:  fmt.Errorf("context: %w", ErrITerm2NotRunning),
		},
		{
			name:     "Python API disabled",
			sentinel: ErrPythonAPIDisabled,
			wrapped:  fmt.Errorf("context: %w", ErrPythonAPIDisabled),
		},
		{
			name:     "permission denied",
			sentinel: ErrPermissionDenied,
			wrapped:  fmt.Errorf("context: %w", ErrPermissionDenied),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test direct error
			if !errors.Is(tt.sentinel, tt.sentinel) {
				t.Errorf("errors.Is() failed for direct error")
			}

			// Test wrapped error
			if !errors.Is(tt.wrapped, tt.sentinel) {
				t.Errorf("errors.Is() failed for wrapped error")
			}

			// Test that other sentinels don't match
			for _, other := range []error{ErrITerm2NotRunning, ErrPythonAPIDisabled, ErrPermissionDenied} {
				if other != tt.sentinel && errors.Is(tt.sentinel, other) {
					t.Errorf("errors.Is() incorrectly matched different sentinel")
				}
			}
		})
	}
}

// TestGetSocketPath verifies socket path construction
func TestGetSocketPath(t *testing.T) {
	path, err := getSocketPath()
	if err != nil {
		t.Fatalf("getSocketPath() returned error: %v", err)
	}
	verifySocketPath(t, path)
}

// TestIsITerm2Running tests iTerm2 process detection
// Note: This test assumes iTerm2 may or may not be running
func TestIsITerm2Running(t *testing.T) {
	running := isITerm2Running()
	t.Logf("iTerm2 running: %v", running)

	// We can't assert a specific value since iTerm2 may or may not be running
	// Just verify the function returns without panic
	// The return value is boolean, so any value is valid
}

// TestIsPythonAPIEnabled tests Python API detection via socket
// Note: This test assumes Python API may or may not be enabled
func TestIsPythonAPIEnabled(t *testing.T) {
	enabled := isPythonAPIEnabled()
	t.Logf("Python API enabled (socket exists): %v", enabled)

	// We can't assert a specific value since API may or may not be enabled
	// Just verify the function returns without panic
}

// TestCheckPrerequisites_ErrorMessages verifies error messages include context
func TestCheckPrerequisites_ErrorMessages(t *testing.T) {
	// Note: This test will behave differently depending on system state
	// We're primarily testing that errors include helpful context

	err := CheckPrerequisites("test-app")
	if err != nil {
		t.Logf("CheckPrerequisites returned error (expected if iTerm2 not running or API disabled): %v", err)

		// Verify error is typed
		if !errors.Is(err, ErrITerm2NotRunning) && !errors.Is(err, ErrPythonAPIDisabled) {
			t.Errorf("CheckPrerequisites returned non-typed error: %v", err)
		}

		// Verify error message includes helpful context
		errMsg := err.Error()
		if !strings.Contains(errMsg, "iTerm2") {
			t.Errorf("error message missing context: %v", err)
		}
	} else {
		t.Log("CheckPrerequisites passed (iTerm2 running with Python API enabled)")
	}
}

// TestEnhanceConnectionError tests error type detection and wrapping
func TestEnhanceConnectionError(t *testing.T) {
	// Determine expected error based on whether iTerm2 is actually running
	var socketErrorExpected error
	if isITerm2Running() {
		// If iTerm2 is running but socket errors occur, it means Python API is disabled
		socketErrorExpected = ErrPythonAPIDisabled
	} else {
		// If iTerm2 is not running, socket errors indicate that
		socketErrorExpected = ErrITerm2NotRunning
	}

	tests := []struct {
		name          string
		inputError    string
		expectedError error
	}{
		{
			name:          "socket not found",
			inputError:    "dial unix /path/to/socket: no such file or directory",
			expectedError: socketErrorExpected,
		},
		{
			name:          "connection refused",
			inputError:    "dial unix /path/to/socket: connection refused",
			expectedError: socketErrorExpected,
		},
		{
			name:          "Python API disabled",
			inputError:    "AppleScript/tell: The Python API is not enabled",
			expectedError: ErrPythonAPIDisabled,
		},
		{
			name:          "permission denied",
			inputError:    "AppleScript/tell: permission denied",
			expectedError: ErrPermissionDenied,
		},
		{
			name:          "user declined",
			inputError:    "AppleScript/tell: user declined authorization",
			expectedError: ErrPermissionDenied,
		},
		{
			name:          "not authorized",
			inputError:    "not authorized to control iTerm2",
			expectedError: ErrPermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputErr := errors.New(tt.inputError)
			enhanced := enhanceConnectionError(inputErr, "test-app")

			if !errors.Is(enhanced, tt.expectedError) {
				t.Errorf("enhanceConnectionError() = %v, want error type %v", enhanced, tt.expectedError)
			}

			// Verify original error is included in message
			if !strings.Contains(enhanced.Error(), tt.inputError) {
				t.Errorf("enhanced error doesn't include original error message")
			}
		})
	}
}

// TestEnhanceConnectionError_UnknownError verifies unknown errors pass through
func TestEnhanceConnectionError_UnknownError(t *testing.T) {
	unknownErr := errors.New("some random error")
	enhanced := enhanceConnectionError(unknownErr, "test-app")

	// Should return the original error unchanged
	if enhanced != unknownErr {
		t.Errorf("enhanceConnectionError() modified unknown error")
	}

	// Should not be wrapped with any sentinel
	if errors.Is(enhanced, ErrITerm2NotRunning) ||
		errors.Is(enhanced, ErrPythonAPIDisabled) ||
		errors.Is(enhanced, ErrPermissionDenied) {
		t.Errorf("enhanceConnectionError() incorrectly wrapped unknown error")
	}
}
