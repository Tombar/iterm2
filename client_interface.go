package iterm2

import "marwan.io/iterm2/api"

// ClientInterface defines the interface for communicating with iTerm2.
// This abstraction enables testing by allowing mock implementations.
type ClientInterface interface {
	// Call sends a request to the iTerm2 server and returns the response
	Call(*api.ClientOriginatedMessage) (*api.ServerOriginatedMessage, error)

	// Close closes the connection to iTerm2
	Close() error
}
