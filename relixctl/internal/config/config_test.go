package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/relixdev/relix/relixctl/internal/config"
)

func TestDefaults(t *testing.T) {
	dir := t.TempDir()
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.RelayURL != "wss://relay.relix.sh" {
		t.Errorf("RelayURL default = %q, want %q", cfg.RelayURL, "wss://relay.relix.sh")
	}
	if cfg.MachineID == "" {
		t.Error("MachineID should be auto-generated, got empty string")
	}
	if cfg.AuthToken != "" {
		t.Errorf("AuthToken default should be empty, got %q", cfg.AuthToken)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	cfg.RelayURL = "wss://custom.relay.com"
	cfg.AuthToken = "tok_abc123"

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	cfg2, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load after save: %v", err)
	}

	if cfg2.RelayURL != "wss://custom.relay.com" {
		t.Errorf("RelayURL = %q, want %q", cfg2.RelayURL, "wss://custom.relay.com")
	}
	if cfg2.AuthToken != "tok_abc123" {
		t.Errorf("AuthToken = %q, want %q", cfg2.AuthToken, "tok_abc123")
	}
	if cfg2.MachineID != cfg.MachineID {
		t.Errorf("MachineID changed: %q -> %q", cfg.MachineID, cfg2.MachineID)
	}
}

func TestMachineIDPersistedOnFirstRun(t *testing.T) {
	dir := t.TempDir()

	cfg1, err := config.Load(dir)
	if err != nil {
		t.Fatalf("first Load: %v", err)
	}
	id1 := cfg1.MachineID

	cfg2, err := config.Load(dir)
	if err != nil {
		t.Fatalf("second Load: %v", err)
	}
	if cfg2.MachineID != id1 {
		t.Errorf("MachineID should be stable across loads: %q != %q", cfg2.MachineID, id1)
	}
}

func TestConfigDir(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "subdir")

	cfg, err := config.Load(cfgDir)
	if err != nil {
		t.Fatalf("Load with new dir: %v", err)
	}

	info, err := os.Stat(cfgDir)
	if err != nil {
		t.Fatalf("config dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("config path is not a directory")
	}
	_ = cfg
}
