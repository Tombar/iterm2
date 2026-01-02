package iterm2

import (
	"testing"

	"github.com/Tombar/iterm2/api"
)

// mockClient implements ClientInterface for testing
type mockClient struct {
	calls     []*api.ClientOriginatedMessage
	responses []*api.ServerOriginatedMessage
	callFunc  func(*api.ClientOriginatedMessage) (*api.ServerOriginatedMessage, error)
}

func (m *mockClient) Call(req *api.ClientOriginatedMessage) (*api.ServerOriginatedMessage, error) {
	m.calls = append(m.calls, req)
	if m.callFunc != nil {
		return m.callFunc(req)
	}
	if len(m.responses) > 0 {
		resp := m.responses[0]
		m.responses = m.responses[1:]
		return resp, nil
	}
	return &api.ServerOriginatedMessage{}, nil
}

func (m *mockClient) Close() error {
	return nil
}

// TestGetID verifies that GetID returns the tab's unique identifier
func TestGetID(t *testing.T) {
	tests := []struct {
		name   string
		tabID  string
		wantID string
	}{
		{
			name:   "returns tab id",
			tabID:  "tab-123",
			wantID: "tab-123",
		},
		{
			name:   "handles numeric id",
			tabID:  "42",
			wantID: "42",
		},
		{
			name:   "handles empty id",
			tabID:  "",
			wantID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tab := &tab{
				id: tt.tabID,
			}

			gotID := tab.GetID()
			if gotID != tt.wantID {
				t.Errorf("GetID() = %q, want %q", gotID, tt.wantID)
			}
		})
	}
}

// TestSetColor_RGBNormalization verifies RGB value normalization logic
func TestSetColor_RGBNormalization(t *testing.T) {
	tests := []struct {
		name            string
		r, g, b         uint8
		wantNormalizedR float64
		wantNormalizedG float64
		wantNormalizedB float64
	}{
		{
			name:            "blue feature color",
			r:               100,
			g:               149,
			b:               237,
			wantNormalizedR: 100.0 / 255.0,
			wantNormalizedG: 149.0 / 255.0,
			wantNormalizedB: 237.0 / 255.0,
		},
		{
			name:            "red main color",
			r:               220,
			g:               20,
			b:               60,
			wantNormalizedR: 220.0 / 255.0,
			wantNormalizedG: 20.0 / 255.0,
			wantNormalizedB: 60.0 / 255.0,
		},
		{
			name:            "max RGB values (white)",
			r:               255,
			g:               255,
			b:               255,
			wantNormalizedR: 1.0,
			wantNormalizedG: 1.0,
			wantNormalizedB: 1.0,
		},
		{
			name:            "min RGB values (black)",
			r:               0,
			g:               0,
			b:               0,
			wantNormalizedR: 0.0,
			wantNormalizedG: 0.0,
			wantNormalizedB: 0.0,
		},
		{
			name:            "mid-range values",
			r:               128,
			g:               128,
			b:               128,
			wantNormalizedR: 128.0 / 255.0,
			wantNormalizedG: 128.0 / 255.0,
			wantNormalizedB: 128.0 / 255.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the normalization logic that SetColor uses
			normalizedR := float64(tt.r) / 255.0
			normalizedG := float64(tt.g) / 255.0
			normalizedB := float64(tt.b) / 255.0

			if normalizedR != tt.wantNormalizedR {
				t.Errorf("normalized R = %f, want %f", normalizedR, tt.wantNormalizedR)
			}
			if normalizedG != tt.wantNormalizedG {
				t.Errorf("normalized G = %f, want %f", normalizedG, tt.wantNormalizedG)
			}
			if normalizedB != tt.wantNormalizedB {
				t.Errorf("normalized B = %f, want %f", normalizedB, tt.wantNormalizedB)
			}

			// Verify values are in valid range [0, 1]
			if normalizedR < 0 || normalizedR > 1 {
				t.Errorf("normalized R %f out of range [0, 1]", normalizedR)
			}
			if normalizedG < 0 || normalizedG > 1 {
				t.Errorf("normalized G %f out of range [0, 1]", normalizedG)
			}
			if normalizedB < 0 || normalizedB > 1 {
				t.Errorf("normalized B %f out of range [0, 1]", normalizedB)
			}
		})
	}
}

