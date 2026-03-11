package cmd

import (
	"context"
	"os"

	"github.com/relixdev/protocol"
	"github.com/relixdev/relix/relixctl/internal/adapter"
)

// homeDir returns the current user's home directory.
func homeDir() (string, error) {
	return os.UserHomeDir()
}

// discoverSessions returns active Claude Code sessions for the given claudeDir.
func discoverSessions(claudeDir string) []protocol.Session {
	adp := adapter.NewClaudeCodeAdapter(claudeDir, nil)
	sessions, _ := adp.Discover(context.Background())
	return sessions
}
