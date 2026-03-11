package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/relixdev/protocol"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// RelayClient is a WebSocket client for the Relix relay server.
type RelayClient struct {
	conn      *websocket.Conn
	machineID string

	mu   sync.Mutex
	done chan struct{}
}

// NewClient returns an unconnected RelayClient.
func NewClient() *RelayClient {
	return &RelayClient{done: make(chan struct{})}
}

// Connect dials the relay WebSocket, sends an MsgAuth envelope, and returns
// any error. The caller must call Close when done.
func (c *RelayClient) Connect(ctx context.Context, relayURL, authToken, machineID string) error {
	conn, _, err := websocket.Dial(ctx, relayURL, nil)
	if err != nil {
		return fmt.Errorf("relay dial: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.machineID = machineID
	c.mu.Unlock()

	// Send auth envelope immediately.
	payload, err := json.Marshal(protocol.AuthMessage{Token: authToken})
	if err != nil {
		conn.Close(websocket.StatusInternalError, "marshal auth")
		return fmt.Errorf("marshal auth payload: %w", err)
	}

	env := protocol.NewEnvelope(protocol.MsgAuth, machineID)
	env.Payload = payload

	if err := wsjson.Write(ctx, conn, env); err != nil {
		conn.Close(websocket.StatusInternalError, "send auth")
		return fmt.Errorf("send auth: %w", err)
	}

	return nil
}

// ReadLoop starts a goroutine that reads envelopes from the WebSocket and
// delivers them to the returned channel. The channel is closed when the
// connection closes or ctx is cancelled.
func (c *RelayClient) ReadLoop(ctx context.Context) <-chan protocol.Envelope {
	ch := make(chan protocol.Envelope, 64)
	go func() {
		defer close(ch)
		for {
			var env protocol.Envelope
			c.mu.Lock()
			conn := c.conn
			c.mu.Unlock()
			if conn == nil {
				return
			}
			if err := wsjson.Read(ctx, conn, &env); err != nil {
				return
			}
			select {
			case ch <- env:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch
}

// Send writes an envelope to the WebSocket.
func (c *RelayClient) Send(ctx context.Context, env protocol.Envelope) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return fmt.Errorf("not connected")
	}
	return wsjson.Write(ctx, conn, env)
}

// MachineID returns the machine ID this client authenticated with.
func (c *RelayClient) MachineID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.machineID
}

// Close closes the WebSocket connection.
func (c *RelayClient) Close() {
	c.mu.Lock()
	conn := c.conn
	c.conn = nil
	c.mu.Unlock()
	if conn != nil {
		conn.Close(websocket.StatusNormalClosure, "")
	}
	select {
	case <-c.done:
	default:
		close(c.done)
	}
}
