package hub

import (
	"testing"
	"time"
)

func TestPairingValidateSucceeds(t *testing.T) {
	ps := newPairingStore()
	ps.RegisterPairing("CODE1", "user1")

	userID, ok := ps.ValidatePairing("CODE1")
	if !ok {
		t.Fatal("ValidatePairing: want ok=true, got false")
	}
	if userID != "user1" {
		t.Errorf("want userID %q, got %q", "user1", userID)
	}
}

func TestPairingConsumedOnValidate(t *testing.T) {
	ps := newPairingStore()
	ps.RegisterPairing("CODE1", "user1")

	// First validate succeeds.
	if _, ok := ps.ValidatePairing("CODE1"); !ok {
		t.Fatal("first validate: want ok=true")
	}

	// Second validate fails (one-time use).
	if _, ok := ps.ValidatePairing("CODE1"); ok {
		t.Error("second validate: want ok=false (code consumed)")
	}
}

func TestPairingUnknownCodeFails(t *testing.T) {
	ps := newPairingStore()

	if _, ok := ps.ValidatePairing("UNKNOWN"); ok {
		t.Error("unknown code: want ok=false")
	}
}

func TestPairingExpiredCodeFails(t *testing.T) {
	ttl := 50 * time.Millisecond
	ps := newPairingStoreWithTTL(ttl)
	ps.RegisterPairing("CODE1", "user1")

	time.Sleep(ttl + 20*time.Millisecond)

	if _, ok := ps.ValidatePairing("CODE1"); ok {
		t.Error("expired code: want ok=false")
	}
}

func TestPairingMultipleCodes(t *testing.T) {
	ps := newPairingStore()
	ps.RegisterPairing("CODE-A", "userA")
	ps.RegisterPairing("CODE-B", "userB")

	if uid, ok := ps.ValidatePairing("CODE-A"); !ok || uid != "userA" {
		t.Errorf("CODE-A: want userA/true, got %q/%v", uid, ok)
	}
	if uid, ok := ps.ValidatePairing("CODE-B"); !ok || uid != "userB" {
		t.Errorf("CODE-B: want userB/true, got %q/%v", uid, ok)
	}
}

func TestPairingOverwriteCode(t *testing.T) {
	ps := newPairingStore()
	ps.RegisterPairing("CODE1", "user1")
	// Overwrite with different user.
	ps.RegisterPairing("CODE1", "user2")

	uid, ok := ps.ValidatePairing("CODE1")
	if !ok {
		t.Fatal("want ok=true after overwrite")
	}
	if uid != "user2" {
		t.Errorf("want user2 after overwrite, got %q", uid)
	}
}
