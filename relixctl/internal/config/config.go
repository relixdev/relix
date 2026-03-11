package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const configFile = "config.json"

// Config holds relixctl runtime configuration.
type Config struct {
	RelayURL  string `json:"relay_url"`
	MachineID string `json:"machine_id"`
	AuthToken string `json:"auth_token"`

	dir string // directory where config.json lives
}

// ConfigDir returns the default ~/.relixctl/ directory, creating it if needed.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".relixctl")
	return dir, os.MkdirAll(dir, 0700)
}

// Load reads config from dir/config.json. If the file doesn't exist it
// returns a default Config with an auto-generated MachineID and saves it.
func Load(dir string) (*Config, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	path := filepath.Join(dir, configFile)
	data, err := os.ReadFile(path)

	var cfg Config
	cfg.dir = dir

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		// First run: apply defaults and persist.
		cfg.RelayURL = "wss://relay.relix.sh"
		cfg.MachineID = uuid.New().String()
		if saveErr := cfg.Save(); saveErr != nil {
			return nil, saveErr
		}
		return &cfg, nil
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	cfg.dir = dir

	// Upgrade: generate MachineID if missing from an older config.
	if cfg.MachineID == "" {
		cfg.MachineID = uuid.New().String()
		if err := cfg.Save(); err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}

// Save writes the current config back to disk.
func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(c.dir, configFile)
	return os.WriteFile(path, data, 0600)
}

// Dir returns the directory this config was loaded from.
func (c *Config) Dir() string {
	return c.dir
}
