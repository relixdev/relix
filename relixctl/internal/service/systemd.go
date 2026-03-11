package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const systemdServiceName = "relixctl.service"

// GenerateUnit returns the content of a systemd user service unit file for
// the given binary path.
func GenerateUnit(binaryPath string) string {
	return fmt.Sprintf(`[Unit]
Description=Relix Agent
After=network.target

[Service]
ExecStart=%s daemon-run
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=default.target
`, binaryPath)
}

// SystemdManager implements ServiceManager using systemd --user.
type SystemdManager struct {
	unitDir string // override for testing; empty uses ~/.config/systemd/user
}

// NewSystemdManager creates a SystemdManager targeting the standard
// ~/.config/systemd/user directory.
func NewSystemdManager() *SystemdManager {
	return &SystemdManager{}
}

func (m *SystemdManager) unitPath() (string, error) {
	if m.unitDir != "" {
		return filepath.Join(m.unitDir, systemdServiceName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "systemd", "user", systemdServiceName), nil
}

// Install writes the unit file, then enables and starts the service.
func (m *SystemdManager) Install(binaryPath string) error {
	path, err := m.unitPath()
	if err != nil {
		return fmt.Errorf("systemd install: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("systemd install: create dir: %w", err)
	}
	content := GenerateUnit(binaryPath)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("systemd install: write unit: %w", err)
	}
	if out, err := exec.Command("systemctl", "--user", "enable", systemdServiceName).CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl enable: %w\n%s", err, out)
	}
	if out, err := exec.Command("systemctl", "--user", "start", systemdServiceName).CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl start: %w\n%s", err, out)
	}
	return nil
}

// Uninstall stops, disables, and removes the service unit file.
func (m *SystemdManager) Uninstall() error {
	if out, err := exec.Command("systemctl", "--user", "stop", systemdServiceName).CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl stop: %w\n%s", err, out)
	}
	if out, err := exec.Command("systemctl", "--user", "disable", systemdServiceName).CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl disable: %w\n%s", err, out)
	}
	path, err := m.unitPath()
	if err != nil {
		return fmt.Errorf("systemd uninstall: %w", err)
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("systemd uninstall: remove unit: %w", err)
	}
	return nil
}

// IsInstalled returns true if the unit file exists.
func (m *SystemdManager) IsInstalled() bool {
	path, err := m.unitPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}
