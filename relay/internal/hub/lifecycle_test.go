package hub_test

import (
	"context"
	"testing"
	"time"

	"github.com/relixdev/relix/relay/internal/hub"
	"nhooyr.io/websocket"
)

func TestStartPingerUnregistersAgentOnClose(t *testing.T) {
	h := hub.New()

	agentConn, agentServerWS := newConnPair(t)
	h.RegisterAgent("user1", "machine1", agentConn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		h.StartPinger(ctx, "user1", "agent", "machine1", agentConn)
	}()

	// Close the server side — pinger should detect the error and unregister.
	agentServerWS.Close(websocket.StatusNormalClosure, "test close")

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("StartPinger did not return after connection closed")
	}

	if got := h.GetAgent("user1", "machine1"); got != nil {
		t.Error("agent should be unregistered after pinger detects close")
	}
}

func TestStartPingerUnregistersMobileOnClose(t *testing.T) {
	h := hub.New()

	mobConn, mobServerWS := newConnPair(t)
	h.RegisterMobile("user1", mobConn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		h.StartPinger(ctx, "user1", "mobile", "", mobConn)
	}()

	mobServerWS.Close(websocket.StatusNormalClosure, "test close")

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("StartPinger did not return after mobile connection closed")
	}

	if got := h.GetMobiles("user1"); len(got) != 0 {
		t.Errorf("mobile should be unregistered after pinger detects close, got %d", len(got))
	}
}

func TestStartPingerRespectsContextCancel(t *testing.T) {
	h := hub.New()

	agentConn, _ := newConnPair(t)
	h.RegisterAgent("user1", "machine1", agentConn)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		defer close(done)
		h.StartPinger(ctx, "user1", "agent", "machine1", agentConn)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("StartPinger did not return after context cancel")
	}
}
