package protocol

import (
	"context"
	"encoding/json"
	"testing"
)

// mockAdapter verifies the interface is implementable
type mockAdapter struct {
	sessions []Session
}

func (m *mockAdapter) Discover(ctx context.Context) ([]Session, error) {
	return m.sessions, nil
}

func (m *mockAdapter) Attach(ctx context.Context, sessionID string) (<-chan Event, error) {
	ch := make(chan Event, 1)
	ch <- Event{Kind: PayloadAssistantMessage, Seq: 1, Data: json.RawMessage(`{"text":"hi"}`)}
	close(ch)
	return ch, nil
}

func (m *mockAdapter) Send(ctx context.Context, sessionID string, msg UserInput) error {
	return nil
}

func (m *mockAdapter) Detach(sessionID string) error {
	return nil
}

func TestMockImplementsCopilotAdapter(t *testing.T) {
	var _ CopilotAdapter = (*mockAdapter)(nil)

	m := &mockAdapter{sessions: []Session{
		{ID: "s_1", Tool: "claude-code", Project: "/tmp", Status: SessionActive},
	}}

	sessions, err := m.Discover(context.Background())
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].Tool != "claude-code" {
		t.Errorf("tool: got %q", sessions[0].Tool)
	}

	ch, err := m.Attach(context.Background(), "s_1")
	if err != nil {
		t.Fatalf("Attach: %v", err)
	}
	event := <-ch
	if event.Kind != PayloadAssistantMessage {
		t.Errorf("event kind: got %q", event.Kind)
	}
}
