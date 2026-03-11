package hub

import (
	"testing"
	"time"

	"github.com/relixdev/protocol"
)

func makeEnv(machineID string) protocol.Envelope {
	return protocol.NewEnvelope(protocol.MsgSessionEvent, machineID)
}

// --- Add / Drain / Len ---

func TestBufferDrainInOrder(t *testing.T) {
	b := newBuffer(bufferOptions{maxMessages: 10, ttl: time.Hour})

	for i := 0; i < 5; i++ {
		e := makeEnv("m1")
		e.SessionID = string(rune('A' + i))
		b.Add(e)
	}

	msgs := b.Drain()
	if len(msgs) != 5 {
		t.Fatalf("want 5 messages, got %d", len(msgs))
	}
	for i, m := range msgs {
		want := string(rune('A' + i))
		if m.SessionID != want {
			t.Errorf("index %d: want SessionID %q, got %q", i, want, m.SessionID)
		}
	}
}

func TestBufferDrainEmpties(t *testing.T) {
	b := newBuffer(bufferOptions{maxMessages: 10, ttl: time.Hour})
	b.Add(makeEnv("m1"))
	b.Drain()

	if b.Len() != 0 {
		t.Errorf("after Drain: want Len 0, got %d", b.Len())
	}
	if got := b.Drain(); len(got) != 0 {
		t.Errorf("second Drain should be empty, got %d", len(got))
	}
}

func TestBufferLenTracksMessages(t *testing.T) {
	b := newBuffer(bufferOptions{maxMessages: 10, ttl: time.Hour})
	if b.Len() != 0 {
		t.Fatalf("initial Len: want 0, got %d", b.Len())
	}
	b.Add(makeEnv("m1"))
	b.Add(makeEnv("m1"))
	if b.Len() != 2 {
		t.Errorf("want Len 2, got %d", b.Len())
	}
}

// --- Max message cap ---

func TestBufferRespectsMaxCount(t *testing.T) {
	const max = 5
	b := newBuffer(bufferOptions{maxMessages: max, ttl: time.Hour})

	for i := 0; i < max+3; i++ {
		e := makeEnv("m1")
		e.SessionID = string(rune('A' + i))
		b.Add(e)
	}

	msgs := b.Drain()
	if len(msgs) != max {
		t.Fatalf("want %d messages (oldest dropped), got %d", max, len(msgs))
	}
	// Should retain newest: indices 3..7 → 'D'..'H'
	if msgs[0].SessionID != "D" {
		t.Errorf("first retained msg: want SessionID %q, got %q", "D", msgs[0].SessionID)
	}
}

// --- TTL pruning ---

// TestBufferPrunesOldMessages verifies that messages whose received time has
// passed the TTL are pruned when a subsequent Add is called.
func TestBufferPrunesOldMessages(t *testing.T) {
	ttl := 60 * time.Millisecond
	b := newBuffer(bufferOptions{maxMessages: 100, ttl: ttl})

	// Add a message that is within TTL now.
	old := makeEnv("m1")
	old.SessionID = "old"
	b.Add(old)

	if b.Len() != 1 {
		t.Fatalf("setup: want 1 message before TTL expires, got %d", b.Len())
	}

	// Wait for the message to expire, then trigger pruning via Add.
	time.Sleep(ttl + 20*time.Millisecond)
	fresh := makeEnv("m1")
	fresh.SessionID = "fresh"
	b.Add(fresh)

	msgs := b.Drain()
	if len(msgs) != 1 {
		t.Fatalf("want 1 message after TTL prune, got %d", len(msgs))
	}
	if msgs[0].SessionID != "fresh" {
		t.Errorf("want fresh message, got %q", msgs[0].SessionID)
	}
}

// TestBufferDrainAfterTTLExpiry verifies Drain returns only live messages
// when called after TTL expires (prune happens on next Add or explicitly).
func TestBufferDrainAfterTTLExpiry(t *testing.T) {
	ttl := 60 * time.Millisecond
	b := newBuffer(bufferOptions{maxMessages: 100, ttl: ttl})

	expired := makeEnv("m1")
	expired.SessionID = "expired"
	b.Add(expired)

	// Let it expire.
	time.Sleep(ttl + 20*time.Millisecond)

	// Add a fresh message — this triggers the prune of the expired one.
	fresh := makeEnv("m1")
	fresh.SessionID = "fresh"
	b.Add(fresh)

	msgs := b.Drain()
	for _, m := range msgs {
		if m.SessionID == "expired" {
			t.Error("expired message should have been pruned")
		}
	}
	if len(msgs) != 1 || msgs[0].SessionID != "fresh" {
		t.Errorf("want only fresh message, got %v", msgs)
	}
}
