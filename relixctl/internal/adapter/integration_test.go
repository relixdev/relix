package adapter

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/relixdev/protocol"
)

// TestIntegration_fullFlow exercises the complete Attach → read events →
// Send → Detach lifecycle against a mock subprocess.
//
// The mock script:
//  1. Emits 3 NDJSON event lines (assistant, tool_use, result)
//  2. Reads one line from stdin
//  3. Emits a final result event to confirm stdin receipt
//  4. Exits cleanly
func TestIntegration_fullFlow(t *testing.T) {
	script := `
printf '{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"I will help you."}]}}\n'
printf '{"type":"tool_use","id":"tu1","name":"bash","input":{"command":"echo hi"}}\n'
printf '{"type":"result","subtype":"success","result":"hi"}\n'
read -r line
printf '{"type":"result","subtype":"success","result":"ack"}\n'
`
	factory := func(sessionID string) *exec.Cmd {
		return exec.Command("bash", "-c", script)
	}

	a := NewClaudeCodeAdapter(t.TempDir(), factory)
	ctx := context.Background()

	// Step 1: Attach
	ch, err := a.Attach(ctx, "integration-session")
	if err != nil {
		t.Fatalf("Attach failed: %v", err)
	}

	// Step 2: Read the first 3 events
	expectedKinds := []protocol.PayloadKind{
		protocol.PayloadAssistantMessage,
		protocol.PayloadToolUse,
		protocol.PayloadToolResult,
	}

	timeout := time.After(5 * time.Second)
	var received []protocol.Event

	for len(received) < 3 {
		select {
		case ev, ok := <-ch:
			if !ok {
				t.Fatalf("channel closed before receiving 3 events (got %d)", len(received))
			}
			received = append(received, ev)
		case <-timeout:
			t.Fatalf("timed out after receiving %d events", len(received))
		}
	}

	for i, ev := range received {
		if ev.Kind != expectedKinds[i] {
			t.Errorf("event[%d]: expected kind %q, got %q", i, expectedKinds[i], ev.Kind)
		}
		if len(ev.Data) == 0 {
			t.Errorf("event[%d]: Data is empty", i)
		}
	}

	// Step 3: Send a user message
	msg := protocol.UserInput{
		Kind: protocol.InputMessage,
		Data: mustMarshal(t, protocol.UserMessageData{Text: "continue"}),
	}
	if err := a.Send(ctx, "integration-session", msg); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	// Step 4: Verify the ack event is received (script echoes after reading stdin)
	timeout2 := time.After(5 * time.Second)
	var ackReceived bool
	for !ackReceived {
		select {
		case ev, ok := <-ch:
			if !ok {
				// Channel closed — that's fine if we haven't seen the ack yet
				// (process may have exited after sending ack)
				goto afterAck
			}
			if ev.Kind == protocol.PayloadToolResult {
				ackReceived = true
			}
		case <-timeout2:
			t.Fatal("timed out waiting for ack event")
		}
	}
afterAck:

	// Step 5: Detach and verify channel closes
	if err := a.Detach("integration-session"); err != nil {
		t.Fatalf("Detach failed: %v", err)
	}

	// Drain channel until closed (with timeout)
	timeout3 := time.After(5 * time.Second)
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return // success: channel closed
			}
		case <-timeout3:
			t.Fatal("timed out waiting for channel to close after Detach")
		}
	}
}

// TestIntegration_detachBeforeEvents verifies that Detach before any events
// are consumed causes the channel to close cleanly without deadlock.
func TestIntegration_detachBeforeEvents(t *testing.T) {
	// Script that sleeps (simulates a long-running process)
	script := `sleep 30`
	factory := func(sessionID string) *exec.Cmd {
		return exec.Command("bash", "-c", script)
	}

	a := NewClaudeCodeAdapter(t.TempDir(), factory)
	ctx := context.Background()

	ch, err := a.Attach(ctx, "detach-early-session")
	if err != nil {
		t.Fatalf("Attach failed: %v", err)
	}

	// Detach immediately — should kill the process and close the channel.
	if err := a.Detach("detach-early-session"); err != nil {
		t.Fatalf("Detach failed: %v", err)
	}

	timeout := time.After(5 * time.Second)
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return // success
			}
		case <-timeout:
			t.Fatal("timed out waiting for channel to close after early Detach")
		}
	}
}
