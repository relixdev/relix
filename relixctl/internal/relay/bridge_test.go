package relay_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/relixdev/protocol"
	"github.com/relixdev/relix/relixctl/internal/relay"
)

func TestBridgeSendEvent(t *testing.T) {
	// Generate two key pairs: agent (bridge side) and mobile (peer side).
	agentPub, agentPriv, err := protocol.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	mobilePub, mobilePriv, err := protocol.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	recv := make(chan protocol.Envelope, 4)
	srv := newTestServer(t, recv, nil)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := relay.NewClient()
	if err := c.Connect(ctx, wsURL(srv), "tok", "m1"); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer c.Close()

	// Drain auth.
	select {
	case <-recv:
	case <-ctx.Done():
		t.Fatal("timeout waiting for auth")
	}

	b := relay.NewBridge(c, agentPub, agentPriv, mobilePub)

	data, _ := json.Marshal(protocol.AssistantMessageData{Text: "hello"})
	event := protocol.Event{
		Kind: protocol.PayloadAssistantMessage,
		Seq:  1,
		Data: json.RawMessage(data),
	}

	if err := b.SendEvent(ctx, event, "session-1"); err != nil {
		t.Fatalf("SendEvent: %v", err)
	}

	select {
	case env := <-recv:
		if env.Type != protocol.MsgSessionEvent {
			t.Fatalf("expected MsgSessionEvent, got %q", env.Type)
		}
		if env.SessionID != "session-1" {
			t.Fatalf("expected session_id session-1, got %q", env.SessionID)
		}
		// Payload should be encrypted — not plain JSON.
		if len(env.Payload) == 0 {
			t.Fatal("expected non-empty encrypted payload")
		}
		// Decrypt with mobile private key (mobile is recipient).
		var p protocol.Payload
		if err := protocol.OpenPayload(env.Payload, agentPub, mobilePriv, &p); err != nil {
			t.Fatalf("OpenPayload: %v", err)
		}
		if p.Kind != protocol.PayloadAssistantMessage {
			t.Fatalf("expected kind assistant_message, got %q", p.Kind)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for session event envelope")
	}
}

func TestBridgeReceiveLoop(t *testing.T) {
	agentPub, agentPriv, err := protocol.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	mobilePub, mobilePriv, err := protocol.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	// Build the encrypted payload that the "mobile" would send.
	innerPayload := protocol.Payload{
		Kind: protocol.PayloadUserMessage,
		Seq:  42,
	}
	innerPayload.Data, _ = json.Marshal(protocol.UserMessageData{Text: "hi agent"})
	ciphertext, err := protocol.SealPayload(innerPayload, agentPub, mobilePriv)
	if err != nil {
		t.Fatal(err)
	}

	inboundEnv := protocol.NewEnvelope(protocol.MsgUserInput, "mobile-machine")
	inboundEnv.SessionID = "session-2"
	inboundEnv.Payload = ciphertext

	// Server sends this envelope right after accepting.
	serverSend := make(chan protocol.Envelope, 1)
	serverSend <- inboundEnv
	close(serverSend)

	recv := make(chan protocol.Envelope, 4)
	srv := newTestServer(t, recv, serverSend)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := relay.NewClient()
	if err := c.Connect(ctx, wsURL(srv), "tok", "m2"); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer c.Close()

	b := relay.NewBridge(c, agentPub, agentPriv, mobilePub)
	payloadCh := b.ReceiveLoop(ctx)

	select {
	case p := <-payloadCh:
		if p.Kind != protocol.PayloadUserMessage {
			t.Fatalf("expected PayloadUserMessage, got %q", p.Kind)
		}
		if p.Seq != 42 {
			t.Fatalf("expected seq 42, got %d", p.Seq)
		}
		var msg protocol.UserMessageData
		if err := json.Unmarshal(p.Data, &msg); err != nil {
			t.Fatalf("unmarshal data: %v", err)
		}
		if msg.Text != "hi agent" {
			t.Fatalf("expected 'hi agent', got %q", msg.Text)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for decrypted payload")
	}

	// Unused but checked for compilation.
	_ = mobilePriv
	_ = mobilePub
}
