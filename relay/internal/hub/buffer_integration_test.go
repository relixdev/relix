package hub_test

import (
	"context"
	"testing"
	"time"

	"github.com/relixdev/protocol"
	"github.com/relixdev/relix/relay/internal/hub"
	"nhooyr.io/websocket/wsjson"
)

// TestBufferDeliveredOnMobileConnect verifies that envelopes sent by an agent
// while no mobile is connected are buffered and delivered when a mobile connects.
func TestBufferDeliveredOnMobileConnect(t *testing.T) {
	h := hub.New()

	agentConn, _ := newConnPair(t)
	h.RegisterAgent("user1", "machine1", agentConn)

	// Send 3 messages with no mobile connected — they should be buffered.
	for i := 0; i < 3; i++ {
		env := protocol.NewEnvelope(protocol.MsgSessionEvent, "machine1")
		env.SessionID = string(rune('A' + i))
		// RouteEnvelope returns error (no mobiles), but should buffer.
		_ = h.RouteEnvelope("user1", "agent", env)
	}

	// Now connect a mobile — it should receive the 3 buffered messages in order.
	mobileConn, mobileServer := newConnPair(t)
	h.RegisterMobile("user1", mobileConn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for i := 0; i < 3; i++ {
		var got protocol.Envelope
		if err := wsjson.Read(ctx, mobileServer, &got); err != nil {
			t.Fatalf("message %d: read failed: %v", i, err)
		}
		want := string(rune('A' + i))
		if got.SessionID != want {
			t.Errorf("message %d: want SessionID %q, got %q", i, want, got.SessionID)
		}
	}
}

// TestBufferNotUsedWhenMobileConnected verifies that when a mobile is already
// connected, messages are delivered directly without buffering.
func TestBufferNotUsedWhenMobileConnected(t *testing.T) {
	h := hub.New()

	agentConn, _ := newConnPair(t)
	mobileConn, mobileServer := newConnPair(t)

	h.RegisterAgent("user1", "machine1", agentConn)
	h.RegisterMobile("user1", mobileConn)

	env := protocol.NewEnvelope(protocol.MsgSessionEvent, "machine1")
	env.SessionID = "direct"

	if err := h.RouteEnvelope("user1", "agent", env); err != nil {
		t.Fatalf("RouteEnvelope: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var got protocol.Envelope
	if err := wsjson.Read(ctx, mobileServer, &got); err != nil {
		t.Fatalf("read: %v", err)
	}
	if got.SessionID != "direct" {
		t.Errorf("want SessionID %q, got %q", "direct", got.SessionID)
	}
}

// TestBufferOrderPreservedAcrossMultipleAgentMessages verifies insertion-order
// delivery of multiple buffered messages from different machines.
func TestBufferOrderPreservedAcrossMultipleAgentMessages(t *testing.T) {
	h := hub.New()

	agentConn, _ := newConnPair(t)
	h.RegisterAgent("user1", "machine1", agentConn)

	const n = 5
	for i := 0; i < n; i++ {
		env := protocol.NewEnvelope(protocol.MsgSessionEvent, "machine1")
		env.SessionID = string(rune('0' + i))
		_ = h.RouteEnvelope("user1", "agent", env)
	}

	mobileConn, mobileServer := newConnPair(t)
	h.RegisterMobile("user1", mobileConn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for i := 0; i < n; i++ {
		var got protocol.Envelope
		if err := wsjson.Read(ctx, mobileServer, &got); err != nil {
			t.Fatalf("message %d: %v", i, err)
		}
		want := string(rune('0' + i))
		if got.SessionID != want {
			t.Errorf("index %d: want %q, got %q", i, want, got.SessionID)
		}
	}
}
