package adapter

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/relixdev/protocol"
)

// AiderCommandFactory creates the exec.Cmd for launching aider.
type AiderCommandFactory func(sessionID string) *exec.Cmd

// defaultAiderCommandFactory returns the real aider CLI command configured for
// scripting mode (no pretty output, auto-accept, no git commits).
func defaultAiderCommandFactory(sessionID string) *exec.Cmd {
	return exec.Command(
		"aider",
		"--no-pretty",
		"--yes-always",
		"--no-auto-commits",
		"--no-suggest-shell-commands",
	)
}

// AiderProcessScanner returns running aider process IDs. Replaceable for tests.
type AiderProcessScanner func(ctx context.Context) ([]aiderProcInfo, error)

type aiderProcInfo struct {
	PID string
	CWD string
}

// defaultAiderProcessScanner uses pgrep to find running aider processes.
func defaultAiderProcessScanner(ctx context.Context) ([]aiderProcInfo, error) {
	// Use pgrep for a safer, more targeted process search.
	cmd := exec.CommandContext(ctx, "pgrep", "-f", "aider")
	out, err := cmd.Output()
	if err != nil {
		// pgrep exits 1 when no processes found — that's not an error.
		return nil, nil
	}

	var procs []aiderProcInfo
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		pid := strings.TrimSpace(line)
		if pid == "" {
			continue
		}
		procs = append(procs, aiderProcInfo{PID: pid, CWD: ""})
	}
	return procs, nil
}

// aiderAttachedSession holds state for a running aider subprocess.
type aiderAttachedSession struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	cancel context.CancelFunc
}

// AiderAdapter implements protocol.CopilotAdapter for Aider.
// Since Aider does not support JSON streaming, this adapter parses text output
// and maps patterns to protocol events.
type AiderAdapter struct {
	commandFactory AiderCommandFactory
	processScanner AiderProcessScanner

	mu       sync.Mutex
	attached map[string]*aiderAttachedSession

	seq atomic.Uint64
}

// NewAiderAdapter creates a new adapter. cmdFactory and scanner may be nil to
// use real defaults.
func NewAiderAdapter(cmdFactory AiderCommandFactory, scanner AiderProcessScanner) *AiderAdapter {
	if cmdFactory == nil {
		cmdFactory = defaultAiderCommandFactory
	}
	if scanner == nil {
		scanner = defaultAiderProcessScanner
	}
	return &AiderAdapter{
		commandFactory: cmdFactory,
		processScanner: scanner,
		attached:       make(map[string]*aiderAttachedSession),
	}
}

// Discover returns active Aider sessions by scanning for running processes.
func (a *AiderAdapter) Discover(ctx context.Context) ([]protocol.Session, error) {
	procs, err := a.processScanner(ctx)
	if err != nil {
		return nil, fmt.Errorf("aider: discover: %w", err)
	}

	sessions := make([]protocol.Session, 0, len(procs))
	for _, p := range procs {
		sessions = append(sessions, protocol.Session{
			ID:     "aider-" + p.PID,
			Tool:   "aider",
			Project: p.CWD,
			Status: protocol.SessionActive,
		})
	}
	return sessions, nil
}

// parseAiderLine classifies a single line of aider text output and returns the
// corresponding event kind and a JSON data payload.
func parseAiderLine(line string) (protocol.PayloadKind, json.RawMessage) {
	trimmed := strings.TrimSpace(line)

	// Skip empty lines and the input prompt marker.
	if trimmed == "" || trimmed == ">" {
		return "", nil
	}

	// Error lines
	if strings.HasPrefix(trimmed, "Error:") || strings.HasPrefix(trimmed, "error:") {
		data, _ := json.Marshal(protocol.ErrorData{Message: trimmed})
		return protocol.PayloadError, data
	}

	// File edit markers: aider shows diffs or file-change notifications.
	if strings.HasPrefix(trimmed, "<<<<<<< SEARCH") ||
		strings.HasPrefix(trimmed, ">>>>>>> REPLACE") ||
		strings.HasPrefix(trimmed, "Applied edit to ") {
		data, _ := json.Marshal(map[string]string{
			"type": "file_edit",
			"text": trimmed,
		})
		return protocol.PayloadToolUse, data
	}

	// Tool-call markers: aider announces shell commands or tool usage.
	if strings.HasPrefix(trimmed, "Run shell command?") ||
		strings.HasPrefix(trimmed, "Running ") {
		data, _ := json.Marshal(map[string]string{
			"type": "tool_call",
			"text": trimmed,
		})
		return protocol.PayloadToolUse, data
	}

	// Everything else is an assistant message.
	data, _ := json.Marshal(protocol.AssistantMessageData{Text: trimmed})
	return protocol.PayloadAssistantMessage, data
}

// Attach spawns an aider subprocess and returns a channel of events.
func (a *AiderAdapter) Attach(ctx context.Context, sessionID string) (<-chan protocol.Event, error) {
	ctx, cancel := context.WithCancel(ctx)

	cmd := a.commandFactory(sessionID)
	cmd = withContext(ctx, cmd)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("aider: stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("aider: stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("aider: start process: %w", err)
	}

	a.mu.Lock()
	a.attached[sessionID] = &aiderAttachedSession{
		cmd:    cmd,
		stdin:  stdin,
		cancel: cancel,
	}
	a.mu.Unlock()

	ch := make(chan protocol.Event, 32)

	go func() {
		defer close(ch)
		defer func() {
			a.mu.Lock()
			delete(a.attached, sessionID)
			a.mu.Unlock()
			cancel()
		}()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()

			kind, data := parseAiderLine(line)
			if kind == "" {
				continue
			}

			ev := protocol.Event{
				Kind: kind,
				Data: data,
				Seq:  a.seq.Add(1),
			}

			select {
			case ch <- ev:
			case <-ctx.Done():
				return
			}
		}

		_ = cmd.Wait()
	}()

	return ch, nil
}

// Send writes user input to the aider subprocess's stdin.
func (a *AiderAdapter) Send(ctx context.Context, sessionID string, msg protocol.UserInput) error {
	a.mu.Lock()
	sess, ok := a.attached[sessionID]
	a.mu.Unlock()
	if !ok {
		return fmt.Errorf("aider: session %q not attached", sessionID)
	}

	var text string
	switch msg.Kind {
	case protocol.InputMessage:
		var data protocol.UserMessageData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return fmt.Errorf("aider: unmarshal user message: %w", err)
		}
		text = data.Text
	case protocol.InputApproval:
		var data protocol.ApprovalResponseData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return fmt.Errorf("aider: unmarshal approval: %w", err)
		}
		if data.Approved {
			text = "y"
		} else {
			text = "n"
		}
	default:
		text = string(msg.Data)
	}

	_, err := fmt.Fprintf(sess.stdin, "%s\n", text)
	if err != nil {
		return fmt.Errorf("aider: write stdin: %w", err)
	}
	return nil
}

// Detach cleanly disconnects from a session.
func (a *AiderAdapter) Detach(sessionID string) error {
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
