package relay

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/relixdev/protocol"
)

const (
	maxBackoff   = 60 * time.Second
	initialDelay = 1 * time.Second
	queueSize    = 100
)

// ReconnectOptions configures a ReconnectingClient.
type ReconnectOptions struct {
	RelayURL     string
	AuthToken    string
	MachineID    string
	OnConnect    func()
	OnDisconnect func()
}

// ReconnectingClient wraps RelayClient with automatic reconnection and
// exponential backoff. Messages queued while disconnected are delivered after
// reconnect (bounded to 100 messages; oldest are dropped when full).
type ReconnectingClient struct {
	opts ReconnectOptions

	mu     sync.Mutex
	client *RelayClient

	queue chan protocol.Envelope

	stopOnce sync.Once
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewReconnectingClient creates a ReconnectingClient. Call Start to begin
// connecting.
func NewReconnectingClient(opts ReconnectOptions) *ReconnectingClient {
	return &ReconnectingClient{
		opts:   opts,
		queue:  make(chan protocol.Envelope, queueSize),
		stopCh: make(chan struct{}),
	}
}

// Start begins the background connect-and-reconnect loop.
func (rc *ReconnectingClient) Start(ctx context.Context) {
	rc.wg.Add(1)
	go rc.loop(ctx)
}

// Stop shuts down the reconnect loop and closes any active connection.
func (rc *ReconnectingClient) Stop() {
	rc.stopOnce.Do(func() { close(rc.stopCh) })
	rc.wg.Wait()
}

// Send queues an envelope for delivery. If currently connected it is sent
// immediately; otherwise it is buffered. If the buffer is full the oldest
// message is dropped to make room.
func (rc *ReconnectingClient) Send(ctx context.Context, env protocol.Envelope) error {
	rc.mu.Lock()
	client := rc.client
	rc.mu.Unlock()

	if client != nil {
		if err := client.Send(ctx, env); err == nil {
			return nil
		}
		// Fall through to queue on error.
	}

	// Non-blocking enqueue; drop oldest if full.
	select {
	case rc.queue <- env:
	default:
		// Drop oldest.
		select {
		case <-rc.queue:
		default:
		}
		rc.queue <- env
	}
	return nil
}

func (rc *ReconnectingClient) loop(ctx context.Context) {
	defer rc.wg.Done()

	delay := initialDelay

	for {
		select {
		case <-rc.stopCh:
			rc.closeClient()
			return
		case <-ctx.Done():
			rc.closeClient()
			return
		default:
		}

		c := NewClient()
		err := c.Connect(ctx, rc.opts.RelayURL, rc.opts.AuthToken, rc.opts.MachineID)
		if err != nil {
			log.Printf("relay: connect failed: %v; retrying in %s", err, delay)
			select {
			case <-time.After(delay):
			case <-rc.stopCh:
				return
			case <-ctx.Done():
				return
			}
			delay = backoff(delay)
			continue
		}

		// Connected.
		delay = initialDelay
		rc.mu.Lock()
		rc.client = c
		rc.mu.Unlock()

		if rc.opts.OnConnect != nil {
			rc.opts.OnConnect()
		}

		// Drain queue.
		rc.drainQueue(ctx, c)

		// Block until connection drops.
		readCtx, readCancel := context.WithCancel(ctx)
		ch := c.ReadLoop(readCtx)
		rc.waitForDrop(ctx, ch)
		readCancel()

		rc.mu.Lock()
		rc.client = nil
		rc.mu.Unlock()

		c.Close()

		if rc.opts.OnDisconnect != nil {
			rc.opts.OnDisconnect()
		}

		// Check if we should stop before retrying.
		select {
		case <-rc.stopCh:
			return
		case <-ctx.Done():
			return
		default:
		}

		log.Printf("relay: disconnected; reconnecting in %s", delay)
		select {
		case <-time.After(delay):
		case <-rc.stopCh:
			return
		case <-ctx.Done():
			return
		}
		delay = backoff(delay)
	}
}

func (rc *ReconnectingClient) drainQueue(ctx context.Context, c *RelayClient) {
	for {
		select {
		case env := <-rc.queue:
			if err := c.Send(ctx, env); err != nil {
				// Re-queue on failure (best effort).
				select {
				case rc.queue <- env:
				default:
				}
				return
			}
		default:
			return
		}
	}
}

// waitForDrop blocks until the read channel closes (connection dropped) or
// stop/ctx is triggered.
func (rc *ReconnectingClient) waitForDrop(ctx context.Context, ch <-chan protocol.Envelope) {
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
			// Discard envelopes at this layer; callers use Bridge.
		case <-rc.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (rc *ReconnectingClient) closeClient() {
	rc.mu.Lock()
	c := rc.client
	rc.client = nil
	rc.mu.Unlock()
	if c != nil {
		c.Close()
	}
}

func backoff(d time.Duration) time.Duration {
	d *= 2
	if d > maxBackoff {
		return maxBackoff
	}
	return d
}
