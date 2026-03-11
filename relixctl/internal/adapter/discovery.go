package adapter

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/relixdev/protocol"
)

// DiscoverSessions scans claudeDir/projects/*/*.jsonl and returns a Session
// for each file found. Non-existent or unreadable dirs return an empty slice.
func DiscoverSessions(claudeDir string) []protocol.Session {
	projectsDir := filepath.Join(claudeDir, "projects")
	projectEntries, err := os.ReadDir(projectsDir)
	if err != nil {
		// Non-existent or unreadable — return empty, not an error.
		return []protocol.Session{}
	}

	var sessions []protocol.Session
	for _, projectEntry := range projectEntries {
		if !projectEntry.IsDir() {
			continue
		}

		// Project dir name is URL-encoded path.
		projectName, err := url.QueryUnescape(projectEntry.Name())
		if err != nil {
			// Fall back to raw name on decode failure.
			projectName = projectEntry.Name()
		}

		projDir := filepath.Join(projectsDir, projectEntry.Name())
		fileEntries, err := os.ReadDir(projDir)
		if err != nil {
			continue
		}

		for _, fileEntry := range fileEntries {
			if fileEntry.IsDir() {
				continue
			}
			name := fileEntry.Name()
			if !strings.HasSuffix(name, ".jsonl") {
				continue
			}

			sessionID := strings.TrimSuffix(name, ".jsonl")

			// Use file mtime as StartedAt.
			info, err := fileEntry.Info()
			if err != nil {
				continue
			}

			sessions = append(sessions, protocol.Session{
				ID:        sessionID,
				Tool:      "claude-code",
				Project:   projectName,
				Status:    protocol.SessionIdle,
				StartedAt: info.ModTime().Unix(),
			})
		}
	}

	return sessions
}
