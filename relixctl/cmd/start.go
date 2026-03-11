package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

const pidFile = "daemon.pid"
const logFile = "daemon.log"

func (r *RootCmd) startCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "start",
		Short:        "Start the relay agent daemon",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := r.cfgDir
			if dir == "" {
				var err error
				dir, err = configDir()
				if err != nil {
					return err
				}
			}
			return startDaemon(dir)
		},
	}
}

func (r *RootCmd) stopCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "stop",
		Short:        "Stop the relay agent daemon",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := r.cfgDir
			if dir == "" {
				var err error
				dir, err = configDir()
				if err != nil {
					return err
				}
			}
			return stopDaemon(dir)
		},
	}
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".relixctl")
	return dir, os.MkdirAll(dir, 0700)
}

func pidFilePath(dir string) string {
	return filepath.Join(dir, pidFile)
}

func logFilePath(dir string) string {
	return filepath.Join(dir, logFile)
}

// readPID reads the PID from the PID file. Returns 0 if not found.
func readPID(dir string) (int, error) {
	data, err := os.ReadFile(pidFilePath(dir))
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("invalid PID file: %w", err)
	}
	return pid, nil
}

// isRunning returns true if a process with pid is alive.
func isRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Signal 0 checks if the process exists without killing it.
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}

func startDaemon(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	// Check if already running.
	pid, err := readPID(dir)
	if err != nil {
		return fmt.Errorf("read PID: %w", err)
	}
	if isRunning(pid) {
		return fmt.Errorf("daemon already running (PID %d)", pid)
	}

	// Find the current executable.
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find executable: %w", err)
	}

	// Open log file for stdout/stderr redirection.
	lf, err := os.OpenFile(logFilePath(dir), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer lf.Close()

	// Fork/exec daemon-run subcommand.
	child := exec.Command(exe, "daemon-run", "--config-dir", dir)
	child.Stdout = lf
	child.Stderr = lf
	child.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := child.Start(); err != nil {
		return fmt.Errorf("start daemon: %w", err)
	}

	// Write PID file.
	pidPath := pidFilePath(dir)
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(child.Process.Pid)), 0600); err != nil {
		// Try to kill the child if we can't record the PID.
		_ = child.Process.Kill()
		return fmt.Errorf("write PID file: %w", err)
	}

	return nil
}

func stopDaemon(dir string) error {
	pid, err := readPID(dir)
	if err != nil {
		return fmt.Errorf("read PID: %w", err)
	}
	if pid == 0 {
		return fmt.Errorf("daemon not running (no PID file)")
	}
	if !isRunning(pid) {
		// Stale PID file — clean up.
		_ = os.Remove(pidFilePath(dir))
		return fmt.Errorf("daemon not running (stale PID %d)", pid)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("send SIGTERM to %d: %w", pid, err)
	}

	// Wait for process to exit (up to 10 seconds).
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if !isRunning(pid) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if isRunning(pid) {
		// Force kill if still alive.
		_ = proc.Kill()
	}

	_ = os.Remove(pidFilePath(dir))
	return nil
}
