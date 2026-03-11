package crypto_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/relixdev/relix/relixctl/internal/crypto"
)

func TestGenerateAndSave(t *testing.T) {
	dir := t.TempDir()
	ks, err := crypto.GenerateAndSave(dir)
	if err != nil {
		t.Fatalf("GenerateAndSave: %v", err)
	}

	// Keys should be non-zero.
	var zeroPub [32]byte
	var zeroPriv [32]byte
	if ks.PublicKey == zeroPub {
		t.Error("PublicKey is all zeros")
	}
	if ks.PrivateKey == zeroPriv {
		t.Error("PrivateKey is all zeros")
	}

	// Files must exist.
	pubPath := filepath.Join(dir, "public.key")
	privPath := filepath.Join(dir, "private.key")
	if _, err := os.Stat(pubPath); err != nil {
		t.Errorf("public.key not found: %v", err)
	}
	if _, err := os.Stat(privPath); err != nil {
		t.Errorf("private.key not found: %v", err)
	}

	// private.key must be 0600.
	info, err := os.Stat(privPath)
	if err != nil {
		t.Fatalf("stat private.key: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("private.key permissions = %04o, want 0600", perm)
	}
}

func TestLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	original, err := crypto.GenerateAndSave(dir)
	if err != nil {
		t.Fatalf("GenerateAndSave: %v", err)
	}

	loaded, err := crypto.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.PublicKey != original.PublicKey {
		t.Errorf("PublicKey mismatch after load")
	}
	if loaded.PrivateKey != original.PrivateKey {
		t.Errorf("PrivateKey mismatch after load")
	}
}

func TestLoadMissingReturnsError(t *testing.T) {
	dir := t.TempDir()
	_, err := crypto.Load(dir)
	if err == nil {
		t.Error("Load on empty dir should return error")
	}
}
