package iterm2

import (
	"fmt"
	"os/exec"
	"time"
)

// OpenITerm2Preferences opens iTerm2's preferences window to the General tab
// where the Python API can be enabled.
//
// This function uses macOS's "open" command to launch System Preferences.
// Note: The user will need to manually navigate to:
// General → Magic → Enable "Python API"
//
// Returns an error if the command fails to execute.
func OpenITerm2Preferences() error {
	// Open iTerm2 preferences
	// This will bring iTerm2 to front and open its preferences window
	cmd := exec.Command("open", "-a", "iTerm", "--args", "prefs")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open iTerm2 preferences: %w", err)
	}
	return nil
}

// WaitForITerm2 polls until iTerm2 is running or the timeout expires.
// This is useful for automation scripts that need to ensure iTerm2 is ready.
//
// The function checks every 500ms whether iTerm2 has started.
//
// Returns nil if iTerm2 starts within the timeout period.
// Returns an error if the timeout expires before iTerm2 starts.
//
// Example usage:
//
//	if err := iterm2.LaunchITerm2(); err != nil {
//	    return err
//	}
//	if err := iterm2.WaitForITerm2(30 * time.Second); err != nil {
//	    return fmt.Errorf("iTerm2 did not start: %w", err)
//	}
func WaitForITerm2(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	// Check immediately first
	if isITerm2Running() {
		return nil
	}

	for {
		select {
		case <-ticker.C:
			// Check deadline before potentially slow isITerm2Running() call
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for iTerm2 to start after %v", timeout)
			}
			if isITerm2Running() {
				return nil
			}
		}
	}
}

// LaunchITerm2 launches the iTerm2 application if it's not already running.
// This uses macOS's "open" command to start iTerm2.
//
// Returns nil if iTerm2 is successfully launched or is already running.
// Returns an error if the launch command fails.
//
// Note: This function returns immediately after executing the launch command.
// Use WaitForITerm2() if you need to wait for iTerm2 to be fully started.
//
// Example usage:
//
//	if err := iterm2.LaunchITerm2(); err != nil {
//	    return fmt.Errorf("failed to launch iTerm2: %w", err)
//	}
//	// Wait for iTerm2 to be ready
//	if err := iterm2.WaitForITerm2(30 * time.Second); err != nil {
//	    return fmt.Errorf("iTerm2 did not start: %w", err)
//	}
func LaunchITerm2() error {
	// Check if already running to avoid unnecessary launch
	if isITerm2Running() {
		return nil
	}

	cmd := exec.Command("open", "-a", "iTerm")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to launch iTerm2: %w. Is iTerm2 installed in /Applications?", err)
	}

	return nil
}

// EnablePythonAPIGuide returns a formatted guide for enabling the Python API in iTerm2.
// This can be printed to help users fix the ErrPythonAPIDisabled error.
//
// Returns a multi-line string with step-by-step instructions.
//
// Example usage:
//
//	if errors.Is(err, iterm2.ErrPythonAPIDisabled) {
//	    fmt.Println(iterm2.EnablePythonAPIGuide())
//	    iterm2.OpenITerm2Preferences()
//	}
func EnablePythonAPIGuide() string {
	return `
┌────────────────────────────────────────────────────────────────┐
│ How to Enable iTerm2 Python API                                │
└────────────────────────────────────────────────────────────────┘

The Python API must be enabled for automation to work (despite the
name, this setting is required for all automation languages including Go).

Steps to enable:

  1. Open iTerm2 (if not already running)

  2. Go to: iTerm2 → Preferences (⌘,)

  3. Navigate to: General tab

  4. Find the "Magic" section

  5. Check: ☑ "Enable Python API"

  6. Restart your application

You can also open preferences automatically by calling:
  iterm2.OpenITerm2Preferences()

More information: https://iterm2.com/python-api-auth.html
`
}

// GetSocketPath returns the Unix socket path used by iTerm2's automation API.
// This is useful for debugging connection issues or verifying the socket exists.
//
// Returns the socket path and an error if the home directory cannot be determined.
//
// The socket only exists when iTerm2 is running with Python API enabled.
//
// Example usage:
//
//	socketPath, err := iterm2.GetSocketPath()
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("iTerm2 socket: %s\n", socketPath)
//
//	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
//	    fmt.Println("Socket does not exist. Is iTerm2 running with Python API enabled?")
//	}
func GetSocketPath() (string, error) {
	return getSocketPath()
}
