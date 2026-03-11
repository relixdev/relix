package hub_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/relixdev/relix/relay/internal/conn"
	"github.com/relixdev/relix/relay/internal/hub"
	"nhooyr.io/websocket"
)

// newConnPair creates an in-process WebSocket pair and returns a *conn.Conn
// for the client side. The server-side raw websocket is returned for cleanup.
func newConnPair(t *testing.T) (*conn.Conn, *websocket.Conn) {
	t.Helper()
	ready := make(chan *websocket.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			t.Errorf("accept: %v", err)
			return
		}
		ready <- ws
	}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	rawClient, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		cancel()
		t.Fatalf("dial: %v", err)
	}
	serverWS := <-ready

	// Close connections before cancelling context so the close handshake
	// does not time out. Cleanups run in LIFO order.
	t.Cleanup(cancel)
	t.Cleanup(func() { serverWS.CloseNow() })
	t.Cleanup(func() { rawClient.CloseNow() })

	return conn.New(rawClient), serverWS
}

func TestRegisterAndGetAgent(t *testing.T) {
	h := hub.New()
	c, _ := newConnPair(t)

	h.RegisterAgent("user1", "machine1", c)

	got := h.GetAgent("user1", "machine1")
	if got != c {
		t.Error("GetAgent: want registered conn, got nil or different conn")
	}
}

func TestGetAgentNotFound(t *testing.T) {
	h := hub.New()
	got := h.GetAgent("user1", "machine1")
	if got != nil {
		t.Error("GetAgent on empty hub: want nil")
	}
}

func TestUnregisterAgent(t *testing.T) {
	h := hub.New()
	c, _ := newConnPair(t)

	h.RegisterAgent("user1", "machine1", c)
	h.UnregisterAgent("user1", "machine1")

	got := h.GetAgent("user1", "machine1")
	if got != nil {
		t.Error("GetAgent after UnregisterAgent: want nil")
	}
}

func TestMultipleAgentsSameUser(t *testing.T) {
	h := hub.New()
	c1, _ := newConnPair(t)
	c2, _ := newConnPair(t)

	h.RegisterAgent("user1", "machine1", c1)
	h.RegisterAgent("user1", "machine2", c2)

	if h.GetAgent("user1", "machine1") != c1 {
		t.Error("machine1: want c1")
	}
	if h.GetAgent("user1", "machine2") != c2 {
		t.Error("machine2: want c2")
	}
}

func TestRegisterAndGetMobiles(t *testing.T) {
	h := hub.New()
	c1, _ := newConnPair(t)
	c2, _ := newConnPair(t)

	h.RegisterMobile("user1", c1)
	h.RegisterMobile("user1", c2)

	mobiles := h.GetMobiles("user1")
	if len(mobiles) != 2 {
		t.Fatalf("GetMobiles: want 2, got %d", len(mobiles))
	}
}

func TestGetMobilesNotFound(t *testing.T) {
	h := hub.New()
	mobiles := h.GetMobiles("user1")
	if len(mobiles) != 0 {
		t.Errorf("GetMobiles on empty hub: want 0, got %d", len(mobiles))
	}
}

func TestUnregisterMobile(t *testing.T) {
	h := hub.New()
	c1, _ := newConnPair(t)
	c2, _ := newConnPair(t)

	h.RegisterMobile("user1", c1)
	h.RegisterMobile("user1", c2)
	h.UnregisterMobile("user1", c1)

	mobiles := h.GetMobiles("user1")
	if len(mobiles) != 1 {
		t.Fatalf("after UnregisterMobile: want 1, got %d", len(mobiles))
	}
	if mobiles[0] != c2 {
		t.Error("remaining mobile should be c2")
	}
}

func TestUnregisterAllMobiles(t *testing.T) {
	h := hub.New()
	c, _ := newConnPair(t)

	h.RegisterMobile("user1", c)
	h.UnregisterMobile("user1", c)

	mobiles := h.GetMobiles("user1")
	if len(mobiles) != 0 {
		t.Errorf("want 0 mobiles, got %d", len(mobiles))
	}
}

func TestAgentsIsolatedByUser(t *testing.T) {
	h := hub.New()
	c1, _ := newConnPair(t)
	c2, _ := newConnPair(t)

	h.RegisterAgent("user1", "machine1", c1)
	h.RegisterAgent("user2", "machine1", c2)

	if h.GetAgent("user1", "machine1") != c1 {
		t.Error("user1/machine1 should be c1")
	}
	if h.GetAgent("user2", "machine1") != c2 {
		t.Error("user2/machine1 should be c2")
	}
}
