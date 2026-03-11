package cmd_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/relixdev/relix/relixctl/cmd"
)

// TestStartWritesPIDFile verifies that `relixctl start` creates a PID file.
func TestStartWritesPIDFile(t *testing.T) {
	cfgDir := t.TempDir()

	root := cmd.New(cfgDir)
	root.SetArgs([]string{"start"})
	if err := root.Execute(); err != nil {
		t.Fatalf("start: %v", err)
	}

	pidFile := filepath.Join(cfgDir, "daemon.pid")
	data, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("PID file not created: %v", err)
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil || pid <= 0 {
		t.Fatalf("invalid PID in file: %q", string(data))
	}

	// Clean up: kill the process we started.
	proc, err := os.FindProcess(pid)
	if err == nil {
		_ = proc.Kill()
		_, _ = proc.Wait()
	}
	_ = os.Remove(pidFile)
}

// TestStartFailsIfAlreadyRunning verifies double-start is rejected.
func TestStartFailsIfAlreadyRunning(t *testing.T) {
	cfgDir := t.TempDir()

	root := cmd.New(cfgDir)
	root.SetArgs([]string{"start"})
	if err := root.Execute(); err != nil {
		t.Fatalf("first start: %v", err)
	}

	pidFile := filepath.Join(cfgDir, "daemon.pid")
	data, _ := os.ReadFile(pidFile)
	pid, _ := strconv.Atoi(string(data))
	defer func() {
		if proc, err := os.FindProcess(pid); err == nil {
			_ = proc.Kill()
			_, _ = proc.Wait()
		}
		_ = os.Remove(pidFile)
	}()

	root2 := cmd.New(cfgDir)
	root2.SetArgs([]string{"start"})
	if err := root2.Execute(); err == nil {
		t.Fatal("expected error starting daemon twice, got nil")
	}
}

// TestStopRemovesPIDFile verifies `relixctl stop` removes the PID file and terminates the process.
func TestStopRemovesPIDFile(t *testing.T) {
	cfgDir := t.TempDir()

	// Start the daemon first.
	root := cmd.New(cfgDir)
	root.SetArgs([]string{"start"})
	if err := root.Execute(); err != nil {
		t.Fatalf("start: %v", err)
	}

	pidFile := filepath.Join(cfgDir, "daemon.pid")
	if _, err := os.Stat(pidFile); err != nil {
		t.Fatalf("PID file missing after start: %v", err)
	}

	// Give the process a moment to start.
	time.Sleep(50 * time.Millisecond)

	// Stop it.
	root2 := cmd.New(cfgDir)
	root2.SetArgs([]string{"stop"})
	if err := root2.Execute(); err != nil {
		t.Fatalf("stop: %v", err)
	}

	// PID file should be gone.
	if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
		t.Fatal("PID file still exists after stop")
	}
}
