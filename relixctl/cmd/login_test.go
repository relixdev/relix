package cmd_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relixdev/relix/relixctl/cmd"
)

func TestLoginCmd_BrowserFlow(t *testing.T) {
	dir := t.TempDir()
	root := cmd.New(dir)

	// Inject a fake browser login that returns a token.
	root.SetLoginFuncs(
		func(clientID, authURL, tokenURL string) (string, error) {
			return "test-jwt-token-browser", nil
		},
		nil,
	)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"login"})

	if err := root.Execute(); err != nil {
		t.Fatalf("login command failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Logged in as") {
		t.Errorf("expected 'Logged in as' in output, got: %s", out)
	}

	// Verify token was persisted in config.
	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var cfg map[string]string
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if cfg["auth_token"] != "test-jwt-token-browser" {
		t.Errorf("auth_token = %q, want %q", cfg["auth_token"], "test-jwt-token-browser")
	}

	// Verify keypair was generated.
	if _, err := os.Stat(filepath.Join(dir, "keys", "public.key")); err != nil {
		t.Errorf("public key not generated: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "keys", "private.key")); err != nil {
		t.Errorf("private key not generated: %v", err)
	}
}

func TestLoginCmd_HeadlessFlow(t *testing.T) {
	dir := t.TempDir()
	root := cmd.New(dir)

	root.SetLoginFuncs(
		nil,
		func(clientID, deviceCodeURL, tokenURL string) (string, error) {
			return "test-jwt-token-device", nil
		},
	)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"login", "--headless"})

	if err := root.Execute(); err != nil {
		t.Fatalf("login --headless failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Logged in as") {
		t.Errorf("expected 'Logged in as' in output, got: %s", out)
	}

	// Verify device token was persisted.
	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var cfg map[string]string
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if cfg["auth_token"] != "test-jwt-token-device" {
		t.Errorf("auth_token = %q, want %q", cfg["auth_token"], "test-jwt-token-device")
	}
}

func TestLoginCmd_DoesNotRegenerateExistingKeys(t *testing.T) {
	dir := t.TempDir()

	// Pre-create a keypair.
	keysDir := filepath.Join(dir, "keys")
	if err := os.MkdirAll(keysDir, 0700); err != nil {
		t.Fatal(err)
	}
	sentinel := []byte("existing-key-data-abcdef0123456789")
	if err := os.WriteFile(filepath.Join(keysDir, "public.key"), sentinel, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(keysDir, "private.key"), sentinel, 0600); err != nil {
		t.Fatal(err)
	}

	root := cmd.New(dir)
	root.SetLoginFuncs(
		func(clientID, authURL, tokenURL string) (string, error) {
			return "tok", nil
		},
		nil,
	)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"login"})

	if err := root.Execute(); err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Verify existing key was NOT overwritten.
	data, err := os.ReadFile(filepath.Join(keysDir, "public.key"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, sentinel) {
		t.Error("existing public key was overwritten")
	}
}

func TestLoginCmd_AuthError(t *testing.T) {
	dir := t.TempDir()
	root := cmd.New(dir)

	root.SetLoginFuncs(
		func(clientID, authURL, tokenURL string) (string, error) {
			return "", fmt.Errorf("oauth timeout")
		},
		nil,
	)

	root.SetArgs([]string{"login"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "oauth timeout") {
		t.Errorf("error %q does not contain 'oauth timeout'", err.Error())
	}
}
