package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const launchdLabel = "com.relix.agent"
const launchdPlistName = launchdLabel + ".plist"

// GeneratePlist returns the XML content of a launchd plist for the given
// binary path. The plist configures the agent to run at load with the
// daemon-run subcommand.
func GeneratePlist(binaryPath string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>daemon-run</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>StandardOutPath</key>
	<string>~/Library/Logs/relix-agent.log</string>
	<key>StandardErrorPath</key>
	<string>~/Library/Logs/relix-agent.log</string>
</dict>
</plist>
`, launchdLabel, binaryPath)
}

// LaunchdManager implements ServiceManager using macOS launchctl.
type LaunchdManager struct {
	plistDir string // override for testing; empty uses ~/Library/LaunchAgents
}

// NewLaunchdManager creates a LaunchdManager targeting the standard
// ~/Library/LaunchAgents directory.
func NewLaunchdManager() *LaunchdManager {
	return &LaunchdManager{}
}

func (m *LaunchdManager) plistPath() (string, error) {
	if m.plistDir != "" {
		return filepath.Join(m.plistDir, launchdPlistName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", launchdPlistName), nil
}

// Install writes the plist and loads it via launchctl.
func (m *LaunchdManager) Install(binaryPath string) error {
	path, err := m.plistPath()
	if err != nil {
		return fmt.Errorf("launchd install: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("launchd install: create dir: %w", err)
	}
	content := GeneratePlist(binaryPath)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("launchd install: write plist: %w", err)
	}
	if out, err := exec.Command("launchctl", "load", path).CombinedOutput(); err != nil {
		return fmt.Errorf("launchctl load: %w\n%s", err, out)
	}
	return nil
}

// Uninstall unloads the service and removes the plist.
func (m *LaunchdManager) Uninstall() error {
	path, err := m.plistPath()
	if err != nil {
		return fmt.Errorf("launchd uninstall: %w", err)
	}
	if out, err := exec.Command("launchctl", "unload", path).CombinedOutput(); err != nil {
		return fmt.Errorf("launchctl unload: %w\n%s", err, out)
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("launchd uninstall: remove plist: %w", err)
	}
	return nil
}

// IsInstalled returns true if the plist file exists.
func (m *LaunchdManager) IsInstalled() bool {
	path, err := m.plistPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}
