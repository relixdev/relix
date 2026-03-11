package hub_test

import (
	"context"
	"testing"
	"time"

	"github.com/relixdev/protocol"
	"github.com/relixdev/relix/relay/internal/hub"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func readEnvelope(t *testing.T, ws *websocket.Conn) protocol.Envelope {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var env protocol.Envelope
	if err := wsjson.Read(ctx, ws, &env); err != nil {
		t.Fatalf("readEnvelope: %v", err)
	}
	return env
}

func TestRouteAgentToMobiles(t *testing.T) {
	h := hub.New()

	agentConn, _ := newConnPair(t)
	mob1Conn, mob1Server := newConnPair(t)
	mob2Conn, mob2Server := newConnPair(t)

	h.RegisterAgent("user1", "machine1", agentConn)
	h.RegisterMobile("user1", mob1Conn)
	h.RegisterMobile("user1", mob2Conn)

	env := protocol.NewEnvelope(protocol.MsgSessionEvent, "machine1")
	env.SessionID = "sess-abc"

	if err := h.RouteEnvelope("user1", "agent", env); err != nil {
		t.Fatalf("RouteEnvelope: %v", err)
	}

	// Both mobile server sides should receive the envelope.
	for i, serverWS := range []*websocket.Conn{mob1Server, mob2Server} {
		got := readEnvelope(t, serverWS)
		if got.Type != protocol.MsgSessionEvent {
			t.Errorf("mobile%d: type: got %q want %q", i+1, got.Type, protocol.MsgSessionEvent)
		}
		if got.MachineID != "machine1" {
			t.Errorf("mobile%d: machine_id: got %q want machine1", i+1, got.MachineID)
		}
	}
}

func TestRouteMobileToAgent(t *testing.T) {
	h := hub.New()

	agentConn, agentServer := newConnPair(t)
	mobileConn, _ := newConnPair(t)

	h.RegisterAgent("user1", "machine1", agentConn)
	h.RegisterMobile("user1", mobileConn)

	env := protocol.NewEnvelope(protocol.MsgUserInput, "machine1")

	if err := h.RouteEnvelope("user1", "mobile", env); err != nil {
		t.Fatalf("RouteEnvelope: %v", err)
	}

	got := readEnvelope(t, agentServer)
	if got.Type != protocol.MsgUserInput {
		t.Errorf("type: got %q want %q", got.Type, protocol.MsgUserInput)
	}
}

func TestRouteMobileToUnknownMachine(t *testing.T) {
	h := hub.New()

	mobileConn, _ := newConnPair(t)
	h.RegisterMobile("user1", mobileConn)

	env := protocol.NewEnvelope(protocol.MsgUserInput, "nonexistent-machine")
	err := h.RouteEnvelope("user1", "mobile", env)
	if err == nil {
		t.Error("expected error routing to unknown machine, got nil")
	}
}

func TestRouteAgentNoMobilesBuffers(t *testing.T) {
	h := hub.New()

	agentConn, _ := newConnPair(t)
	h.RegisterAgent("user1", "machine1", agentConn)

	env := protocol.NewEnvelope(protocol.MsgSessionEvent, "machine1")
	// When no mobiles are connected, the message is buffered — no error returned.
	if err := h.RouteEnvelope("user1", "agent", env); err != nil {
		t.Errorf("expected nil error (buffered), got: %v", err)
	}
}
