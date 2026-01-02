//go:build integration
// +build integration

package iterm2_test

import (
	"os"
	"testing"
	"time"

	"github.com/Tombar/iterm2"
)

// TestIntegration_TabLifecycle tests the full lifecycle of tab operations with real iTerm2
func TestIntegration_TabLifecycle(t *testing.T) {
	if os.Getenv("ITERM2_INTEGRATION_TESTS") == "" {
		t.Skip("Set ITERM2_INTEGRATION_TESTS=1 to run integration tests")
	}

	app, err := iterm2.NewApp("iterm2-integration-test")
	if err != nil {
		t.Fatalf("Failed to connect to iTerm2: %v\nEnsure iTerm2 is running and Python API is enabled", err)
	}
	defer app.Close()

	// Get or create a window
	windows, err := app.ListWindows()
	if err != nil {
		t.Fatalf("Failed to list windows: %v", err)
	}
	if len(windows) == 0 {
		t.Fatal("No windows available - please open iTerm2 with at least one window")
	}

	window := windows[0]

	t.Run("create_and_style_tab", func(t *testing.T) {
		// Create a new tab
		tab, err := window.CreateTab()
		if err != nil {
			t.Fatalf("Failed to create tab: %v", err)
		}

		// Verify GetID returns a non-empty ID
		tabID := tab.GetID()
		if tabID == "" {
			t.Error("GetID() returned empty string")
		}
		t.Logf("Created tab with ID: %s", tabID)

		// Set tab title
		if err := tab.SetTitle("Integration Test Tab"); err != nil {
			t.Errorf("Failed to set tab title: %v", err)
		}

		// Set tab color (blue)
		if err := tab.SetColor(100, 149, 237); err != nil {
			t.Errorf("Failed to set tab color: %v", err)
		}

		// Give iTerm2 time to apply changes
		time.Sleep(500 * time.Millisecond)

		// Close the tab
		if err := tab.Close(); err != nil {
			t.Errorf("Failed to close tab: %v", err)
		}

		t.Log("Successfully created, styled, and closed tab")
	})

	t.Run("multiple_color_changes", func(t *testing.T) {
		tab, err := window.CreateTab()
		if err != nil {
			t.Fatalf("Failed to create tab: %v", err)
		}
		defer tab.Close()

		colors := []struct {
			name    string
			r, g, b uint8
		}{
			{"blue", 100, 149, 237},
			{"red", 220, 20, 60},
			{"green", 46, 139, 87},
		}

		for _, color := range colors {
			if err := tab.SetColor(color.r, color.g, color.b); err != nil {
				t.Errorf("Failed to set %s color: %v", color.name, err)
			}
			time.Sleep(300 * time.Millisecond)
		}

		t.Log("Successfully changed colors multiple times")
	})

	t.Run("list_and_close_tabs", func(t *testing.T) {
		// Create multiple tabs
		var tabs []iterm2.Tab
		for i := 0; i < 3; i++ {
			tab, err := window.CreateTab()
			if err != nil {
				t.Fatalf("Failed to create tab %d: %v", i, err)
			}
			tabs = append(tabs, tab)
		}

		// List tabs
		listedTabs, err := window.ListTabs()
		if err != nil {
			t.Fatalf("Failed to list tabs: %v", err)
		}

		if len(listedTabs) < 3 {
			t.Errorf("Expected at least 3 tabs, got %d", len(listedTabs))
		}

		// Close all created tabs
		for i, tab := range tabs {
			if err := tab.Close(); err != nil {
				t.Errorf("Failed to close tab %d: %v", i, err)
			}
		}

		t.Log("Successfully created, listed, and closed multiple tabs")
	})

	t.Run("tab_with_session_operations", func(t *testing.T) {
		tab, err := window.CreateTab()
		if err != nil {
			t.Fatalf("Failed to create tab: %v", err)
		}
		defer tab.Close()

		// Set color before getting sessions
		if err := tab.SetColor(255, 215, 0); err != nil {
			t.Fatalf("Failed to set color: %v", err)
		}

		// List sessions in the tab
		sessions, err := tab.ListSessions()
		if err != nil {
			t.Fatalf("Failed to list sessions: %v", err)
		}

		if len(sessions) == 0 {
			t.Error("Expected at least one session in tab")
		}

		t.Logf("Tab has %d session(s)", len(sessions))
	})
}

// TestIntegration_ErrorCases tests error handling with real iTerm2
func TestIntegration_ErrorCases(t *testing.T) {
	if os.Getenv("ITERM2_INTEGRATION_TESTS") == "" {
		t.Skip("Set ITERM2_INTEGRATION_TESTS=1 to run integration tests")
	}

	app, err := iterm2.NewApp("iterm2-integration-test")
	if err != nil {
		t.Fatalf("Failed to connect to iTerm2: %v", err)
	}
	defer app.Close()

	t.Run("close_already_closed_tab", func(t *testing.T) {
		windows, err := app.ListWindows()
		if err != nil || len(windows) == 0 {
			t.Skip("No windows available")
		}

		tab, err := windows[0].CreateTab()
		if err != nil {
			t.Fatalf("Failed to create tab: %v", err)
		}

		// Close once
		if err := tab.Close(); err != nil {
			t.Fatalf("First close failed: %v", err)
		}

		// Try to close again - should handle gracefully
		err = tab.Close()
		// Note: iTerm2 may return NOT_FOUND or handle gracefully
		t.Logf("Second close returned: %v", err)
	})
}
