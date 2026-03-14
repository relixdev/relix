package adapter

import (
	"context"
	"encoding/json"
	"os/exec"
	"testing"
	"time"

	"github.com/relixdev/protocol"
)

// mockAiderFactory returns a CommandFactory that runs a shell script.
func mockAiderFactory(script string) AiderCommandFactory {
	return func(sessionID string) *exec.Cmd {
		return exec.Command("bash", "-c", script, "--", sessionID)
	}
}

func mockScanner(procs []aiderProcInfo) AiderProcessScanner {
	return func(ctx context.Context) ([]aiderProcInfo, error) {
		return procs, nil
	}
}

func collectEvents(t *testing.T, ch <-chan protocol.Event) []protocol.Event {
	t.Helper()
	var events []protocol.Event
	timeout := time.After(5 * time.Second)
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				return events
			}
			events = append(events, ev)
		case <-timeout:
			t.Fatal("timed out waiting for events")
		}
	}
}

func TestAiderDiscover(t *testing.T) {
	scanner := mockScanner([]aiderProcInfo{
		{PID: "1234", CWD: "/home/user/project"},
		{PID: "5678", CWD: "/tmp/other"},
	})
	a := NewAiderAdapter(nil, scanner)

	sessions, err := a.Discover(context.Background())
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	if sessions[0].Tool != "aider" {
		t.Errorf("expected tool %q, got %q", "aider", sessions[0].Tool)
	}
	if sessions[0].ID != "aider-1234" {
		t.Errorf("expected ID %q, got %q", "aider-1234", sessions[0].ID)
	}
}

func TestAiderDiscover_empty(t *testing.T) {
	scanner := mockScanner(nil)
	a := NewAiderAdapter(nil, scanner)

	sessions, err := a.Discover(context.Background())
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestAiderAttach_returnsEvents(t *testing.T) {
	script := `
echo "Aider v0.50.0"
echo "Applied edit to main.go"
echo "Here is the updated code"
echo "Error: could not find file"
`
	a := NewAiderAdapter(mockAiderFactory(script), mockScanner(nil))
	ch, err := a.Attach(context.Background(), "test-session")
	if err != nil {
		t.Fatalf("Attach failed: %v", err)
	}

	events := collectEvents(t, ch)

	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}

	expectedKinds := []protocol.PayloadKind{
		protocol.PayloadAssistantMessage, // "Aider v0.50.0"
		protocol.PayloadToolUse,          // "Applied edit to main.go"
		protocol.PayloadAssistantMessage, // "Here is the updated code"
		protocol.PayloadError,            // "Error: could not find file"
	}
	for i, ev := range events {
		if ev.Kind != expectedKinds[i] {
			t.Errorf("event[%d]: expected kind %q, got %q", i, expectedKinds[i], ev.Kind)
		}
		if ev.Data == nil {
			t.Errorf("event[%d]: Data is nil", i)
		}
		if ev.Seq == 0 {
			t.Errorf("event[%d]: Seq should be > 0", i)
		}
	}
}

func TestAiderAttach_toolCallPatterns(t *testing.T) {
	script := `
echo "Run shell command? ls -la"
echo "Running pytest tests/test_main.py"
echo "<<<<<<< SEARCH"
echo ">>>>>>> REPLACE"
`
	a := NewAiderAdapter(mockAiderFactory(script), mockScanner(nil))
	ch, err := a.Attach(context.Background(), "test-session")
	if err != nil {
		t.Fatalf("Attach failed: %v", err)
	}

	events := collectEvents(t, ch)
	for i, ev := range events {
		if ev.Kind != protocol.PayloadToolUse {
			t.Errorf("event[%d]: expected kind %q, got %q", i, protocol.PayloadToolUse, ev.Kind)
		}
	}
}

func TestAiderAttach_channelClosesOnExit(t *testing.T) {
	script := `echo "done"`
	a := NewAiderAdapter(mockAiderFactory(script), mockScanner(nil))
	ch, err := a.Attach(context.Background(), "test-session")
	if err != nil {
		t.Fatalf("Attach failed: %v", err)
	}

	timeout := time.After(5 * time.Second)
	var closed bool
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				closed = true
				goto done
			}
		case <-timeout:
			t.Fatal("timed out waiting for channel close")
		}
	}
done:
	if !closed {
		t.Error("channel was not closed after process exit")
	}
}

func TestAiderAttach_skipsEmptyLines(t *testing.T) {
	script := `
echo ""
echo "hello"
echo ""
echo ">"
echo "world"
`
	a := NewAiderAdapter(mockAiderFactory(script), mockScanner(nil))
	ch, err := a.Attach(context.Background(), "test-session")
	if err != nil {
		t.Fatalf("Attach failed: %v", err)
	}

	events := collectEvents(t, ch)
	if len(events) != 2 {
		t.Fatalf("expected 2 events (skipping empty and prompt), got %d", len(events))
	}
}

