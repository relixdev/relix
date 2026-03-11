package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relixdev/relix/relixctl/cmd"
)

func TestUninstallCmd_RemovesConfigDir(t *testing.T) {
	dir := t.TempDir()

	// Populate the config dir with typical files.
	keysDir := filepath.Join(dir, "keys")
	if err := os.MkdirAll(keysDir, 0700); err != nil {
		t.Fatalf("create keys dir: %v", err)
	}
	for _, name := range []string{"config.json", "daemon.pid", "daemon.log"} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("data"), 0600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	if err := os.WriteFile(filepath.Join(keysDir, "public.key"), []byte("key"), 0644); err != nil {
		t.Fatalf("write public.key: %v", err)
	}

	root := cmd.New(dir)
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"uninstall"})

	if err := root.Execute(); err != nil {
		t.Fatalf("uninstall command failed: %v", err)
	}

	// Config dir should be gone.
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("expected config dir %s to be removed, stat err: %v", dir, err)
	}

	out := buf.String()
	if !strings.Contains(strings.ToLower(out), "uninstall") {
		t.Errorf("expected confirmation in output, got: %s", out)
	}
}
