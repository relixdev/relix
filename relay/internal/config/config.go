package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

// Config holds all relay server configuration.
type Config struct {
	Port              int
	JWTSecret         string
	BufferMaxMessages int
	BufferMaxBytes    int64
	BufferTTL         time.Duration
}

// Load reads configuration from environment variables, applying defaults.
// Returns an error if RELAY_JWT_SECRET is not set.
func Load() (*Config, error) {
	cfg := &Config{
		Port:              8080,
		BufferMaxMessages: 1000,
		BufferMaxBytes:    10 * 1024 * 1024, // 10 MB
		BufferTTL:         24 * time.Hour,
	}

	if v := os.Getenv("RELAY_PORT"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, errors.New("RELAY_PORT: " + err.Error())
		}
		cfg.Port = n
	}

	cfg.JWTSecret = os.Getenv("RELAY_JWT_SECRET")
	if cfg.JWTSecret == "" {
		return nil, errors.New("RELAY_JWT_SECRET is required")
	}

	if v := os.Getenv("RELAY_BUFFER_MAX_MESSAGES"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, errors.New("RELAY_BUFFER_MAX_MESSAGES: " + err.Error())
		}
		cfg.BufferMaxMessages = n
	}

	if v := os.Getenv("RELAY_BUFFER_MAX_BYTES"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, errors.New("RELAY_BUFFER_MAX_BYTES: " + err.Error())
		}
		cfg.BufferMaxBytes = n
	}

	if v := os.Getenv("RELAY_BUFFER_TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, errors.New("RELAY_BUFFER_TTL: " + err.Error())
		}
		cfg.BufferTTL = d
	}

	return cfg, nil
}
