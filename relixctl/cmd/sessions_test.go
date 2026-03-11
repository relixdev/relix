package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/relixdev/relix/relixctl/cmd"
)

func TestSessionsCmd_NoSessions(t *testing.T) {
	dir := t.TempDir()
	root := cmd.New(dir)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sessions"})

	if err := root.Execute(); err != nil {
		t.Fatalf("sessions command failed: %v", err)
	}

	out := buf.String()
	// With an empty claude dir, should print header or "no sessions" message.
	if out == "" {
		t.Error("expected non-empty output from sessions command")
	}
}

func TestSessionsCmd_TableHeader(t *testing.T) {
	dir := t.TempDir()
	root := cmd.New(dir)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sessions"})

	if err := root.Execute(); err != nil {
		t.Fatalf("sessions command failed: %v", err)
	}

	out := buf.String()
	// Should have table-like output with column headers.
	lower := strings.ToLower(out)
	for _, col := range []string{"id", "project", "status"} {
		if !strings.Contains(lower, col) {
			t.Errorf("sessions output missing column %q; got:\n%s", col, out)
		}
	}
}
