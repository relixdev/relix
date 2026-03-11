package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relixdev/relix/relixctl/cmd"
)

// newTestRoot returns a root command wired to a temp config dir.
func newTestRoot(t *testing.T) (*cmd.RootCmd, string) {
	t.Helper()
	dir := t.TempDir()
	root := cmd.New(dir)
	return root, dir
}

func TestVersionFlag(t *testing.T) {
	root, _ := newTestRoot(t)
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"--version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(buf.String(), "relixctl") {
		t.Errorf("version output %q does not contain 'relixctl'", buf.String())
	}
}

func TestConfigSetAndGet(t *testing.T) {
	root, dir := newTestRoot(t)

	// set relay_url
	root.SetArgs([]string{"config", "set", "relay_url", "wss://custom.relay.com"})
	if err := root.Execute(); err != nil {
		t.Fatalf("config set: %v", err)
	}

	// get relay_url — fresh root pointing at same dir
	root2 := cmd.New(dir)
	buf := &bytes.Buffer{}
	root2.SetOut(buf)
	root2.SetArgs([]string{"config", "get", "relay_url"})
	if err := root2.Execute(); err != nil {
		t.Fatalf("config get: %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "wss://custom.relay.com" {
		t.Errorf("config get relay_url = %q, want %q", got, "wss://custom.relay.com")
	}
}

func TestConfigSetAuthToken(t *testing.T) {
	root, dir := newTestRoot(t)

	root.SetArgs([]string{"config", "set", "auth_token", "tok_xyz"})
	if err := root.Execute(); err != nil {
		t.Fatalf("config set auth_token: %v", err)
	}

	// Verify config file was written
	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("read config.json: %v", err)
	}
	if !strings.Contains(string(data), "tok_xyz") {
		t.Errorf("config.json does not contain token: %s", data)
	}
}

func TestSubcommandStubsExist(t *testing.T) {
	stubs := [][]string{
		{"login"},
		{"pair"},
		{"status"},
		{"sessions"},
		{"start"},
		{"stop"},
		{"uninstall"},
	}
	for _, args := range stubs {
		root, _ := newTestRoot(t)
		root.SetArgs(args)
		// Stub commands should not return an "unknown command" error.
		err := root.Execute()
		if err != nil && strings.Contains(err.Error(), "unknown command") {
			t.Errorf("command %q not registered: %v", args[0], err)
		}
	}
}