func TestAiderSend_message(t *testing.T) {
	// Script reads a line from stdin and echoes it back as output.
	script := `
read -r line
echo "$line"
`
	a := NewAiderAdapter(mockAiderFactory(script), mockScanner(nil))
	ch, err := a.Attach(context.Background(), "test-session")
	if err != nil {
		t.Fatalf("Attach failed: %v", err)
	}

	msg := protocol.UserInput{
		Kind: protocol.InputMessage,
		Data: mustMarshalAider(t, protocol.UserMessageData{Text: "fix the bug"}),
	}
	if err := a.Send(context.Background(), "test-session", msg); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	events := collectEvents(t, ch)
	if len(events) == 0 {
		t.Fatal("expected at least one event after Send")
	}

	// The echoed line should be "fix the bug" → assistant message.
	if events[0].Kind != protocol.PayloadAssistantMessage {
		t.Errorf("expected kind %q, got %q", protocol.PayloadAssistantMessage, events[0].Kind)
	}

	var data protocol.AssistantMessageData
	if err := json.Unmarshal(events[0].Data, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if data.Text != "fix the bug" {
		t.Errorf("expected text %q, got %q", "fix the bug", data.Text)
	}
}

func TestAiderSend_approval(t *testing.T) {
	script := `
read -r line
echo "$line"
`
	a := NewAiderAdapter(mockAiderFactory(script), mockScanner(nil))
	ch, err := a.Attach(context.Background(), "test-session")
	if err != nil {
		t.Fatalf("Attach failed: %v", err)
	}

	msg := protocol.UserInput{
		Kind: protocol.InputApproval,
		Data: mustMarshalAider(t, protocol.ApprovalResponseData{Approved: true}),
	}
	if err := a.Send(context.Background(), "test-session", msg); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	events := collectEvents(t, ch)
	if len(events) == 0 {
		t.Fatal("expected at least one event after Send")
	}

	// Approved → "y" echoed back.
	var data protocol.AssistantMessageData
	if err := json.Unmarshal(events[0].Data, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if data.Text != "y" {
		t.Errorf("expected text %q, got %q", "y", data.Text)
	}
}

func TestAiderSend_notAttached(t *testing.T) {
	a := NewAiderAdapter(nil, mockScanner(nil))
	err := a.Send(context.Background(), "nonexistent", protocol.UserInput{
		Kind: protocol.InputMessage,
		Data: mustMarshalAider(t, protocol.UserMessageData{Text: "hello"}),
	})
	if err == nil {
		t.Fatal("expected error for non-attached session")
	}
}

func TestAiderDetach(t *testing.T) {
	// Long-running script that we'll detach from.
	script := `while true; do sleep 0.1; echo "tick"; done`
	a := NewAiderAdapter(mockAiderFactory(script), mockScanner(nil))

	ch, err := a.Attach(context.Background(), "test-session")
	if err != nil {
		t.Fatalf("Attach failed: %v", err)
	}

	// Wait for at least one event to confirm it's running.
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for first event")
	}

	if err := a.Detach("test-session"); err != nil {
		t.Fatalf("Detach failed: %v", err)
	}

	// Channel should close after detach.
	timeout := time.After(5 * time.Second)
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return // success
			}
		case <-timeout:
			t.Fatal("timed out waiting for channel close after Detach")
		}
	}
}

func TestAiderDetach_notAttached(t *testing.T) {
	a := NewAiderAdapter(nil, mockScanner(nil))
	// Detaching a non-attached session should not error.
	if err := a.Detach("nonexistent"); err != nil {
		t.Fatalf("Detach should not error for non-attached session: %v", err)
	}
}

func TestParseAiderLine(t *testing.T) {
	tests := []struct {
		line     string
		wantKind protocol.PayloadKind
	}{
		{"", ""},
		{">", ""},
		{"Hello world", protocol.PayloadAssistantMessage},
		{"Error: file not found", protocol.PayloadError},
		{"error: something broke", protocol.PayloadError},
		{"Applied edit to main.go", protocol.PayloadToolUse},
		{"<<<<<<< SEARCH", protocol.PayloadToolUse},
		{">>>>>>> REPLACE", protocol.PayloadToolUse},
		{"Run shell command? ls", protocol.PayloadToolUse},
		{"Running pytest", protocol.PayloadToolUse},
	}

	for _, tt := range tests {
		kind, data := parseAiderLine(tt.line)
		if kind != tt.wantKind {
			t.Errorf("parseAiderLine(%q): got kind %q, want %q", tt.line, kind, tt.wantKind)
		}
		if tt.wantKind != "" && data == nil {
			t.Errorf("parseAiderLine(%q): data should not be nil for kind %q", tt.line, tt.wantKind)
		}
	}
}

func mustMarshalAider(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
