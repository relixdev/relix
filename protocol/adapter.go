package protocol

import "context"

// CopilotAdapter is the interface that each coding tool integration must implement.
// The agent daemon uses this to bridge between a coding tool and the relay.
type CopilotAdapter interface {
	// Discover returns active sessions for this tool on the machine.
	Discover(ctx context.Context) ([]Session, error)

	// Attach connects to a specific session and returns a channel of events.
	// The channel is closed when the session ends or Detach is called.
	Attach(ctx context.Context, sessionID string) (<-chan Event, error)

	// Send delivers user input or an approval response to the session.
	Send(ctx context.Context, sessionID string, msg UserInput) error

	// Detach cleanly disconnects from a session.
	Detach(sessionID string) error
}
