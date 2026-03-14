package adapter

import (
	"context"
	"fmt"

	"github.com/relixdev/protocol"
)

// MultiAdapter multiplexes multiple CopilotAdapters behind a single interface.
// Discover returns sessions from all adapters. Attach/Send/Detach route to the
// adapter that owns the session (determined during Attach).
type MultiAdapter struct {
	adapters []protocol.CopilotAdapter

	// ownership maps sessionID → adapter index, set during Attach.
	// Protected by the caller (daemon) which serializes reconcile calls.
	ownership map[string]int
}

// NewMultiAdapter creates a multiplexer over the given adapters.
func NewMultiAdapter(adapters ...protocol.CopilotAdapter) *MultiAdapter {
	return &MultiAdapter{
		adapters:  adapters,
		ownership: make(map[string]int),
	}
}

// Discover returns sessions from all underlying adapters.
func (m *MultiAdapter) Discover(ctx context.Context) ([]protocol.Session, error) {
	var all []protocol.Session
	for _, a := range m.adapters {
		sessions, err := a.Discover(ctx)
		if err != nil {
			continue // best-effort: don't let one adapter break discovery
		}
		all = append(all, sessions...)
	}
	return all, nil
}

// Attach delegates to each adapter in order until one succeeds.
func (m *MultiAdapter) Attach(ctx context.Context, sessionID string) (<-chan protocol.Event, error) {
	for i, a := range m.adapters {
		ch, err := a.Attach(ctx, sessionID)
		if err == nil {
			m.ownership[sessionID] = i
			return ch, nil
		}
	}
	return nil, fmt.Errorf("multi: no adapter could attach session %q", sessionID)
}

// Send routes to the adapter that owns the session.
func (m *MultiAdapter) Send(ctx context.Context, sessionID string, msg protocol.UserInput) error {
	idx, ok := m.ownership[sessionID]
	if !ok {
		return fmt.Errorf("multi: session %q not owned by any adapter", sessionID)
	}
	return m.adapters[idx].Send(ctx, sessionID, msg)
}

// Detach routes to the adapter that owns the session.
func (m *MultiAdapter) Detach(sessionID string) error {
	idx, ok := m.ownership[sessionID]
	if !ok {
		return nil
	}
	delete(m.ownership, sessionID)
	return m.adapters[idx].Detach(sessionID)
}
