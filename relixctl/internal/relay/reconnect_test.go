package relay_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/relixdev/protocol"
	"github.com/relixdev/relix/relixctl/internal/relay"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// droppingServer accepts one connection, consumes the auth envelope, then
// closes the connection to simulate a drop. Subsequent connections are kept
// alive until the test ends. connectCount is incremented for each accepted conn.
func droppingServer(t *testing.T, connectCount *atomic.Int32, recv chan<- protocol.Envelope) *httptest.Server {
	t.Helper()
	first := atomic.Bool{}
	first.Store(true)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			return
		}
		connectCount.Add(1)
		ctx := r.Context()

		if first.Swap(false) {
			// First connection: drain auth then close abruptly.
			var env protocol.Envelope
			_ = wsjson.Read(ctx, conn, &env)
			conn.Close(websocket.StatusGoingAway, "drop")
			return
		}

		// Subsequent connections: drain the auth envelope first, then forward rest.
		go func() {
			first := true
			for {
				var env protocol.Envelope
				if err := wsjson.Read(ctx, conn, &env); err != nil {
					return
				}
				if first {
					first = false
					continue // skip the auth envelope
				}
				select {
				case recv <- env:
				case <-ctx.Done():
					return
				}
			}
		}()
		<-ctx.Done()
		conn.Close(websocket.StatusNormalClosure, "")
	}))
	return srv
}

func TestReconnectingClientReconnects(t *testing.T) {
	var connectCount atomic.Int32
	recv := make(chan protocol.Envelope, 8)

	srv := droppingServer(t, &connectCount, recv)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	onConnectCalled := make(chan struct{}, 4)

	rc := relay.NewReconnectingClient(relay.ReconnectOptions{
		RelayURL:  wsURL(srv),
		AuthToken: "tok",
		MachineID: "machine-rc",
		OnConnect: func() {
			onConnectCalled <- struct{}{}
		},
	})

	rc.Start(ctx)
	defer rc.Stop()

	// Wait for first OnConnect (initial connection).
	select {
	case <-onConnectCalled:
	case <-ctx.Done():
		t.Fatal("timeout waiting for first OnConnect")
	}

	// Wait for reconnect OnConnect (after server drops).
	select {
	case <-onConnectCalled:
	case <-ctx.Done():
		t.Fatal("timeout waiting for second OnConnect after reconnect")
	}

	if connectCount.Load() < 2 {
		t.Fatalf("expected at least 2 connections, got %d", connectCount.Load())
	}
}

func TestReconnectingClientQueuesMessages(t *testing.T) {
	var connectCount atomic.Int32
	recv := make(chan protocol.Envelope, 8)

	srv := droppingServer(t, &connectCount, recv)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reconnected := make(chan struct{}, 2)

	rc := relay.NewReconnectingClient(relay.ReconnectOptions{
		RelayURL:  wsURL(srv),
		AuthToken: "tok",
		MachineID: "machine-q",
		OnConnect: func() {
			reconnected <- struct{}{}
		},
	})

	rc.Start(ctx)
	defer rc.Stop()

	// Wait for initial connection.
	select {
	case <-reconnected:
	case <-ctx.Done():
		t.Fatal("timeout for initial connect")
	}

	// Queue a message (may be during disconnect).
	env := protocol.NewEnvelope(protocol.MsgPing, "machine-q")
	if err := rc.Send(ctx, env); err != nil {
		t.Logf("Send returned error (expected during reconnect): %v", err)
	}

	// Wait for reconnect.
	select {
	case <-reconnected:
	case <-ctx.Done():
		t.Fatal("timeout waiting for reconnect")
	}

	// After reconnect, send another envelope that should definitely be delivered.
	env2 := protocol.NewEnvelope(protocol.MsgPing, "machine-q")
	if err := rc.Send(ctx, env2); err != nil {
		t.Fatalf("Send after reconnect: %v", err)
	}

	// Verify at least one message arrives at the server (the post-reconnect send).
	select {
	case got := <-recv:
		if got.Type != protocol.MsgPing {
			t.Fatalf("expected MsgPing, got %q", got.Type)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for delivered message")
	}
}

func TestReconnectingClientOnDisconnect(t *testing.T) {
	var connectCount atomic.Int32
	recv := make(chan protocol.Envelope, 4)

	srv := droppingServer(t, &connectCount, recv)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	disconnected := make(chan struct{}, 2)
	connected := make(chan struct{}, 2)

	rc := relay.NewReconnectingClient(relay.ReconnectOptions{
		RelayURL:  wsURL(srv),
		AuthToken: "tok",
		MachineID: "machine-dc",
		OnConnect: func() {
			connected <- struct{}{}
		},
		OnDisconnect: func() {
			disconnected <- struct{}{}
		},
	})

	rc.Start(ctx)
	defer rc.Stop()

	// Wait for initial connect.
	select {
	case <-connected:
	case <-ctx.Done():
		t.Fatal("timeout waiting for connect")
	}

	// Wait for disconnect callback.
	select {
	case <-disconnected:
	case <-ctx.Done():
		t.Fatal("timeout waiting for OnDisconnect")
	}
}
