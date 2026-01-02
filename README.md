# iTerm2

### Go library for automating iTerm2 Scripts

### Install

go get github.com/Tombar/iterm2

### Usage

#### Basic Usage

```golang
package main

import (
    "fmt"
    "github.com/Tombar/iterm2"
)

func main() {
    app, err := iterm2.NewApp("MyCoolPlugin")
    if err != nil {
        fmt.Printf("Failed to connect: %v\n", err)
        return
    }
    defer app.Close()

    // use app to create or list windows, tabs, and sessions and send various commands to the terminal.
}
```

#### Robust Usage with Prerequisite Checking

For production use, check prerequisites before connecting to provide better error messages:

```golang
package main

import (
    "errors"
    "fmt"
    "time"
    "github.com/Tombar/iterm2"
)

func main() {
    // Check if iTerm2 is running and Python API is enabled
    if err := iterm2.CheckPrerequisites("MyCoolPlugin"); err != nil {
        if errors.Is(err, iterm2.ErrITerm2NotRunning) {
            fmt.Println("iTerm2 is not running. Launching...")
            if err := iterm2.LaunchITerm2(); err != nil {
                fmt.Printf("Failed to launch iTerm2: %v\n", err)
                return
            }
            // Wait for iTerm2 to be ready
            if err := iterm2.WaitForITerm2(30 * time.Second); err != nil {
                fmt.Printf("iTerm2 did not start in time: %v\n", err)
                return
            }
        } else if errors.Is(err, iterm2.ErrPythonAPIDisabled) {
            fmt.Println(iterm2.EnablePythonAPIGuide())
            iterm2.OpenITerm2Preferences()
            return
        } else {
            fmt.Printf("Prerequisites check failed: %v\n", err)
            return
        }
    }

    // Request permission (shows dialog on first run)
    if err := iterm2.RequestPermission("MyCoolPlugin"); err != nil {
        if errors.Is(err, iterm2.ErrPermissionDenied) {
            fmt.Println("Permission denied. Please approve in iTerm2 settings.")
        }
        return
    }

    // Connect to iTerm2
    app, err := iterm2.NewApp("MyCoolPlugin")
    if err != nil {
        fmt.Printf("Failed to connect: %v\n", err)
        return
    }
    defer app.Close()

    // Your automation logic here...
}
```

### Error Handling

The library provides typed errors that you can check with `errors.Is()`:

- `ErrITerm2NotRunning` - iTerm2 is not running
- `ErrPythonAPIDisabled` - Python API is not enabled in Preferences
- `ErrPermissionDenied` - User denied permission for the application

### Helper Functions

- `CheckPrerequisites(appName)` - Verify iTerm2 is running and API is enabled
- `RequestPermission(appName)` - Test authorization (triggers dialog on first run)
- `LaunchITerm2()` - Launch iTerm2 if not already running
- `WaitForITerm2(timeout)` - Wait for iTerm2 to be ready
- `GetSocketPath()` - Get the Unix socket path for debugging
- `EnablePythonAPIGuide()` - Get formatted instructions for enabling the Python API
- `OpenITerm2Preferences()` - Open iTerm2 Preferences window

### How do I actually run the script?

- Since you will be using this library in a "main" program, you can literally just run the Go program through "go run" or install your program/binary globally through "go install" and then run it from any terminal.

- A nicer way to run the script is to "register" the plugin with iTerm2 so you can run it from iTerm's command pallette (cmd+shift+o). This means you won't need a terminal tab open or to remember what the plugin name is. See the following section on how to do that:

- Ensure you enable the Python API: https://iterm2.com/python-api-auth.html

### Installing the plugin into iTerm2

1. Install your program into your local PATH (for example running `go install`)
2. `go get github.com/Tombar/iterm2/cmd/goiterm`
3. `goiterm install <bin>`
4. From any iTerm window run "cmd+shift+o" and look for `<bin>.py`.

### Progress

This is currently a work in progress and it is a subset of what the iTerm2 WebSocket protocol provides.

I don't intend to implement all of it as I am mainly implementing only the parts that I need for daily work. 

If you'd like to add more features, feel free to open an issue and a PR. 
