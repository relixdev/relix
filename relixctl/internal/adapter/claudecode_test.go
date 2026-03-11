package adapter

import (
	"context"
	"encoding/json"
	"os/exec"
	"testing"
	"time"

	"github.com/relixdev/protocol"
)

// mockScript returns a CommandFactory that runs a shell script.
// The script receives the sessionID as $1.
func mockScriptFactory(script string) CommandFactory {
	return func(sessionID string) *exec.Cmd {
		return exec.Command("bash", "-c", script, "--", sessionID)
	}
}

func TestAttach_returnsEvents(t *testing.T) {
	// Script outputs 3 NDJSON lines then exits.
	script := `
echo '{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"hello"}]}}'
echo '{"type":"tool_use","id":"tu1","name":"bash","input":{"command":"ls"}}'
echo '{"type":"result","subtype":"success","result":"done"}'
`
	a := NewClaudeCodeAdapter(t.TempDir(), mockScriptFactory(script))
	ctx := context.Background()

	ch, err := a.Attach(ctx, "test-session")
	if err != nil {
		t.Fatalf("Attach failed: %v", err)
	}

	var events []protocol.Event
	timeout := time.After(5 * time.Second)
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				goto done
			}
			events = append(events, ev)
		case <-timeout:
			t.Fatal("timed out waiting for events")
		}
	}
done:

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	kinds := []protocol.PayloadKind{
		protocol.PayloadAssistantMessage,
		protocol.PayloadToolUse,
		protocol.PayloadToolResult,
	}
	for i, ev := range events {
		if ev.Kind != kinds[i] {
			t.Errorf("event[%d]: expected kind %q, got %q", i, kinds[i], ev.Kind)
		}
		if ev.Data == nil {
			t.Errorf("event[%d]: Data is nil", i)
		}
	}
}

func TestAttach_errorEvent(t *testing.T) {
	script := `echo '{"type":"error","error":{"type":"api_error","message":"something went wrong"}}'`
	a := NewClaudeCodeAdapter(t.TempDir(), mockScriptFactory(script))
	ctx := context.Background()

	ch, err := a.Attach(ctx, "test-session")
	if err != nil {
		t.Fatalf("Attach failed: %v", err)
	}

	timeout := time.After(5 * time.Second)
	var events []protocol.Event
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				goto done
			}
			events = append(events, ev)
		case <-timeout:
			t.Fatal("timed out")
		}
	}
done:

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Kind != protocol.PayloadError {
		t.Errorf("expected kind %q, got %q", protocol.PayloadError, events[0].Kind)
	}
}

func TestAttach_channelClosesOnExit(t *testing.T) {
	script := `echo '{"type":"result","subtype":"success","result":"ok"}'`
	a := NewClaudeCodeAdapter(t.TempDir(), mockScriptFactory(script))
	ctx := context.Background()

	ch, err := a.Attach(ctx, "test-session")
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

func TestSend_writesJSON(t *testing.T) {
	// Script: read one line from stdin, confirm it contains expected content,
	// then emit a result event. We use Python to safely handle JSON in the line.
	// Falls back to a simpler approach: read stdin line, write a fixed event.
	// We verify Send() returns nil (stdin writable) and the event is received.
	script := `
read -r line
# Emit a result event confirming we received stdin
printf '{"type":"result","subtype":"success","result":"received"}\n'
`
	a := NewClaudeCodeAdapter(t.TempDir(), mockScriptFactory(script))
	ctx := context.Background()

	ch, err := a.Attach(ctx, "test-session")
	if err != nil {
		t.Fatalf("Attach failed: %v", err)
	}

	msg := protocol.UserInput{
		Kind: protocol.InputMessage,
		Data: mustMarshal(t, protocol.UserMessageData{Text: "hello world"}),
	}
	if err := a.Send(ctx, "test-session", msg); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	timeout := time.After(5 * time.Second)
	var events []protocol.Event
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				goto done
			}
			events = append(events, ev)
		case <-timeout:
			t.Fatal("timed out waiting for event after Send")
		}
	}
done:
	if len(events) == 0 {
		t.Fatal("expected at least one event after Send")
	}
	if events[0].Kind != protocol.PayloadToolResult {
		t.Errorf("expected kind %q, got %q", protocol.PayloadToolResult, events[0].Kind)
	}
}

func TestSend_messageFormat(t *testing.T) {
	// Verify Send() formats InputMessage correctly by capturing what was written
	// to stdin. The script reads stdin and writes it as the result field using jq
	// if available, otherwise just confirms receipt.
	// We use a pipe-based approach: the script reads one line and writes it
	// verbatim as a JSON string using Python (available on macOS/Linux).
	script := `python3 -c "
import sys, json
line = sys.stdin.readline().rstrip('\n')
obj = json.loads(line)
content = obj.get('message', {}).get('content', '')
print(json.dumps({'type': 'result', 'subtype': 'success', 'result': content}))
"`
	a := NewClaudeCodeAdapter(t.TempDir(), mockScriptFactory(script))
	ctx := context.Background()

	ch, err := a.Attach(ctx, "test-session")
	if err != nil {
		t.Fatalf("Attach failed: %v", err)
	}

	msg := protocol.UserInput{
		Kind: protocol.InputMessage,
		Data: mustMarshal(t, protocol.UserMessageData{Text: "hello world"}),
	}
	if err := a.Send(ctx, "test-session", msg); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	timeout := time.After(5 * time.Second)
	var events []protocol.Event
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				goto done
			}
			events = append(events, ev)
		case <-timeout:
			t.Fatal("timed out waiting for event")
		}
	}
done:
	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}

	// Parse the result field to verify the content was "hello world".
	var resultEnvelope struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(events[0].Data, &resultEnvelope); err != nil {
		t.Fatalf("unmarshal result: %v (data: %s)", err, events[0].Data)
	}
	if resultEnvelope.Result != "hello world" {
		t.Errorf("expected result %q, got %q", "hello world", resultEnvelope.Result)
	}
}

func mustMarshal(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
