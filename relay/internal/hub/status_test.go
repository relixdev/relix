package hub_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/relixdev/protocol"
	"github.com/relixdev/relix/relay/internal/hub"
	"nhooyr.io/websocket/wsjson"
)

// newMobilePair registers a mobile conn in h and returns the server-side ws
// for reading messages that the hub sends to that mobile.
func newMobilePair(t *testing.T, h *hub.Hub, userID string) {
	t.Helper()
	mobClientConn, _ := newConnPair(t)
	h.RegisterMobile(userID, mobClientConn)
}

func TestRegisterAgentBroadcastsOnlineStatus(t *testing.T) {
	h := hub.New()

	// Register mobile: hub will write to mobClientConn; we read from mobServerWS.
	mobClientConn, mobServerWS := newConnPair(t)
	h.RegisterMobile("user1", mobClientConn)

	agentConn, _ := newConnPair(t)
	h.RegisterAgent("user1", "machine1", agentConn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var env protocol.Envelope
	if err := wsjson.Read(ctx, mobServerWS, &env); err != nil {
		t.Fatalf("read status envelope: %v", err)
	}

	if env.Type != protocol.MsgMachineStatus {
		t.Errorf("type: got %q want %q", env.Type, protocol.MsgMachineStatus)
	}

	var status protocol.MachineStatusMessage
	if err := json.Unmarshal(env.Payload, &status); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if status.MachineID != "machine1" {
		t.Errorf("machine_id: got %q want machine1", status.MachineID)
	}
	if status.Status != protocol.StatusOnline {
		t.Errorf("status: got %q want %q", status.Status, protocol.StatusOnline)
	}
}

func TestUnregisterAgentBroadcastsOfflineStatus(t *testing.T) {
	h := hub.New()

	mobClientConn, mobServerWS := newConnPair(t)
	h.RegisterMobile("user1", mobClientConn)

	agentConn, _ := newConnPair(t)
	h.RegisterAgent("user1", "machine1", agentConn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Drain the online broadcast first.
	var online protocol.Envelope
	if err := wsjson.Read(ctx, mobServerWS, &online); err != nil {
		t.Fatalf("drain online broadcast: %v", err)
	}

	// Unregister agent — should broadcast offline status.
	h.UnregisterAgent("user1", "machine1")

	var env protocol.Envelope
	if err := wsjson.Read(ctx, mobServerWS, &env); err != nil {
		t.Fatalf("read offline status envelope: %v", err)
	}

	if env.Type != protocol.MsgMachineStatus {
		t.Errorf("type: got %q want %q", env.Type, protocol.MsgMachineStatus)
	}

	var status protocol.MachineStatusMessage
	if err := json.Unmarshal(env.Payload, &status); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if status.MachineID != "machine1" {
		t.Errorf("machine_id: got %q want machine1", status.MachineID)
	}
	if status.Status != protocol.StatusOffline {
		t.Errorf("status: got %q want %q", status.Status, protocol.StatusOffline)
	}
}
