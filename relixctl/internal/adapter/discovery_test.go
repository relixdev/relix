package adapter

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/relixdev/protocol"
)

func TestDiscoverSessions_empty(t *testing.T) {
	dir := t.TempDir()
	sessions := DiscoverSessions(dir)
	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestDiscoverSessions_nonexistent(t *testing.T) {
	sessions := DiscoverSessions("/tmp/relix-nonexistent-dir-xyz-123")
	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions for nonexistent dir, got %d", len(sessions))
	}
}

func TestDiscoverSessions_finds_sessions(t *testing.T) {
	dir := t.TempDir()

	// Claude Code stores sessions under projects/<url-encoded-path>/<sessionID>.jsonl
	// URL-encoded project name: %2FUsers%2Fzach%2Fmyproject → /Users/zach/myproject
	projectDir := filepath.Join(dir, "projects", "%2FUsers%2Fzach%2Fmyproject")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	sessionFile := filepath.Join(projectDir, "abc123def456.jsonl")
	if err := os.WriteFile(sessionFile, []byte(`{"type":"summary"}`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	before := time.Now().Add(-time.Second)
	sessions := DiscoverSessions(dir)
	after := time.Now().Add(time.Second)

	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}

	s := sessions[0]
	if s.ID != "abc123def456" {
		t.Errorf("expected ID %q, got %q", "abc123def456", s.ID)
	}
	if s.Tool != "claude-code" {
		t.Errorf("expected Tool %q, got %q", "claude-code", s.Tool)
	}
	if s.Project != "/Users/zach/myproject" {
		t.Errorf("expected Project %q, got %q", "/Users/zach/myproject", s.Project)
	}
	if s.Status != protocol.SessionIdle {
		t.Errorf("expected Status %q, got %q", protocol.SessionIdle, s.Status)
	}

	startedAt := time.Unix(s.StartedAt, 0)
	if startedAt.Before(before) || startedAt.After(after) {
		t.Errorf("StartedAt %v not in expected range [%v, %v]", startedAt, before, after)
	}
}

func TestDiscoverSessions_multiple_projects(t *testing.T) {
	dir := t.TempDir()

	projects := []string{"%2Fhome%2Falice%2Fproj1", "%2Fhome%2Fbob%2Fproj2"}
	sessionIDs := []string{"session001", "session002"}

	for i, proj := range projects {
		projDir := filepath.Join(dir, "projects", proj)
		if err := os.MkdirAll(projDir, 0755); err != nil {
			t.Fatal(err)
		}
		f := filepath.Join(projDir, sessionIDs[i]+".jsonl")
		if err := os.WriteFile(f, []byte{}, 0644); err != nil {
			t.Fatal(err)
		}
	}

	sessions := DiscoverSessions(dir)
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}

	ids := map[string]bool{}
	for _, s := range sessions {
		ids[s.ID] = true
	}
	for _, id := range sessionIDs {
		if !ids[id] {
			t.Errorf("expected session ID %q not found", id)
		}
	}
}

func TestDiscoverSessions_ignores_non_jsonl(t *testing.T) {
	dir := t.TempDir()
	projDir := filepath.Join(dir, "projects", "%2Ftmp%2Ftest")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatal(err)
	}
	// A .json file should be ignored, only .jsonl counts
	if err := os.WriteFile(filepath.Join(projDir, "notasession.json"), []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projDir, "realsession.jsonl"), []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	sessions := DiscoverSessions(dir)
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].ID != "realsession" {
		t.Errorf("expected ID %q, got %q", "realsession", sessions[0].ID)
	}
}
