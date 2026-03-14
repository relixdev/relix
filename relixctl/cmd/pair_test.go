package cmd_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relixdev/protocol"
	"github.com/relixdev/relix/relixctl/cmd"
	"github.com/relixdev/relix/relixctl/internal/crypto"
)

func TestPairCmd_Success(t *testing.T) {
	dir := t.TempDir()

	// Pre-set auth token in config so pair doesn't reject us.
	cfg := map[string]string{
		"relay_url":  "wss://relay.example.com",
		"auth_token": "test-jwt",
		"machine_id": "mch-test",
	}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0600); err != nil {
		t.Fatal(err)
	}

	// Pre-create keypair.
	keysDir := filepath.Join(dir, "keys")
	kp, err := crypto.GenerateAndSave(keysDir)
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	_ = kp

	var peerPub protocol.PublicKey
	peerPub[0] = 0x42 // distinguishable

	root := cmd.New(dir)
	root.SetPairFunc(func(relayURL, authToken, code string, ks *crypto.KeyPair, kDir string) (protocol.PublicKey, error) {
		if authToken != "test-jwt" {
			t.Errorf("pair got auth token %q, want %q", authToken, "test-jwt")
		}
		if code != "123456" {
			t.Errorf("pair got code %q, want %q", code, "123456")
		}
		if !strings.HasSuffix(relayURL, "/pair") {
			t.Errorf("pair relay URL %q should end with /pair", relayURL)
		}
		return peerPub, nil
	})

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"pair", "123456"})

	if err := root.Execute(); err != nil {
		t.Fatalf("pair command failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "123456") {
		t.Errorf("output should display the code, got: %s", out)
	}
	if !strings.Contains(out, "Paired with") {
		t.Errorf("output should confirm pairing, got: %s", out)
	}
}

func TestPairCmd_RequiresLogin(t *testing.T) {
	dir := t.TempDir()

	// Config with no auth token.
	cfg := map[string]string{
		"relay_url":  "wss://relay.example.com",
		"machine_id": "mch-test",
	}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0600); err != nil {
		t.Fatal(err)
	}

	root := cmd.New(dir)
	root.SetArgs([]string{"pair", "123456"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for unauthenticated pair")
	}
	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("error %q should mention 'not logged in'", err.Error())
	}
}

func TestPairCmd_RequiresCodeArg(t *testing.T) {
	dir := t.TempDir()
	root := cmd.New(dir)

	var buf bytes.Buffer
	root.SetErr(&buf)
	root.SetArgs([]string{"pair"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no code arg provided")
	}
}

func TestPairCmd_PairError(t *testing.T) {
	dir := t.TempDir()

	cfg := map[string]string{
		"relay_url":  "wss://relay.example.com",
		"auth_token": "test-jwt",
		"machine_id": "mch-test",
	}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0600); err != nil {
		t.Fatal(err)
	}

	// Pre-create keypair.
	keysDir := filepath.Join(dir, "keys")
	if _, err := crypto.GenerateAndSave(keysDir); err != nil {
		t.Fatal(err)
	}

	root := cmd.New(dir)
	root.SetPairFunc(func(relayURL, authToken, code string, ks *crypto.KeyPair, kDir string) (protocol.PublicKey, error) {
		return protocol.PublicKey{}, fmt.Errorf("connection refused")
	})

	root.SetArgs([]string{"pair", "999999"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("error %q does not contain 'connection refused'", err.Error())
	}
}

func TestPairCmd_GeneratesKeypairIfMissing(t *testing.T) {
	dir := t.TempDir()

	cfg := map[string]string{
		"relay_url":  "wss://relay.example.com",
		"auth_token": "test-jwt",
		"machine_id": "mch-test",
	}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0600); err != nil {
		t.Fatal(err)
	}

	// No pre-created keypair — pair should generate one.

	root := cmd.New(dir)
	root.SetPairFunc(func(relayURL, authToken, code string, ks *crypto.KeyPair, kDir string) (protocol.PublicKey, error) {
		// Verify we got a valid keypair.
		if ks == nil {
			t.Error("keypair should not be nil")
		}
		return protocol.PublicKey{}, nil
	})

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"pair", "111111"})

	if err := root.Execute(); err != nil {
		t.Fatalf("pair failed: %v", err)
	}

	// Verify keypair files exist.
	keysDir := filepath.Join(dir, "keys")
	if _, err := os.Stat(filepath.Join(keysDir, "public.key")); err != nil {
		t.Errorf("public key not generated: %v", err)
	}
}
