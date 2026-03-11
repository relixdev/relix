package conn

import (
	"context"

	"github.com/relixdev/protocol"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// Conn wraps a nhooyr.io/websocket.Conn with typed Envelope read/write methods.
type Conn struct {
	ws *websocket.Conn
}

// New wraps an existing websocket connection.
func New(ws *websocket.Conn) *Conn {
	return &Conn{ws: ws}
}

// ReadEnvelope reads a single JSON message and unmarshals it into a protocol.Envelope.
func (c *Conn) ReadEnvelope(ctx context.Context) (protocol.Envelope, error) {
	var env protocol.Envelope
	if err := wsjson.Read(ctx, c.ws, &env); err != nil {
		return protocol.Envelope{}, err
	}
	return env, nil
}

// WriteEnvelope marshals env to JSON and writes it as a WebSocket message.
func (c *Conn) WriteEnvelope(ctx context.Context, env protocol.Envelope) error {
	return wsjson.Write(ctx, c.ws, env)
}

// Close sends a normal closure frame and closes the underlying connection.
func (c *Conn) Close() error {
	return c.ws.Close(websocket.StatusNormalClosure, "")
}
