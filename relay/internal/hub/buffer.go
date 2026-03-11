package hub

import (
	"time"

	"github.com/relixdev/protocol"
)

// bufferOptions configures a buffer instance.
type bufferOptions struct {
	maxMessages int
	ttl         time.Duration
}

// bufferedEnvelope pairs an envelope with the wall-clock time it was received,
// enabling sub-second TTL checks independent of Envelope.Timestamp precision.
type bufferedEnvelope struct {
	env      protocol.Envelope
	received time.Time
}

// buffer is a per-machine FIFO queue of envelopes with a max-count cap and TTL.
// It is not safe for concurrent use; callers must hold an appropriate lock.
type buffer struct {
	opts bufferOptions
	msgs []bufferedEnvelope
}

// newBuffer creates a buffer with the given options.
func newBuffer(opts bufferOptions) *buffer {
	return &buffer{
		opts: opts,
		msgs: make([]bufferedEnvelope, 0, opts.maxMessages),
	}
}

// Add appends env to the buffer, first pruning messages older than TTL, then
// dropping the oldest entry if the buffer would exceed maxMessages.
// If env itself arrives after TTL has already elapsed it is silently discarded.
func (b *buffer) Add(env protocol.Envelope) {
	now := time.Now()
	b.pruneAt(now)

	if b.opts.maxMessages > 0 && len(b.msgs) >= b.opts.maxMessages {
		// Drop oldest message to make room.
		b.msgs = b.msgs[1:]
	}
	b.msgs = append(b.msgs, bufferedEnvelope{env: env, received: now})
}

// Drain returns all buffered messages in insertion order and clears the buffer.
func (b *buffer) Drain() []protocol.Envelope {
	out := make([]protocol.Envelope, len(b.msgs))
	for i, be := range b.msgs {
		out[i] = be.env
	}
	b.msgs = make([]bufferedEnvelope, 0, b.opts.maxMessages)
	return out
}

// Len returns the current number of buffered messages.
func (b *buffer) Len() int {
	return len(b.msgs)
}

// pruneAt removes messages whose received time is older than the TTL, relative to now.
func (b *buffer) pruneAt(now time.Time) {
	if b.opts.ttl <= 0 {
		return
	}
	cutoff := now.Add(-b.opts.ttl)
	keep := 0
	for _, m := range b.msgs {
		if !m.received.Before(cutoff) {
			b.msgs[keep] = m
			keep++
		}
	}
	b.msgs = b.msgs[:keep]
}
