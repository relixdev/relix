package daemon_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/relixdev/protocol"
	"github.com/relixdev/relix/relixctl/internal/daemon"
)

// --- mock adapter ---

type mockAdapter struct {
	mu       sync.Mutex
	sessions []protocol.Session

	attachCalls  []string
	detachCalls  []string
	sendCalls    []protocol.UserInput
	attachChans  map[string]chan protocol.Event
}

func newMockAdapter(sessions []protocol.Session) *mockAdapter {
	return &mockAdapter{
		sessions:    sessions,
		attachChans: make(map[string]chan protocol.Event),
	}
}

func (m *mockAdapter) Discover(_ context.Context) ([]protocol.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]protocol.Session, len(m.sessions))
	copy(out, m.sessions)
	return out, nil
}

func (m *mockAdapter) Attach(_ context.Context, sessionID string) (<-chan protocol.Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.attachCalls = append(m.attachCalls, sessionID)
	ch := make(chan protocol.Event, 8)
	m.attachChans[sessionID] = ch
	return ch, nil
}

func (m *mockAdapter) Send(_ context.Context, sessionID string, msg protocol.UserInput) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendCalls = append(m.sendCalls, msg)
	return nil
}

func (m *mockAdapter) Detach(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.detachCalls = append(m.detachCalls, sessionID)
	if ch, ok := m.attachChans[sessionID]; ok {
		close(ch)
		delete(m.attachChans, sessionID)
	}
	return nil
}

func (m *mockAdapter) getAttachCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, len(m.attachCalls))
	copy(out, m.attachCalls)
	return out
}

func (m *mockAdapter) getDetachCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, len(m.detachCalls))
	copy(out, m.detachCalls)
	return out
}

func (m *mockAdapter) getSendCalls() []protocol.UserInput {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]protocol.UserInput, len(m.sendCalls))
	copy(out, m.sendCalls)
	return out
}

func (m *mockAdapter) sendEvent(sessionID string, ev protocol.Event) {
	m.mu.Lock()
	ch := m.attachChans[sessionID]
	m.mu.Unlock()
	if ch != nil {
		ch <- ev
	}
}

// --- mock bridge ---

type mockBridge struct {
	mu       sync.Mutex
	sent     []daemon.SentEvent
	inbound  chan protocol.Payload
	statusCalls []string
}

type SentEvent = daemon.SentEvent

func newMockBridge() *mockBridge {
	return &mockBridge{
		inbound: make(chan protocol.Payload, 8),
	}
}

func (b *mockBridge) SendEvent(_ context.Context, event protocol.Event, sessionID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.sent = append(b.sent, daemon.SentEvent{Event: event, SessionID: sessionID})
	return nil
}

func (b *mockBridge) ReceiveLoop(_ context.Context) <-chan protocol.Payload {
	return b.inbound
}

func (b *mockBridge) SendMachineStatus(_ context.Context, status string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.statusCalls = append(b.statusCalls, status)
	return nil
}

func (b *mockBridge) getSent() []daemon.SentEvent {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]daemon.SentEvent, len(b.sent))
	copy(out, b.sent)
	return out
}

func (b *mockBridge) getStatusCalls() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]string, len(b.statusCalls))
	copy(out, b.statusCalls)
	return out
}

// --- tests ---

