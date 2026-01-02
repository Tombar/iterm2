# Testing Guide

This library includes two types of tests: **unit tests** and **integration tests**.

## Unit Tests

Unit tests use mocks to test individual components without requiring iTerm2. They're fast and can run anywhere.

### Running Unit Tests

```bash
# Run all unit tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run TestGetID
```

### Coverage

Unit tests cover:
- `Tab.GetID()` - Returns tab identifier
- `Tab.Close()` - Sends close request with proper status handling
- `Tab.SetColor()` - Sets RGB color via profile properties
- RGB normalization logic (0-255 → 0-1 range)
- Error cases (tab not found, user declined, no sessions)

## Integration Tests

Integration tests run against a real iTerm2 instance. They verify actual behavior and protocol correctness.

### Prerequisites

1. **iTerm2 must be running** with at least one window open
2. **Python API must be enabled**: Go to iTerm2 → Preferences → General → Magic → Enable Python API

### Running Integration Tests

```bash
# Set environment variable to enable integration tests
export ITERM2_INTEGRATION_TESTS=1

# Run integration tests (requires iTerm2)
go test -tags=integration -v ./...

# Run specific integration test
go test -tags=integration -v -run TestIntegration_TabLifecycle
```

### Integration Test Coverage

- Full tab lifecycle (create → style → close)
- Multiple color changes
- Listing and closing multiple tabs
- Tab operations with sessions
- Error cases (closing already-closed tabs)

### What Integration Tests Verify

✅ Real WebSocket communication with iTerm2
✅ Protobuf message correctness
✅ Profile property format (JSON color structure)
✅ Tab color actually appears in iTerm2
✅ Close operations work as expected
✅ Error handling with real iTerm2 responses

## Testing Architecture

### ClientInterface

The library uses dependency injection via `ClientInterface`:

```go
type ClientInterface interface {
    Call(*api.ClientOriginatedMessage) (*api.ServerOriginatedMessage, error)
    Close() error
}
```

This allows:
- **Unit tests**: Use `mockClient` for fast, isolated testing
- **Integration tests**: Use real `*client.Client` for end-to-end validation
- **Your code**: Inject either for testing your sessh implementation

### Mock Client Example

```go
mock := &mockClient{
    responses: []*api.ServerOriginatedMessage{
        {
            Submessage: &api.ServerOriginatedMessage_CloseResponse{
                CloseResponse: &api.CloseResponse{
                    Statuses: []api.CloseResponse_Status{api.CloseResponse_OK},
                },
            },
        },
    },
}

tab := &tab{c: mock, id: "test-tab"}
err := tab.Close()
// Verify mock.calls to assert correct request was sent
```

## CI/CD Recommendations

### Fast Feedback Loop

```bash
# In CI: Run unit tests on every commit
go test ./...
```

### Full Validation

```bash
# In CI: Run integration tests on main branch or releases
# (requires macOS runner with iTerm2)
export ITERM2_INTEGRATION_TESTS=1
go test -tags=integration ./...
```

## Test Organization

```
iterm2/
├── tab_test.go              # Unit tests (always run)
│   ├── TestGetID
│   ├── TestClose
│   ├── TestSetColor
│   └── TestSetColor_RGBNormalization
├── integration_test.go      # Integration tests (opt-in)
│   ├── TestIntegration_TabLifecycle
│   └── TestIntegration_ErrorCases
└── TESTING.md              # This file
```

## Writing New Tests

### Adding Unit Tests

1. Use `mockClient` to control responses
2. Verify protobuf message structure
3. Test error paths
4. Keep tests fast (<1ms each)

```go
func TestNewFeature(t *testing.T) {
    mock := &mockClient{
        responses: []*api.ServerOriginatedMessage{/* ... */},
    }

    tab := &tab{c: mock, id: "test"}
    err := tab.NewFeature()

    // Assert on mock.calls to verify request structure
}
```

### Adding Integration Tests

1. Add `//go:build integration` tag at top
2. Check `ITERM2_INTEGRATION_TESTS` environment variable
3. Use real `iterm2.NewApp()`
4. Add cleanup with `defer tab.Close()`

```go
//go:build integration

func TestIntegration_NewFeature(t *testing.T) {
    if os.Getenv("ITERM2_INTEGRATION_TESTS") == "" {
        t.Skip("Set ITERM2_INTEGRATION_TESTS=1")
    }

    app, _ := iterm2.NewApp("test")
    defer app.Close()
    // Test with real iTerm2...
}
```

## Troubleshooting

### Integration Tests Not Running

**Problem**: `testing: warning: no tests to run`

**Solution**: Add `-tags=integration` flag:
```bash
go test -tags=integration ./...
```

### Integration Tests Skipped

**Problem**: Tests show as skipped

**Solution**: Set environment variable:
```bash
export ITERM2_INTEGRATION_TESTS=1
```

### Python API Not Enabled

**Problem**: `AppleScript/tell: execution error: iTerm got an error: The Python API is not enabled.`

**Solution**: Enable the Python API in iTerm2 preferences:
1. Open iTerm2 → Preferences (⌘,)
2. Go to the **General** tab
3. Find the **Magic** section
4. Check **"Enable Python API"**

**Note**: Despite the name "Python API", this setting is required for all automation including Go programs. iTerm2 uses this setting to control access to its WebSocket-based automation protocol.

### Connection Failed

**Problem**: `Failed to connect to iTerm2`

**Solutions**:
1. Ensure iTerm2 is running
2. Enable Python API: Preferences → General → Magic
3. Check socket exists: `ls ~/Library/Application\ Support/iTerm2/private/socket`

### Permission Denied

**Problem**: iTerm2 shows authorization dialog

**Solution**: Click "Allow" - the app name `iterm2-integration-test` will be remembered for future runs.
