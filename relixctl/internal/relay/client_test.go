package relay_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/relixdev/protocol"
	"github.com/relixdev/relix/relixctl/internal/relay"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// newTestServer creates an httptest server that upgrades HTTP to WebSocket and
// sends received envelopes to serverRecv. It also sends any envelopes queued in
// serverSend to the client after accepting.
func newTestServer(t *testing.T, serverRecv chan<- protocol.Envelope, serverSend <-chan protocol.Envelope) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			t.Logf("server accept error: %v", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		ctx := r.Context()

		// Reader goroutine.
		go func() {
			for {
				var env protocol.Envelope
				if err := wsjson.Read(ctx, conn, &env); err != nil {
					return
				}
				select {
				case serverRecv <- env:
				case <-ctx.Done():
					return
				}
			}
		}()

		// Send queued messages to client.
		if serverSend != nil {
			for env := range serverSend {
				if err := wsjson.Write(ctx, conn, env); err != nil {
					return
				}
			}
		}

		// Keep alive until context done.
		<-ctx.Done()
	}))
	return srv
}

func wsURL(srv *httptest.Server) string {
	return "ws" + strings.TrimPrefix(srv.URL, "http")
}

func TestClientConnectSendsAuth(t *testing.T) {
	recv := make(chan protocol.Envelope, 4)
	srv := newTestServer(t, recv, nil)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := relay.NewClient()
	if err := c.Connect(ctx, wsURL(srv), "test-token", "machine-1"); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer c.Close()

	select {
	case env := <-recv:
		if env.Type != protocol.MsgAuth {
			t.Fatalf("expected MsgAuth, got %q", env.Type)
		}
		if env.MachineID != "machine-1" {
			t.Fatalf("expected machine_id machine-1, got %q", env.MachineID)
		}
		var auth protocol.AuthMessage
		if err := json.Unmarshal(env.Payload, &auth); err != nil {
			t.Fatalf("unmarshal auth payload: %v", err)
		}
		if auth.Token != "test-token" {
			t.Fatalf("expected token test-token, got %q", auth.Token)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for auth envelope")
	}
}

func TestClientSendReceive(t *testing.T) {
	recv := make(chan protocol.Envelope, 4)
	serverSend := make(chan protocol.Envelope, 1)

	srv := newTestServer(t, recv, serverSend)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := relay.NewClient()
	if err := c.Connect(ctx, wsURL(srv), "tok", "machine-2"); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer c.Close()

	// Drain the auth envelope.
	select {
	case <-recv:
	case <-ctx.Done():
		t.Fatal("timeout waiting for auth")
	}

	// Client sends an envelope.
	outEnv := protocol.NewEnvelope(protocol.MsgPing, "machine-2")
	if err := c.Send(ctx, outEnv); err != nil {
		t.Fatalf("Send: %v", err)
	}

	select {
	case got := <-recv:
		if got.Type != protocol.MsgPing {
			t.Fatalf("expected MsgPing, got %q", got.Type)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for sent envelope")
	}

	// Server sends an envelope to the client.
	inEnv := protocol.NewEnvelope(protocol.MsgPong, "server")
	serverSend <- inEnv
	close(serverSend)

	ch := c.ReadLoop(ctx)
	select {
	case got, ok := <-ch:
		if !ok {
			t.Fatal("ReadLoop channel closed unexpectedly")
		}
		if got.Type != protocol.MsgPong {
			t.Fatalf("expected MsgPong, got %q", got.Type)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for ReadLoop envelope")
	}
}
