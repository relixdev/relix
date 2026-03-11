package conn_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/relixdev/protocol"
	"github.com/relixdev/relix/relay/internal/conn"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func newTestPair(t *testing.T) (client *conn.Conn, serverConn *websocket.Conn, cleanup func()) {
	t.Helper()

	ready := make(chan *websocket.Conn, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			t.Errorf("server accept: %v", err)
			return
		}
		ready <- ws
	}))

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	rawClient, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		cancel()
		srv.Close()
		t.Fatalf("dial: %v", err)
	}

	serverWS := <-ready
	c := conn.New(rawClient)

	return c, serverWS, func() {
		cancel()
		_ = serverWS.Close(websocket.StatusNormalClosure, "")
		srv.Close()
	}
}

func TestReadEnvelope(t *testing.T) {
	client, serverWS, cleanup := newTestPair(t)
	defer cleanup()

	want := protocol.Envelope{
		V:         1,
		Type:      protocol.MsgPing,
		MachineID: "machine-1",
		Timestamp: 12345,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Server sends; client reads.
	if err := wsjson.Write(ctx, serverWS, want); err != nil {
		t.Fatalf("server write: %v", err)
	}

	got, err := client.ReadEnvelope(ctx)
	if err != nil {
		t.Fatalf("ReadEnvelope: %v", err)
	}

	if got.Type != want.Type {
		t.Errorf("Type: got %q, want %q", got.Type, want.Type)
	}
	if got.MachineID != want.MachineID {
		t.Errorf("MachineID: got %q, want %q", got.MachineID, want.MachineID)
	}
}

func TestWriteEnvelope(t *testing.T) {
	client, serverWS, cleanup := newTestPair(t)
	defer cleanup()

	want := protocol.Envelope{
		V:         1,
		Type:      protocol.MsgPong,
		MachineID: "machine-2",
		Timestamp: 99999,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Client sends; server reads.
	if err := client.WriteEnvelope(ctx, want); err != nil {
		t.Fatalf("WriteEnvelope: %v", err)
	}

	var got protocol.Envelope
	if err := wsjson.Read(ctx, serverWS, &got); err != nil {
		t.Fatalf("server read: %v", err)
	}

	if got.Type != want.Type {
		t.Errorf("Type: got %q, want %q", got.Type, want.Type)
	}
	if got.MachineID != want.MachineID {
		t.Errorf("MachineID: got %q, want %q", got.MachineID, want.MachineID)
	}
}

func TestClose(t *testing.T) {
	client, _, cleanup := newTestPair(t)
	defer cleanup()

	if err := client.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}
