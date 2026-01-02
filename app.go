package iterm2

import (
	"fmt"
	"io"
	"strings"

	"github.com/Tombar/iterm2/api"
	"github.com/Tombar/iterm2/client"
)

// App represents an open iTerm2 application
type App interface {
	io.Closer

	CreateWindow() (Window, error)
	ListWindows() ([]Window, error)
	SelectMenuItem(item string) error
	Activate(raiseAllWindows, ignoreOtherApps bool) error
}

// NewApp establishes a connection with iTerm2 and returns an App.
//
// For better error handling and user guidance, consider using
// CheckPrerequisites() and RequestPermission() before calling NewApp():
//
//	if err := iterm2.CheckPrerequisites("MyApp"); err != nil {
//	    if errors.Is(err, iterm2.ErrITerm2NotRunning) {
//	        iterm2.LaunchITerm2()
//	        iterm2.WaitForITerm2(30 * time.Second)
//	    } else if errors.Is(err, iterm2.ErrPythonAPIDisabled) {
//	        fmt.Println(iterm2.EnablePythonAPIGuide())
//	        iterm2.OpenITerm2Preferences()
//	        return
//	    }
//	}
//	if err := iterm2.RequestPermission("MyApp"); err != nil {
//	    // Handle permission denied
//	}
//	app, err := iterm2.NewApp("MyApp")
//
// NewApp returns typed errors that can be checked with errors.Is():
//   - ErrITerm2NotRunning: iTerm2 is not running
//   - ErrPythonAPIDisabled: Python API is not enabled in Preferences
//   - ErrPermissionDenied: User denied permission for this application
//
// Name is used to register your application with iTerm2 so that it doesn't
// require explicit permissions every time you run the plugin. The name appears
// in iTerm2's authorization dialog on first run.
func NewApp(name string) (App, error) {
	c, err := client.New(name)
	if err != nil {
		// Enhance error with typed sentinels for better error handling
		return nil, enhanceConnectionError(err, name)
	}

	return &app{c: c}, nil
}

// enhanceConnectionError wraps client connection errors with typed sentinels.
// This allows users to programmatically detect and handle specific failure modes.
func enhanceConnectionError(err error, appName string) error {
	errMsg := strings.ToLower(err.Error())

	// Check for iTerm2 not running (socket connection failure)
	if strings.Contains(errMsg, "no such file or directory") ||
		strings.Contains(errMsg, "connection refused") {
		if !isITerm2Running() {
			return fmt.Errorf("%w: %v", ErrITerm2NotRunning, err)
		}
		// If iTerm2 is running but socket doesn't exist, API is disabled
		return fmt.Errorf("%w: %v", ErrPythonAPIDisabled, err)
	}

	// Check for Python API disabled
	if isPythonAPIError(err) {
		return fmt.Errorf("%w: %v", ErrPythonAPIDisabled, err)
	}

	// Check for permission denied
	if isPermissionError(err) {
		return fmt.Errorf("%w: %v", ErrPermissionDenied, err)
	}

	// Unknown error - return as-is
	return err
}

type app struct {
	c *client.Client
}

func (a *app) Activate(raiseAllWindows bool, ignoreOtherApps bool) error {
	_, err := a.c.Call(&api.ClientOriginatedMessage{
		Submessage: &api.ClientOriginatedMessage_ActivateRequest{ActivateRequest: &api.ActivateRequest{
			OrderWindowFront: b(true),
			ActivateApp: &api.ActivateRequest_App{
				RaiseAllWindows:   &raiseAllWindows,
				IgnoringOtherApps: &ignoreOtherApps,
			},
		}},
	})
	return err
}

func (a *app) CreateWindow() (Window, error) {
	resp, err := a.c.Call(&api.ClientOriginatedMessage{
		Submessage: &api.ClientOriginatedMessage_CreateTabRequest{
			CreateTabRequest: &api.CreateTabRequest{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("could not create window tab: %w", err)
	}
	ctr := resp.GetCreateTabResponse()
	if ctr.GetStatus() != api.CreateTabResponse_OK {
		return nil, fmt.Errorf("unexpected window tab status: %s", ctr.GetStatus())
	}
	return &window{
		c:       a.c,
		id:      ctr.GetWindowId(),
		session: ctr.GetSessionId(),
	}, nil
}

func (a *app) ListWindows() ([]Window, error) {
	list := []Window{}
	resp, err := a.c.Call(&api.ClientOriginatedMessage{
		Submessage: &api.ClientOriginatedMessage_ListSessionsRequest{
			ListSessionsRequest: &api.ListSessionsRequest{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("could not list sessions: %w", err)
	}
	for _, w := range resp.GetListSessionsResponse().GetWindows() {
		list = append(list, &window{
			c:  a.c,
			id: w.GetWindowId(),
		})
	}
	return list, nil
}

func (a *app) Close() error {
	return a.c.Close()
}

func str(s string) *string {
	return &s
}

func b(b bool) *bool {
	return &b
}

func (a *app) SelectMenuItem(item string) error {
	resp, err := a.c.Call(&api.ClientOriginatedMessage{
		Submessage: &api.ClientOriginatedMessage_MenuItemRequest{
			MenuItemRequest: &api.MenuItemRequest{
				Identifier: &item,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error selecting menu item %q: %w", item, err)
	}
	if resp.GetMenuItemResponse().GetStatus() != api.MenuItemResponse_OK {
		return fmt.Errorf("menu item %q returned unexpected status: %q", item, resp.GetMenuItemResponse().GetStatus().String())
	}
	return nil
}
