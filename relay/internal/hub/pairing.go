package hub

import (
	"sync"
	"time"
)

const defaultPairingTTL = 5 * time.Minute

// pairingEntry holds a pending pairing code with its expiry.
type pairingEntry struct {
	userID  string
	expires time.Time
}

// PairingStore stores pending pairing codes with a TTL.
// It is safe for concurrent use.
type PairingStore struct {
	mu      sync.Mutex
	entries map[string]pairingEntry
	ttl     time.Duration
}

// newPairingStore creates a PairingStore with the default 5-minute TTL.
func newPairingStore() *PairingStore {
	return newPairingStoreWithTTL(defaultPairingTTL)
}

// newPairingStoreWithTTL creates a PairingStore with a custom TTL.
func newPairingStoreWithTTL(ttl time.Duration) *PairingStore {
	return &PairingStore{
		entries: make(map[string]pairingEntry),
		ttl:     ttl,
	}
}

// RegisterPairing stores code → userID with a TTL expiry.
// Overwrites any existing entry for that code.
func (ps *PairingStore) RegisterPairing(code, userID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.entries[code] = pairingEntry{
		userID:  userID,
		expires: time.Now().Add(ps.ttl),
	}
}

// ValidatePairing validates and consumes a pairing code (one-time use).
// Returns the associated userID and true if the code is valid and unexpired.
// Returns ("", false) if the code is unknown, expired, or already consumed.
func (ps *PairingStore) ValidatePairing(code string) (string, bool) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	entry, ok := ps.entries[code]
	if !ok {
		return "", false
	}
	// Always delete to enforce one-time use.
	delete(ps.entries, code)

	if time.Now().After(entry.expires) {
		return "", false
	}
	return entry.userID, true
}
