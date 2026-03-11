package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/relixdev/protocol"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// connectAndAuth dials the /ws endpoint, sends an auth envelope, and returns
// the websocket ready for subsequent reads/writes.
func connectAndAuth(t *testing.T, ts *httptest.Server, role, userID, machineID string) *websocket.Conn {
	t.Helper()
	ws := dialWS(t, ts)
	tok := makeToken(t, userID, role)
	sendAuth(t, ws, tok, machineID)
	return ws
}

// readEnv reads a single envelope with a 3-second deadline.
func readEnv(t *testing.T, ws *websocket.Conn) protocol.Envelope {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var env protocol.Envelope
	if err := wsjson.Read(ctx, ws, &env); err != nil {
		t.Fatalf("readEnv: %v", err)
	}
	return env
}

// writeEnv sends a single envelope with a 3-second deadline.
func writeEnv(t *testing.T, ws *websocket.Conn, env protocol.Envelope) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := wsjson.Write(ctx, ws, env); err != nil {
		t.Fatalf("writeEnv: %v", err)
	}
}

// TestE2E_AgentToMobile verifies that an envelope sent by the agent is
// delivered to a connected mobile client.
func TestE2E_AgentToMobile(t *testing.T) {
	ts := newTestServer(t)

	mobileWS := connectAndAuth(t, ts, "mobile", "user1", "")
	time.Sleep(30 * time.Millisecond)

	agentWS := connectAndAuth(t, ts, "agent", "user1", "machine1")

	// Mobile receives the machine-status "online" broadcast.
	statusEnv := readEnv(t, mobileWS)
	if statusEnv.Type != protocol.MsgMachineStatus {
		t.Fatalf("expected MsgMachineStatus, got %q", statusEnv.Type)
	}

	// Agent sends a session_event.
	event := protocol.NewEnvelope(protocol.MsgSessionEvent, "machine1")
	event.SessionID = "sess-abc"
	writeEnv(t, agentWS, event)

	// Mobile should receive it.
	got := readEnv(t, mobileWS)
	if got.Type != protocol.MsgSessionEvent {
		t.Errorf("type: got %q, want %q", got.Type, protocol.MsgSessionEvent)
	}
	if got.MachineID != "machine1" {
		t.Errorf("machine_id: got %q, want machine1", got.MachineID)
	}
}

// TestE2E_MobileToAgent verifies that an envelope sent by a mobile client is
// delivered to the correct agent.
func TestE2E_MobileToAgent(t *testing.T) {
	ts := newTestServer(t)

	agentWS := connectAndAuth(t, ts, "agent", "user1", "machine1")
	time.Sleep(30 * time.Millisecond)

	mobileWS := connectAndAuth(t, ts, "mobile", "user1", "")
	time.Sleep(30 * time.Millisecond)

	// Mobile sends user_input targeting machine1.
	input := protocol.Envelope{
		V:         protocol.ProtocolVersion,
		Type:      protocol.MsgUserInput,
		MachineID: "machine1",
		Timestamp: time.Now().Unix(),
	}
	writeEnv(t, mobileWS, input)

	// Agent should receive it.
	got := readEnv(t, agentWS)
	if got.Type != protocol.MsgUserInput {
		t.Errorf("type: got %q, want %q", got.Type, protocol.MsgUserInput)
	}
}

// TestE2E_BufferingAndDrain verifies that messages sent by an agent while no
// mobile is connected are buffered and delivered when a mobile reconnects.
func TestE2E_BufferingAndDrain(t *testing.T) {
	ts := newTestServer(t)

	// Connect agent with no mobile present — messages will be buffered.
	agentWS := connectAndAuth(t, ts, "agent", "user2", "machine2")
	time.Sleep(30 * time.Millisecond)

	// Agent sends an event — should be buffered since no mobile is connected.
	buffered := protocol.NewEnvelope(protocol.MsgSessionEvent, "machine2")
	buffered.SessionID = "sess-buf"
	writeEnv(t, agentWS, buffered)

	time.Sleep(30 * time.Millisecond)

	// Mobile connects — should immediately receive the buffered message.
	mobileWS := connectAndAuth(t, ts, "mobile", "user2", "")

	got := readEnv(t, mobileWS)
	if got.Type != protocol.MsgSessionEvent {
		t.Errorf("type: got %q, want %q", got.Type, protocol.MsgSessionEvent)
	}
	if got.SessionID != "sess-buf" {
		t.Errorf("session_id: got %q, want sess-buf", got.SessionID)
	}
}

// TestE2E_MachineStatusBroadcast verifies that machine online/offline status
// events are correctly broadcast to connected mobile clients.
func TestE2E_MachineStatusBroadcast(t *testing.T) {
	ts := newTestServer(t)

	// Mobile connects first so it receives subsequent broadcasts.
	mobileWS := connectAndAuth(t, ts, "mobile", "user3", "")
	time.Sleep(30 * time.Millisecond)

	// Agent connects → mobile should receive "online" broadcast.
	agentWS := connectAndAuth(t, ts, "agent", "user3", "machine3")

	online := readEnv(t, mobileWS)
	if online.Type != protocol.MsgMachineStatus {
		t.Fatalf("expected MsgMachineStatus (online), got %q", online.Type)
	}
	var onlineMsg protocol.MachineStatusMessage
	if err := json.Unmarshal(online.Payload, &onlineMsg); err != nil {
		t.Fatalf("unmarshal online payload: %v", err)
	}
	if onlineMsg.Status != protocol.StatusOnline {
		t.Errorf("online status: got %q, want %q", onlineMsg.Status, protocol.StatusOnline)
	}

	// Agent disconnects → mobile should receive "offline" broadcast.
	agentWS.Close(websocket.StatusNormalClosure, "")
	time.Sleep(50 * time.Millisecond)

	offline := readEnv(t, mobileWS)
	if offline.Type != protocol.MsgMachineStatus {
		t.Fatalf("expected MsgMachineStatus (offline), got %q", offline.Type)
	}
	var offlineMsg protocol.MachineStatusMessage
	if err := json.Unmarshal(offline.Payload, &offlineMsg); err != nil {
		t.Fatalf("unmarshal offline payload: %v", err)
	}
	if offlineMsg.Status != protocol.StatusOffline {
		t.Errorf("offline status: got %q, want %q", offlineMsg.Status, protocol.StatusOffline)
	}
}

// TestE2E_MetricsEndpoint verifies the /metrics HTTP endpoint is reachable
// and returns Prometheus exposition format content.
func TestE2E_MetricsEndpoint(t *testing.T) {
	ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want 200", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		t.Error("Content-Type header missing on /metrics response")
	}
}
