package daemon

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/relixdev/protocol"
)

const defaultPollInterval = 10 * time.Second

// SentEvent records an event forwarded to the bridge, used by tests.
type SentEvent struct {
	Event     protocol.Event
	SessionID string
}

// Adapter is the interface the daemon uses to interact with Claude Code.
type Adapter interface {
	Discover(ctx context.Context) ([]protocol.Session, error)
	Attach(ctx context.Context, sessionID string) (<-chan protocol.Event, error)
	Send(ctx context.Context, sessionID string, msg protocol.UserInput) error
	Detach(sessionID string) error
}

// Bridge is the interface the daemon uses to communicate with the relay.
type Bridge interface {
	SendEvent(ctx context.Context, event protocol.Event, sessionID string) error
	ReceiveLoop(ctx context.Context) <-chan protocol.Payload
	SendMachineStatus(ctx context.Context, status string) error
}

// Options configures Daemon behaviour. Zero values use defaults.
type Options struct {
	// PollInterval is how often sessions are re-discovered.
	// Defaults to 10 seconds.
	PollInterval time.Duration
}

// Daemon orchestrates session discovery, event forwarding, and relay I/O.
type Daemon struct {
	adapter Adapter
	bridge  Bridge
	opts    Options
}

// New creates a Daemon with the supplied adapter and bridge.
func New(adapter Adapter, bridge Bridge, opts Options) *Daemon {
	if opts.PollInterval == 0 {
		opts.PollInterval = defaultPollInterval
	}
	return &Daemon{
		adapter: adapter,
		bridge:  bridge,
		opts:    opts,
	}
}

// Run starts the daemon main loop. It blocks until ctx is cancelled.
func (d *Daemon) Run(ctx context.Context) error {
	// Send "online" status.
	if err := d.bridge.SendMachineStatus(ctx, "online"); err != nil {
		log.Printf("daemon: send online status: %v", err)
	}

	// attached tracks the sessions we have goroutines running for.
	// key: sessionID, value: cancel func for that session's goroutine.
	attached := make(map[string]context.CancelFunc)
	var wg sync.WaitGroup

	defer func() {
		// Send "offline" and detach all on shutdown.
		offCtx, offCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer offCancel()
		if err := d.bridge.SendMachineStatus(offCtx, "offline"); err != nil {
			log.Printf("daemon: send offline status: %v", err)
		}
		for id, cancel := range attached {
			cancel()
			if err := d.adapter.Detach(id); err != nil {
				log.Printf("daemon: detach %s: %v", id, err)
			}
		}
		wg.Wait()
	}()

	// Start the relay receive loop.
	inbound := d.bridge.ReceiveLoop(ctx)

	ticker := time.NewTicker(d.opts.PollInterval)
	defer ticker.Stop()

	// Run an initial discovery immediately.
	d.reconcile(ctx, attached, &wg)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			d.reconcile(ctx, attached, &wg)

		case payload, ok := <-inbound:
			if !ok {
				// Bridge closed; keep running so reconnect can re-establish.
				inbound = nil
				continue
			}
			d.dispatchInbound(ctx, attached, payload)
		}
	}
}

// reconcile discovers current sessions and attaches/detaches as needed.
func (d *Daemon) reconcile(ctx context.Context, attached map[string]context.CancelFunc, wg *sync.WaitGroup) {
	sessions, err := d.adapter.Discover(ctx)
	if err != nil {
		log.Printf("daemon: discover: %v", err)
		return
	}

	current := make(map[string]bool, len(sessions))
	for _, s := range sessions {
		current[s.ID] = true
	}

	// Attach new sessions.
	for _, s := range sessions {
		if _, ok := attached[s.ID]; ok {
			continue
		}
		sessCtx, cancel := context.WithCancel(ctx)
		attached[s.ID] = cancel

		eventCh, err := d.adapter.Attach(sessCtx, s.ID)
		if err != nil {
			log.Printf("daemon: attach %s: %v", s.ID, err)
			cancel()
			delete(attached, s.ID)
			continue
		}

		sessionID := s.ID
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				cancel()
			}()
			for ev := range eventCh {
				if err := d.bridge.SendEvent(sessCtx, ev, sessionID); err != nil {
					log.Printf("daemon: send event for %s: %v", sessionID, err)
				}
			}
		}()
	}

	// Detach removed sessions.
	for id, cancel := range attached {
		if !current[id] {
			cancel()
			if err := d.adapter.Detach(id); err != nil {
				log.Printf("daemon: detach %s: %v", id, err)
			}
			delete(attached, id)
		}
	}
}

// dispatchInbound routes an inbound payload to the correct adapter session.
// We dispatch to all attached sessions since the payload's session routing
// is handled at the bridge layer; the payload carries session context via
// the envelope (already filtered by ReceiveLoop). We use the first attached
// session as the target, or decode the session ID from the payload if present.
func (d *Daemon) dispatchInbound(ctx context.Context, attached map[string]context.CancelFunc, p protocol.Payload) {
	// Determine which session to route to. The payload data may embed a
	// session_id field; if not, send to all attached sessions.
	var target struct {
		SessionID string `json:"session_id"`
	}
	_ = json.Unmarshal(p.Data, &target)

	input := protocol.UserInput{
		Data: p.Data,
	}
	switch p.Kind {
	case protocol.PayloadUserMessage:
		input.Kind = protocol.InputMessage
	case protocol.PayloadApproval:
		input.Kind = protocol.InputApproval
	default:
		input.Kind = protocol.InputMessage
	}

	if target.SessionID != "" {
		if _, ok := attached[target.SessionID]; ok {
			if err := d.adapter.Send(ctx, target.SessionID, input); err != nil {
				log.Printf("daemon: send to %s: %v", target.SessionID, err)
			}
		}
		return
	}

	// Broadcast to all attached sessions.
	for id := range attached {
		if err := d.adapter.Send(ctx, id, input); err != nil {
			log.Printf("daemon: send to %s: %v", id, err)
		}
	}
}
