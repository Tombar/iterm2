package iterm2

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Tombar/iterm2/client"
)

// Sentinel errors for iTerm2 prerequisite failures.
// These can be checked with errors.Is() for programmatic error handling.
var (
	// ErrITerm2NotRunning indicates the iTerm2 application is not running.
	ErrITerm2NotRunning = errors.New("iTerm2 is not running")

	// ErrPythonAPIDisabled indicates the Python API is not enabled in iTerm2 preferences.
	// To fix: Open iTerm2 → Preferences → General → Magic → Enable "Python API"
	ErrPythonAPIDisabled = errors.New("iTerm2 Python API is not enabled")

	// ErrPermissionDenied indicates the user denied permission for this application.
	// This typically occurs on first run when iTerm2 prompts for authorization.
	ErrPermissionDenied = errors.New("iTerm2 permission denied for this application")
)

// CheckPrerequisites verifies that iTerm2 is running and the Python API is enabled.
// It does NOT check permissions (use RequestPermission for that).
//
// This function performs fast, non-invasive checks:
//  1. Verifies iTerm2 process is running
//  2. Verifies Python API socket exists
//
// Returns nil if all checks pass, or a typed error that can be checked with errors.Is():
//   - ErrITerm2NotRunning: iTerm2 is not running
//   - ErrPythonAPIDisabled: Python API is not enabled
//
// Example usage:
//
//	if err := iterm2.CheckPrerequisites("myapp"); err != nil {
//	    if errors.Is(err, iterm2.ErrITerm2NotRunning) {
//	        fmt.Println("Please launch iTerm2")
//	    } else if errors.Is(err, iterm2.ErrPythonAPIDisabled) {
//	        fmt.Println("Please enable Python API in iTerm2 preferences")
//	    }
//	}
func CheckPrerequisites(appName string) error {
	// Check if iTerm2 is running
	if !isITerm2Running() {
		return fmt.Errorf("%w: iTerm2 process not found. Launch with: open -a iTerm", ErrITerm2NotRunning)
	}

	// Check if Python API is enabled (socket exists)
	if !isPythonAPIEnabled() {
		// Re-check iTerm2 is still running to avoid TOCTOU race
		if !isITerm2Running() {
			return fmt.Errorf("%w: iTerm2 process not found. Launch with: open -a iTerm", ErrITerm2NotRunning)
		}
		socketPath, err := getSocketPath()
		if err != nil {
			return fmt.Errorf("%w: could not determine socket path: %v", ErrPythonAPIDisabled, err)
		}
		return fmt.Errorf("%w: socket not found at %s. Enable in iTerm2 → Preferences → General → Magic → Python API",
			ErrPythonAPIDisabled, socketPath)
	}

	return nil
}

// isPermissionError checks if an error message indicates a permission/authorization issue.
func isPermissionError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "permission") ||
		strings.Contains(errMsg, "denied") ||
		strings.Contains(errMsg, "declined") ||
		strings.Contains(errMsg, "not authorized") ||
		strings.Contains(errMsg, "authorization")
}

// isPythonAPIError checks if an error message indicates the Python API is not enabled.
func isPythonAPIError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "python api is not enabled") ||
		strings.Contains(errMsg, "python api")
}

// RequestPermission attempts to request permission from iTerm2 for the specified application.
// This performs a minimal connection attempt that will show the authorization dialog if the
// user has not previously granted permission.
//
// On subsequent runs with the same appName, this function will return immediately without
// showing a dialog (iTerm2 remembers approved applications).
//
// Returns nil if permission is granted, or a typed error:
//   - ErrITerm2NotRunning: iTerm2 is not running
//   - ErrPythonAPIDisabled: Python API is not enabled
//   - ErrPermissionDenied: User declined the permission prompt
//
// Example usage:
//
//	if err := iterm2.RequestPermission("myapp"); err != nil {
//	    if errors.Is(err, iterm2.ErrPermissionDenied) {
//	        fmt.Println("Access denied. Please allow 'myapp' in iTerm2 settings")
//	    }
//	}
func RequestPermission(appName string) error {
	// First check prerequisites (iTerm2 running + API enabled)
	if err := CheckPrerequisites(appName); err != nil {
		return err
	}

	// Attempt minimal connection to test permission
	// This will trigger the authorization dialog if not previously granted
	c, err := client.New(appName)
	if err != nil {
		// Parse error to determine the issue type
		if isPythonAPIError(err) {
			return fmt.Errorf("%w: %v", ErrPythonAPIDisabled, err)
		}
		if isPermissionError(err) {
			return fmt.Errorf("%w: %v", ErrPermissionDenied, err)
		}
		// Other error (return as-is)
		return err
	}

	// Permission granted - close immediately
	_ = c.Close()
	return nil
}

// isITerm2Running checks if the iTerm2 process is currently running.
// Returns true if iTerm2.app is found in the process list.
func isITerm2Running() bool {
	// Use pgrep to check for iTerm.app process
	// -f searches full command line (needed because process runs as /Applications/iTerm.app/Contents/MacOS/iTerm2)
	cmd := exec.Command("pgrep", "-f", "iTerm.app")
	err := cmd.Run()
	return err == nil // pgrep returns 0 if process found
}

// isPythonAPIEnabled checks if the Python API is enabled by verifying
// the Unix socket file exists.
func isPythonAPIEnabled() bool {
	socketPath, err := getSocketPath()
	if err != nil {
		return false
	}

	// Check if socket file exists
	_, err = os.Stat(socketPath)
	return err == nil
}

// getSocketPath returns the Unix socket path for iTerm2's API.
// Returns error if home directory cannot be determined.
func getSocketPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(homeDir, "Library", "Application Support", "iTerm2", "private", "socket"), nil
}
