package adapter

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"

	"github.com/relixdev/protocol"
)

// CommandFactory creates the exec.Cmd for a given sessionID.
// This is configurable so tests can substitute a mock process.
type CommandFactory func(sessionID string) *exec.Cmd

// defaultCommandFactory returns the real claude CLI command.
func defaultCommandFactory(sessionID string) *exec.Cmd {
	return exec.Command(
		"claude", "-p",
		"--resume", sessionID,
		"--input-format", "stream-json",
		"--output-format", "stream-json",
		"--verbose",
	)
}

// attachedSession holds state for a running subprocess.
type attachedSession struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	cancel context.CancelFunc
}

// ClaudeCodeAdapter implements protocol.CopilotAdapter for Claude Code.
type ClaudeCodeAdapter struct {
	claudeDir      string
	commandFactory CommandFactory

	mu       sync.Mutex
	attached map[string]*attachedSession

	seq atomic.Uint64
}

// NewClaudeCodeAdapter creates a new adapter. cmdFactory may be nil to use the
// real claude binary.
func NewClaudeCodeAdapter(claudeDir string, cmdFactory CommandFactory) *ClaudeCodeAdapter {
	if cmdFactory == nil {
		cmdFactory = defaultCommandFactory
	}
	return &ClaudeCodeAdapter{
		claudeDir:      claudeDir,
		commandFactory: cmdFactory,
		attached:       make(map[string]*attachedSession),
	}
}

// Discover returns active Claude Code sessions on this machine.
func (a *ClaudeCodeAdapter) Discover(ctx context.Context) ([]protocol.Session, error) {
	return DiscoverSessions(a.claudeDir), nil
}

// claudeEventType is the "type" field in raw Claude Code NDJSON output.
type claudeRawEvent struct {
	Type string          `json:"type"`
	Raw  json.RawMessage // the full line, preserved as Data
}

// mapKind converts Claude Code event type strings to protocol.PayloadKind.
func mapKind(claudeType string) protocol.PayloadKind {
	switch claudeType {
	case "assistant":
		return protocol.PayloadAssistantMessage
	case "tool_use":
		return protocol.PayloadToolUse
	case "result":
		return protocol.PayloadToolResult
	case "error":
		return protocol.PayloadError
	default:
		return protocol.PayloadKind(claudeType)
	}
}

// Attach spawns a claude subprocess for sessionID and returns a channel of events.
// The channel is closed when the process exits or Detach is called.
func (a *ClaudeCodeAdapter) Attach(ctx context.Context, sessionID string) (<-chan protocol.Event, error) {
	ctx, cancel := context.WithCancel(ctx)

	cmd := a.commandFactory(sessionID)
	cmd = withContext(ctx, cmd)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("adapter: stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("adapter: stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("adapter: start process: %w", err)
	}

	a.mu.Lock()
	a.attached[sessionID] = &attachedSession{
		cmd:    cmd,
		stdin:  stdin,
		cancel: cancel,
	}
	a.mu.Unlock()

	ch := make(chan protocol.Event, 32)

	go func() {
		defer close(ch)
		defer func() {
			// Clean up attached state when the process exits.
			a.mu.Lock()
			delete(a.attached, sessionID)
			a.mu.Unlock()
			cancel()
		}()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			// Parse just the "type" field.
			var raw struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(line, &raw); err != nil {
				continue // skip malformed lines
			}

			ev := protocol.Event{
				Kind: mapKind(raw.Type),
				Data: json.RawMessage(line),
				Seq:  a.seq.Add(1),
			}

			select {
			case ch <- ev:
			case <-ctx.Done():
				return
			}
		}

		// Wait for process to finish so we don't leak zombies.
		_ = cmd.Wait()
	}()

	return ch, nil
}

// Send writes user input to the subprocess's stdin pipe.
// For InputMessage: writes a Claude Code user message JSON line.
// For InputApproval: writes the approval data directly.
func (a *ClaudeCodeAdapter) Send(ctx context.Context, sessionID string, msg protocol.UserInput) error {
	a.mu.Lock()
	sess, ok := a.attached[sessionID]
	a.mu.Unlock()
	if !ok {
		return fmt.Errorf("adapter: session %q not attached", sessionID)
	}

	var line []byte
	var err error

	switch msg.Kind {
	case protocol.InputMessage:
		// Unmarshal to get the text field.
		var data protocol.UserMessageData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return fmt.Errorf("adapter: unmarshal user message: %w", err)
		}
		line, err = json.Marshal(map[string]any{
			"type": "user",
			"message": map[string]any{
				"role":    "user",
				"content": data.Text,
			},
		})
		if err != nil {
			return fmt.Errorf("adapter: marshal user message: %w", err)
		}
	case protocol.InputApproval:
		// Write approval data directly.
		line = msg.Data
	default:
		line = msg.Data
	}

	line = append(line, '\n')
	_, err = sess.stdin.Write(line)
	if err != nil {
		return fmt.Errorf("adapter: write stdin: %w", err)
	}
	return nil
}

// Detach cleanly disconnects from a session by cancelling its context and
// closing stdin.
func (a *ClaudeCodeAdapter) Detach(sessionID string) error {
	a.mu.Lock()
	sess, ok := a.attached[sessionID]
	if ok {
		delete(a.attached, sessionID)
	}
	a.mu.Unlock()

	if !ok {
		return nil
	}

	sess.cancel()
	_ = sess.stdin.Close()
	return nil
}

// withContext wraps cmd so it is cancelled when ctx is done.
// exec.CommandContext replaces the Cmd if available; here we reassign.
func withContext(ctx context.Context, cmd *exec.Cmd) *exec.Cmd {
	// Re-create the command with context so it gets killed on cancel.
	newCmd := exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
	newCmd.Env = cmd.Env
	newCmd.Dir = cmd.Dir
	return newCmd
}
