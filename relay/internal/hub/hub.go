package hub

import (
	"context"
	"sync"
	"time"

	"github.com/relixdev/protocol"
	"github.com/relixdev/relix/relay/internal/conn"
	"github.com/relixdev/relix/relay/internal/metrics"
)

const (
	defaultBufferMaxMessages = 1000
	defaultBufferTTL         = 24 * time.Hour
)

// Hub manages all active connections, keyed by userID.
type Hub struct {
	mu      sync.RWMutex
	agents  map[string]map[string]*conn.Conn // [userID][machineID] → conn
	mobiles map[string][]*conn.Conn          // [userID] → list of mobile conns
	buffers map[string]*buffer               // [userID] → message buffer
	pairing *PairingStore

	bufOpts bufferOptions
}

// HubOption configures a Hub.
type HubOption func(*Hub)

// WithBufferOptions sets the buffer configuration for the hub.
func WithBufferOptions(maxMessages int, ttl time.Duration) HubOption {
	return func(h *Hub) {
		h.bufOpts = bufferOptions{maxMessages: maxMessages, ttl: ttl}
	}
}

// New creates an empty Hub.
func New(opts ...HubOption) *Hub {
	h := &Hub{
		agents:  make(map[string]map[string]*conn.Conn),
		mobiles: make(map[string][]*conn.Conn),
		buffers: make(map[string]*buffer),
		pairing: newPairingStore(),
		bufOpts: bufferOptions{
			maxMessages: defaultBufferMaxMessages,
			ttl:         defaultBufferTTL,
		},
	}
	for _, o := range opts {
		o(h)
	}
	return h
}

// Pairing returns the hub's PairingStore.
func (h *Hub) Pairing() *PairingStore {
	return h.pairing
}

// RegisterAgent stores the agent connection for (userID, machineID) and
// broadcasts an online status to all mobile connections for that user.
func (h *Hub) RegisterAgent(userID, machineID string, c *conn.Conn) {
	h.mu.Lock()
	if h.agents[userID] == nil {
		h.agents[userID] = make(map[string]*conn.Conn)
	}
	h.agents[userID][machineID] = c
	h.mu.Unlock()

	metrics.ActiveConnections.WithLabelValues("agent").Inc()
	h.broadcastMachineStatus(userID, machineID, protocol.StatusOnline)
}

// UnregisterAgent removes the agent connection for (userID, machineID) and
// broadcasts an offline status to all mobile connections for that user.
func (h *Hub) UnregisterAgent(userID, machineID string) {
	h.mu.Lock()
	if m := h.agents[userID]; m != nil {
		delete(m, machineID)
		if len(m) == 0 {
			delete(h.agents, userID)
		}
	}
	h.mu.Unlock()

	metrics.ActiveConnections.WithLabelValues("agent").Dec()
	h.broadcastMachineStatus(userID, machineID, protocol.StatusOffline)
}

// GetAgent returns the agent connection for (userID, machineID), or nil if not found.
func (h *Hub) GetAgent(userID, machineID string) *conn.Conn {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if m := h.agents[userID]; m != nil {
		return m[machineID]
	}
	return nil
}

// RegisterMobile appends a mobile connection for userID and drains any
// buffered messages to it.
func (h *Hub) RegisterMobile(userID string, c *conn.Conn) {
	h.mu.Lock()
	h.mobiles[userID] = append(h.mobiles[userID], c)
	buf := h.buffers[userID]
	h.mu.Unlock()

	metrics.ActiveConnections.WithLabelValues("mobile").Inc()
	if buf != nil {
		h.drainBufferToConn(userID, buf, c)
	}
}

// UnregisterMobile removes a specific mobile connection for userID.
func (h *Hub) UnregisterMobile(userID string, c *conn.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	list := h.mobiles[userID]
	out := list[:0]
	for _, m := range list {
		if m != c {
			out = append(out, m)
		}
	}
	if len(out) == 0 {
		delete(h.mobiles, userID)
	} else {
		h.mobiles[userID] = out
	}
	metrics.ActiveConnections.WithLabelValues("mobile").Dec()
}

// GetMobiles returns a snapshot of all mobile connections for userID.
func (h *Hub) GetMobiles(userID string) []*conn.Conn {
	h.mu.RLock()
	defer h.mu.RUnlock()
	list := h.mobiles[userID]
	if len(list) == 0 {
		return nil
	}
	snapshot := make([]*conn.Conn, len(list))
	copy(snapshot, list)
	return snapshot
}

// bufferForUser returns (creating if necessary) the buffer for userID.
// Must be called with h.mu held.
func (h *Hub) bufferForUser(userID string) *buffer {
	b := h.buffers[userID]
	if b == nil {
		b = newBuffer(h.bufOpts)
		h.buffers[userID] = b
	}
	return b
}

// drainBufferToConn sends all buffered messages to conn c, ignoring write errors.
func (h *Hub) drainBufferToConn(userID string, buf *buffer, c *conn.Conn) {
	h.mu.Lock()
	msgs := buf.Drain()
	h.mu.Unlock()

	metrics.BufferSize.WithLabelValues(userID).Set(0)

	ctx := context.Background()
	for _, env := range msgs {
		_ = c.WriteEnvelope(ctx, env)
	}
}
