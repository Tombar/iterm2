package iterm2

import (
	"fmt"

	"marwan.io/iterm2/api"
)

// Tab abstracts an iTerm2 window tab
type Tab interface {
	SetTitle(string) error
	ListSessions() ([]Session, error)
	SetColor(r, g, b uint8) error
	Close() error
	GetID() string
}

type tab struct {
	c        ClientInterface
	id       string
	windowID string
}

func (t *tab) SetTitle(s string) error {
	_, err := t.c.Call(&api.ClientOriginatedMessage{
		Submessage: &api.ClientOriginatedMessage_InvokeFunctionRequest{
			InvokeFunctionRequest: &api.InvokeFunctionRequest{
				Invocation: str(fmt.Sprintf(`iterm2.set_title(title: "%s")`, s)),
				Context: &api.InvokeFunctionRequest_Method_{
					Method: &api.InvokeFunctionRequest_Method{
						Receiver: &t.id,
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("could not call set_title: %w", err)
	}
	return nil
}

func (t *tab) ListSessions() ([]Session, error) {
	list := []Session{}
	resp, err := t.c.Call(&api.ClientOriginatedMessage{
		Submessage: &api.ClientOriginatedMessage_ListSessionsRequest{
			ListSessionsRequest: &api.ListSessionsRequest{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing sessions for tab %q: %w", t.id, err)
	}
	lsr := resp.GetListSessionsResponse()
	for _, window := range lsr.GetWindows() {
		if window.GetWindowId() != t.windowID {
			continue
		}
		for _, wt := range window.GetTabs() {
			if wt.GetTabId() != t.id {
				continue
			}
			for _, link := range wt.GetRoot().GetLinks() {
				list = append(list, &session{
					c:  t.c,
					id: link.GetSession().GetUniqueIdentifier(),
				})
			}
		}
	}
	return list, nil
}

// GetID returns the unique identifier for this tab
func (t *tab) GetID() string {
	return t.id
}

// SetColor sets the tab's background color using RGB values (0-255)
func (t *tab) SetColor(r, g, b uint8) error {
	// Get the first session in the tab to set its profile property
	sessions, err := t.ListSessions()
	if err != nil {
		return fmt.Errorf("could not list sessions for tab %q: %w", t.id, err)
	}
	if len(sessions) == 0 {
		return fmt.Errorf("tab %q has no sessions", t.id)
	}

	sess, ok := sessions[0].(*session)
	if !ok {
		return fmt.Errorf("session type assertion failed")
	}

	// Set both tab color and use_tab_color properties
	// RGB values need to be normalized to 0-1 range for iTerm2
	colorJSON := fmt.Sprintf(`{"Red Component": %f, "Green Component": %f, "Blue Component": %f}`,
		float64(r)/255.0, float64(g)/255.0, float64(b)/255.0)

	_, err = t.c.Call(&api.ClientOriginatedMessage{
		Submessage: &api.ClientOriginatedMessage_SetProfilePropertyRequest{
			SetProfilePropertyRequest: &api.SetProfilePropertyRequest{
				Target: &api.SetProfilePropertyRequest_Session{
					Session: sess.id,
				},
				Assignments: []*api.SetProfilePropertyRequest_Assignment{
					{
						Key:       str("Tab Color"),
						JsonValue: str(colorJSON),
					},
					{
						Key:       str("Use Tab Color"),
						JsonValue: str("true"),
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("could not set color for tab %q: %w", t.id, err)
	}
	return nil
}

// Close closes this tab
func (t *tab) Close() error {
	resp, err := t.c.Call(&api.ClientOriginatedMessage{
		Submessage: &api.ClientOriginatedMessage_CloseRequest{
			CloseRequest: &api.CloseRequest{
				Target: &api.CloseRequest_Tabs{
					Tabs: &api.CloseRequest_CloseTabs{
						TabIds: []string{t.id},
					},
				},
				Force: b(false),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("could not close tab %q: %w", t.id, err)
	}

	closeResp := resp.GetCloseResponse()
	if len(closeResp.GetStatuses()) > 0 {
		status := closeResp.GetStatuses()[0]
		if status != api.CloseResponse_OK {
			return fmt.Errorf("failed to close tab %q: status %v", t.id, status)
		}
	}
	return nil
}