func TestDaemonDiscoversAndAttaches(t *testing.T) {
	sessions := []protocol.Session{
		{ID: "sess-1", Tool: "claude-code", Project: "/home/user/project"},
	}
	adapter := newMockAdapter(sessions)
	bridge := newMockBridge()

	d := daemon.New(adapter, bridge, daemon.Options{PollInterval: 20 * time.Millisecond})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() { errCh <- d.Run(ctx) }()

	// Wait for attach to be called.
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if calls := adapter.getAttachCalls(); len(calls) > 0 {
			if calls[0] != "sess-1" {
				t.Fatalf("expected attach to sess-1, got %q", calls[0])
			}
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if len(adapter.getAttachCalls()) == 0 {
		t.Fatal("expected Attach to be called but it wasn't")
	}

	cancel()
	select {
	case <-errCh:
	case <-time.After(2 * time.Second):
		t.Fatal("daemon did not stop after cancel")
	}
}

func TestDaemonForwardsEventsToBridge(t *testing.T) {
	sessions := []protocol.Session{
		{ID: "sess-2", Tool: "claude-code", Project: "/home/user/project"},
	}
	adapter := newMockAdapter(sessions)
	bridge := newMockBridge()

	d := daemon.New(adapter, bridge, daemon.Options{PollInterval: 20 * time.Millisecond})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go d.Run(ctx) //nolint:errcheck

	// Wait for attach.
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if len(adapter.getAttachCalls()) > 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if len(adapter.getAttachCalls()) == 0 {
		t.Fatal("Attach not called")
	}

	// Send an event through the mock adapter.
	data, _ := json.Marshal(protocol.AssistantMessageData{Text: "hello"})
	ev := protocol.Event{Kind: protocol.PayloadAssistantMessage, Seq: 1, Data: json.RawMessage(data)}
	adapter.sendEvent("sess-2", ev)

	// Wait for bridge to receive it.
	deadline = time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if sent := bridge.getSent(); len(sent) > 0 {
			if sent[0].SessionID != "sess-2" {
				t.Fatalf("expected session sess-2, got %q", sent[0].SessionID)
			}
			if sent[0].Event.Kind != protocol.PayloadAssistantMessage {
				t.Fatalf("expected assistant message kind")
			}
			cancel()
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("event not forwarded to bridge")
}

func TestDaemonDispatchesIncomingToAdapter(t *testing.T) {
	sessions := []protocol.Session{
		{ID: "sess-3", Tool: "claude-code", Project: "/home/user/project"},
	}
	adapter := newMockAdapter(sessions)
	bridge := newMockBridge()

	d := daemon.New(adapter, bridge, daemon.Options{PollInterval: 20 * time.Millisecond})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go d.Run(ctx) //nolint:errcheck

	// Wait for attach.
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if len(adapter.getAttachCalls()) > 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	// Send a payload from the bridge inbound channel.
	msgData, _ := json.Marshal(protocol.UserMessageData{Text: "user says hi"})
	p := protocol.Payload{
		Kind: protocol.PayloadUserMessage,
		Seq:  10,
		Data: json.RawMessage(msgData),
	}
	bridge.inbound <- p

	// Wait for adapter.Send to be called.
	deadline = time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if calls := adapter.getSendCalls(); len(calls) > 0 {
			cancel()
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("incoming message not dispatched to adapter")
}

func TestDaemonSendsOnlineOnStart(t *testing.T) {
	adapter := newMockAdapter(nil)
	bridge := newMockBridge()

	d := daemon.New(adapter, bridge, daemon.Options{PollInterval: 50 * time.Millisecond})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go d.Run(ctx) //nolint:errcheck

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if calls := bridge.getStatusCalls(); len(calls) > 0 && calls[0] == "online" {
			cancel()
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("expected 'online' status to be sent on start")
}

func TestDaemonDetachesRemovedSessions(t *testing.T) {
	sessions := []protocol.Session{
		{ID: "sess-gone", Tool: "claude-code", Project: "/home/user/project"},
	}
	adapter := newMockAdapter(sessions)
	bridge := newMockBridge()

	d := daemon.New(adapter, bridge, daemon.Options{PollInterval: 20 * time.Millisecond})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go d.Run(ctx) //nolint:errcheck

	// Wait for initial attach.
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if len(adapter.getAttachCalls()) > 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if len(adapter.getAttachCalls()) == 0 {
		t.Fatal("Attach not called")
	}

	// Remove the session from discovery.
	adapter.mu.Lock()
	adapter.sessions = nil
	adapter.mu.Unlock()

	// Wait for detach.
	deadline = time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if calls := adapter.getDetachCalls(); len(calls) > 0 {
			if calls[0] != "sess-gone" {
				t.Fatalf("expected detach of sess-gone, got %q", calls[0])
			}
			cancel()
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("Detach not called for removed session")
}