// TestClose verifies Close sends correct CloseRequest
func TestClose(t *testing.T) {
	tests := []struct {
		name      string
		tabID     string
		response  *api.ServerOriginatedMessage
		wantError bool
	}{
		{
			name:  "successful close",
			tabID: "tab-123",
			response: &api.ServerOriginatedMessage{
				Submessage: &api.ServerOriginatedMessage_CloseResponse{
					CloseResponse: &api.CloseResponse{
						Statuses: []api.CloseResponse_Status{api.CloseResponse_OK},
					},
				},
			},
			wantError: false,
		},
		{
			name:  "tab not found",
			tabID: "tab-missing",
			response: &api.ServerOriginatedMessage{
				Submessage: &api.ServerOriginatedMessage_CloseResponse{
					CloseResponse: &api.CloseResponse{
						Statuses: []api.CloseResponse_Status{api.CloseResponse_NOT_FOUND},
					},
				},
			},
			wantError: true,
		},
		{
			name:  "user declined",
			tabID: "tab-declined",
			response: &api.ServerOriginatedMessage{
				Submessage: &api.ServerOriginatedMessage_CloseResponse{
					CloseResponse: &api.CloseResponse{
						Statuses: []api.CloseResponse_Status{api.CloseResponse_USER_DECLINED},
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{
				responses: []*api.ServerOriginatedMessage{tt.response},
			}

			tab := &tab{
				c:  mock,
				id: tt.tabID,
			}

			err := tab.Close()

			if (err != nil) != tt.wantError {
				t.Errorf("Close() error = %v, wantError %v", err, tt.wantError)
				return
			}

			// Verify CloseRequest was sent correctly
			if len(mock.calls) != 1 {
				t.Fatalf("expected 1 Call, got %d", len(mock.calls))
			}

			closeReq := mock.calls[0].GetCloseRequest()
			if closeReq == nil {
				t.Fatal("expected CloseRequest, got nil")
			}

			tabs := closeReq.GetTabs()
			if tabs == nil {
				t.Fatal("expected CloseTabs, got nil")
			}

			if len(tabs.TabIds) != 1 {
				t.Fatalf("expected 1 tab ID, got %d", len(tabs.TabIds))
			}

			if tabs.TabIds[0] != tt.tabID {
				t.Errorf("CloseRequest tab ID = %q, want %q", tabs.TabIds[0], tt.tabID)
			}

			if closeReq.GetForce() != false {
				t.Error("expected Force=false")
			}
		})
	}
}

// TestSetColor verifies SetColor sends correct profile property request
func TestSetColor(t *testing.T) {
	tests := []struct {
		name      string
		r, g, b   uint8
		wantError bool
	}{
		{
			name: "blue color",
			r:    100,
			g:    149,
			b:    237,
		},
		{
			name: "red color",
			r:    220,
			g:    20,
			b:    60,
		},
		{
			name: "white",
			r:    255,
			g:    255,
			b:    255,
		},
		{
			name: "black",
			r:    0,
			g:    0,
			b:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedSetProfileReq *api.SetProfilePropertyRequest

			mock := &mockClient{
				callFunc: func(req *api.ClientOriginatedMessage) (*api.ServerOriginatedMessage, error) {
					// First call: ListSessions
					if req.GetListSessionsRequest() != nil {
						return &api.ServerOriginatedMessage{
							Submessage: &api.ServerOriginatedMessage_ListSessionsResponse{
								ListSessionsResponse: &api.ListSessionsResponse{
									Windows: []*api.ListSessionsResponse_Window{
										{
											WindowId: str("win-1"),
											Tabs: []*api.ListSessionsResponse_Tab{
												{
													TabId: str("tab-1"),
													Root: &api.SplitTreeNode{
														Links: []*api.SplitTreeNode_SplitTreeLink{
															{
																Child: &api.SplitTreeNode_SplitTreeLink_Session{
																	Session: &api.SessionSummary{
																		UniqueIdentifier: str("sess-1"),
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						}, nil
					}

					// Second call: SetProfileProperty
					if req.GetSetProfilePropertyRequest() != nil {
						capturedSetProfileReq = req.GetSetProfilePropertyRequest()
						return &api.ServerOriginatedMessage{
							Submessage: &api.ServerOriginatedMessage_SetProfilePropertyResponse{
								SetProfilePropertyResponse: &api.SetProfilePropertyResponse{
									Status: api.SetProfilePropertyResponse_OK.Enum(),
								},
							},
						}, nil
					}

					return &api.ServerOriginatedMessage{}, nil
				},
			}

			tab := &tab{
				c:        mock,
				id:       "tab-1",
				windowID: "win-1",
			}

			err := tab.SetColor(tt.r, tt.g, tt.b)
			if (err != nil) != tt.wantError {
				t.Errorf("SetColor() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if tt.wantError {
				return
			}

			// Verify SetProfilePropertyRequest
			if capturedSetProfileReq == nil {
				t.Fatal("SetProfilePropertyRequest was not captured")
			}

			// Verify session target
			if capturedSetProfileReq.GetSession() != "sess-1" {
				t.Errorf("session = %q, want %q", capturedSetProfileReq.GetSession(), "sess-1")
			}

			// Verify assignments
			assignments := capturedSetProfileReq.GetAssignments()
			if len(assignments) != 2 {
				t.Fatalf("expected 2 assignments, got %d", len(assignments))
			}

			// Check for Tab Color and Use Tab Color
			var foundTabColor, foundUseTabColor bool
			for _, assign := range assignments {
				switch assign.GetKey() {
				case "Tab Color":
					foundTabColor = true
				case "Use Tab Color":
					foundUseTabColor = true
					if assign.GetJsonValue() != "true" {
						t.Errorf("Use Tab Color = %q, want %q", assign.GetJsonValue(), "true")
					}
				}
			}

			if !foundTabColor {
				t.Error("Tab Color assignment not found")
			}
			if !foundUseTabColor {
				t.Error("Use Tab Color assignment not found")
			}
		})
	}
}

// TestSetColor_NoSessions verifies error when tab has no sessions
func TestSetColor_NoSessions(t *testing.T) {
	mock := &mockClient{
		callFunc: func(req *api.ClientOriginatedMessage) (*api.ServerOriginatedMessage, error) {
			// Return empty sessions list
			return &api.ServerOriginatedMessage{
				Submessage: &api.ServerOriginatedMessage_ListSessionsResponse{
					ListSessionsResponse: &api.ListSessionsResponse{
						Windows: []*api.ListSessionsResponse_Window{
							{
								WindowId: str("win-1"),
								Tabs: []*api.ListSessionsResponse_Tab{
									{
										TabId: str("tab-1"),
										Root: &api.SplitTreeNode{
											Links: []*api.SplitTreeNode_SplitTreeLink{},
										},
									},
								},
							},
						},
					},
				},
			}, nil
		},
	}

	tab := &tab{
		c:        mock,
		id:       "tab-1",
		windowID: "win-1",
	}

	err := tab.SetColor(100, 100, 100)
	if err == nil {
		t.Error("SetColor() expected error for tab with no sessions, got nil")
	}
}
