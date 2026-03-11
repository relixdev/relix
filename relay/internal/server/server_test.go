package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/relixdev/protocol"
	"github.com/relixdev/relix/relay/internal/hub"
	"github.com/relixdev/relix/relay/internal/server"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const testSecret = "test-secret-key"

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	h := hub.New()
	srv := server.New(testSecret, h)
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)
	return ts
}

func makeToken(t *testing.T, userID, role string) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	signed, err := tok.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func dialWS(t *testing.T, ts *httptest.Server) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	ws, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial ws: %v", err)
	}
	t.Cleanup(func() { ws.Close(websocket.StatusNormalClosure, "") })
	return ws
}

func sendAuth(t *testing.T, ws *websocket.Conn, token, machineID string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	payload, _ := json.Marshal(protocol.AuthMessage{Token: token})
	env := protocol.Envelope{
		V:         protocol.ProtocolVersion,
		Type:      protocol.MsgAuth,
		MachineID: machineID,
		Timestamp: time.Now().Unix(),
		Payload:   payload,
	}
	if err := wsjson.Write(ctx, ws, env); err != nil {
		t.Fatalf("sendAuth: %v", err)
	}
}

// --- Existing tests (updated for new New() signature) ---

func TestHealthEndpoint(t *testing.T) {
	h := hub.New()
	srv := server.New(testSecret, h)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("want status=ok, got %q", body["status"])
	}
}

func TestWSEndpointUpgrades(t *testing.T) {
	ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/ws")
	if err != nil {
		t.Fatalf("GET /ws: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		t.Errorf("/ws route not registered (got 404)")
	}
}

// --- Integration test: agent sends event, mobile receives it ---

func TestAgentToMobileRouting(t *testing.T) {
	ts := newTestServer(t)

	agentToken := makeToken(t, "user1", "agent")
	mobileToken := makeToken(t, "user1", "mobile")

	// Connect mobile first so it's registered before agent sends.
	mobileWS := dialWS(t, ts)
	sendAuth(t, mobileWS, mobileToken, "")

	// Give the server a moment to register the mobile.
	time.Sleep(50 * time.Millisecond)

	// Connect agent.
	agentWS := dialWS(t, ts)
	sendAuth(t, agentWS, agentToken, "machine1")

	// Drain the machine-status "online" broadcast the mobile receives.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var statusEnv protocol.Envelope
	if err := wsjson.Read(ctx, mobileWS, &statusEnv); err != nil {
		t.Fatalf("read status envelope: %v", err)
	}
	if statusEnv.Type != protocol.MsgMachineStatus {
		t.Errorf("expected MsgMachineStatus broadcast, got %q", statusEnv.Type)
	}

	// Agent sends a session event.
	event := protocol.Envelope{
		V:         protocol.ProtocolVersion,
		Type:      protocol.MsgSessionEvent,
		MachineID: "machine1",
		SessionID: "sess-1",
		Timestamp: time.Now().Unix(),
	}
	if err := wsjson.Write(ctx, agentWS, event); err != nil {
		t.Fatalf("agent write: %v", err)
	}

	// Mobile should receive it.
	var got protocol.Envelope
	if err := wsjson.Read(ctx, mobileWS, &got); err != nil {
		t.Fatalf("mobile read: %v", err)
	}
	if got.Type != protocol.MsgSessionEvent {
		t.Errorf("type: got %q want %q", got.Type, protocol.MsgSessionEvent)
	}
	if got.MachineID != "machine1" {
		t.Errorf("machine_id: got %q want machine1", got.MachineID)
	}
}

func TestMobileToAgentRouting(t *testing.T) {
	ts := newTestServer(t)

	agentToken := makeToken(t, "user1", "agent")
	mobileToken := makeToken(t, "user1", "mobile")

	// Connect agent first.
	agentWS := dialWS(t, ts)
	sendAuth(t, agentWS, agentToken, "machine1")

	time.Sleep(50 * time.Millisecond)

	// Connect mobile.
	mobileWS := dialWS(t, ts)
	sendAuth(t, mobileWS, mobileToken, "")

	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Mobile sends user input targeting machine1.
	input := protocol.Envelope{
		V:         protocol.ProtocolVersion,
		Type:      protocol.MsgUserInput,
		MachineID: "machine1",
		Timestamp: time.Now().Unix(),
	}
	if err := wsjson.Write(ctx, mobileWS, input); err != nil {
		t.Fatalf("mobile write: %v", err)
	}

	// Agent should receive it.
	var got protocol.Envelope
	if err := wsjson.Read(ctx, agentWS, &got); err != nil {
		t.Fatalf("agent read: %v", err)
	}
	if got.Type != protocol.MsgUserInput {
		t.Errorf("type: got %q want %q", got.Type, protocol.MsgUserInput)
	}
}

func TestWSRejectsInvalidToken(t *testing.T) {
	ts := newTestServer(t)

	ws := dialWS(t, ts)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	payload, _ := json.Marshal(protocol.AuthMessage{Token: "not-a-valid-token"})
	env := protocol.Envelope{
		V:         protocol.ProtocolVersion,
		Type:      protocol.MsgAuth,
		MachineID: "machine1",
		Timestamp: time.Now().Unix(),
		Payload:   payload,
	}
	if err := wsjson.Write(ctx, ws, env); err != nil {
		t.Fatalf("write auth: %v", err)
	}

	// Server should close the connection.
	var discard protocol.Envelope
	err := wsjson.Read(ctx, ws, &discard)
	if err == nil {
		t.Error("expected connection to be closed after invalid token, but read succeeded")
	}
}
