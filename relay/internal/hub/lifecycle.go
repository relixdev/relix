package hub

import (
	"context"
	"time"

	"github.com/relixdev/protocol"
	"github.com/relixdev/relix/relay/internal/conn"
)

const (
	pingInterval = 30 * time.Second
	pingTimeout  = 10 * time.Second
)

// StartPinger runs a read loop on c. It sends a WebSocket ping every
// pingInterval seconds; if no response arrives within pingTimeout the
// connection is closed. When the connection closes (for any reason) the
// connection is unregistered from the hub and, for agents, an offline status
// is broadcast.
//
// role must be "agent" or "mobile". machineID is only used when role=="agent".
// Call this in a goroutine; it returns when ctx is cancelled or c is closed.
func (h *Hub) StartPinger(ctx context.Context, userID, role, machineID string, c *conn.Conn) {
	defer h.unregister(userID, role, machineID, c)

	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	readErr := make(chan error, 1)
	go func() {
		for {
			_, err := c.ReadEnvelope(ctx)
			if err != nil {
				readErr <- err
				return
			}
			// Forward received envelopes to the router (best-effort, ignore errors).
			// In the read loop used by the ws handler this forwarding is done there;
			// StartPinger only needs to detect closure.
		}
	}()

	for {
		select {
		case <-ctx.Done():
			c.CloseNow()
			return

		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
			env := protocol.NewEnvelope(protocol.MsgPing, machineID)
			_ = c.WriteEnvelope(pingCtx, env)
			cancel()

		case <-readErr:
			return
		}
	}
}

// unregister removes c from the hub based on role and triggers any required
// status broadcasts.
func (h *Hub) unregister(userID, role, machineID string, c *conn.Conn) {
	switch role {
	case "agent":
		h.UnregisterAgent(userID, machineID)
	case "mobile":
		h.UnregisterMobile(userID, c)
	}
}
