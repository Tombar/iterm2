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
- **Tab operations**:
  - `Tab.GetID()` - Returns tab identifier
  - `Tab.Close()` - Sends close request with proper status handling
  - `Tab.SetColor()` - Sets RGB color via profile properties
  - RGB normalization logic (0-255 → 0-1 range)
  - Error cases (tab not found, user declined, no sessions)

- **Prerequisite checking** (errors_test.go):
  - Sentinel error detection with `errors.Is()`
  - iTerm2 running detection
  - Python API enabled detection
  - Socket path construction
  - Error message wrapping and classification

- **Helper functions** (helpers_test.go):
  - Socket path retrieval
  - Python API setup guide generation
  - WaitForITerm2 timeout behavior
  - LaunchITerm2 idempotency

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

- **Tab lifecycle** (TestIntegration_TabLifecycle):
  - Full tab lifecycle (create → style → close)
  - Multiple color changes
  - Listing and closing multiple tabs
  - Tab operations with sessions

- **Error cases** (TestIntegration_ErrorCases):
  - Closing already-closed tabs

- **Prerequisites** (TestIntegration_Prerequisites):
  - CheckPrerequisites() with real iTerm2
  - RequestPermission() with authorization
  - Helper function validation (socket path, guide generation)
  - Graceful handling when Python API is disabled

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
├── tab_test.go              # Tab operation unit tests
│   ├── TestGetID
│   ├── TestClose
│   ├── TestSetColor
│   └── TestSetColor_RGBNormalization
├── errors_test.go           # Prerequisite checking unit tests
│   ├── TestSentinelErrors
│   ├── TestGetSocketPath
│   ├── TestIsITerm2Running
│   ├── TestIsPythonAPIEnabled
│   ├── TestCheckPrerequisites_ErrorMessages
│   └── TestEnhanceConnectionError
├── helpers_test.go          # Helper function unit tests
│   ├── TestGetSocketPath_Public
│   ├── TestEnablePythonAPIGuide
│   ├── TestWaitForITerm2_Timeout
│   ├── TestLaunchITerm2_IdempotentWhenRunning
│   └── TestOpenITerm2Preferences
├── integration_test.go      # Integration tests (opt-in)
│   ├── TestIntegration_TabLifecycle
│   ├── TestIntegration_ErrorCases
│   └── TestIntegration_Prerequisites
└── TESTING.md              # This file
```

## Testing Prerequisite Checking in Your Code

When writing applications that use this library, you can test prerequisite handling:

```go
func TestMyApp_HandlesITerm2NotRunning(t *testing.T) {
    // Prerequisite checks don't need mocking - they work on real system state
    err := iterm2.CheckPrerequisites("test-app")

    if err != nil {
        // Test your error handling logic
        if errors.Is(err, iterm2.ErrITerm2NotRunning) {
            // Verify your app handles this gracefully
        }
    }
}
```

The prerequisite functions are designed to be environment-aware:
- `CheckPrerequisites()` checks actual system state (process list, socket files)
- Tests automatically adapt based on whether iTerm2 is running
- No mocking required for prerequisite testing

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

### Quick Diagnostic

Use the prerequisite checking functions to diagnose issues:

```go
// Check if iTerm2 is running
if err := iterm2.CheckPrerequisites("diagnostic"); err != nil {
    if errors.Is(err, iterm2.ErrITerm2NotRunning) {
        fmt.Println("iTerm2 is not running")
    } else if errors.Is(err, iterm2.ErrPythonAPIDisabled) {
        fmt.Println("Python API is not enabled")
        fmt.Println(iterm2.EnablePythonAPIGuide())
    }
}

// Check socket path
if path, err := iterm2.GetSocketPath(); err == nil {
    fmt.Printf("Socket path: %s\n", path)
}
```

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

**Solution**: Use the built-in helper to get setup instructions:

```go
fmt.Println(iterm2.EnablePythonAPIGuide())
iterm2.OpenITerm2Preferences()  // Opens preferences automatically
```

Or manually enable:
1. Open iTerm2 → Preferences (⌘,)
2. Go to the **General** tab
3. Find the **Magic** section
4. Check **"Enable Python API"**

**Note**: Despite the name "Python API", this setting is required for all automation including Go programs. iTerm2 uses this setting to control access to its WebSocket-based automation protocol.

### Connection Failed

**Problem**: `Failed to connect to iTerm2`

**Quick diagnosis**:
```go
if err := iterm2.CheckPrerequisites("test"); err != nil {
    log.Printf("Prerequisite failed: %v", err)
}
```

**Solutions**:
1. Check if iTerm2 is running:
   ```go
   if err := iterm2.LaunchITerm2(); err != nil {
       log.Fatal(err)
   }
   iterm2.WaitForITerm2(30 * time.Second)
   ```

2. Enable Python API: Preferences → General → Magic
   ```go
   iterm2.OpenITerm2Preferences()
   ```

3. Verify socket exists:
   ```bash
   ls ~/Library/Application\ Support/iTerm2/private/socket
   ```
   Or programmatically:
   ```go
   path, _ := iterm2.GetSocketPath()
   fmt.Println(path)
   ```

### Permission Denied

**Problem**: iTerm2 shows authorization dialog

**Solution**: Click "Allow" - the app name will be remembered for future runs.

**Testing permission proactively**:
```go
if err := iterm2.RequestPermission("myapp"); err != nil {
    if errors.Is(err, iterm2.ErrPermissionDenied) {
        fmt.Println("User denied permission. Please approve in iTerm2 settings.")
    }
}
```

This function:
- Triggers the authorization dialog on first run
- Returns immediately on subsequent runs (permission is cached)
- Allows graceful error handling before attempting real operations
