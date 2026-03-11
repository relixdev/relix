package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/relixdev/relix/relixctl/cmd"
)

func TestStatusCmd_Output(t *testing.T) {
	dir := t.TempDir()
	root := cmd.New(dir)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"status"})

	if err := root.Execute(); err != nil {
		t.Fatalf("status command failed: %v", err)
	}

	out := buf.String()

	// Verify key fields appear in output.
	for _, want := range []string{"relay", "machine", "sessions"} {
		if !strings.Contains(strings.ToLower(out), want) {
			t.Errorf("output missing %q field; got:\n%s", want, out)
		}
	}
}
