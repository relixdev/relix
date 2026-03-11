package config_test

import (
	"testing"
	"time"

	"github.com/relixdev/relix/relay/internal/config"
)

func TestDefaults(t *testing.T) {
	t.Setenv("RELAY_JWT_SECRET", "test-secret")
	t.Setenv("RELAY_PORT", "")
	t.Setenv("RELAY_BUFFER_MAX_MESSAGES", "")
	t.Setenv("RELAY_BUFFER_MAX_BYTES", "")
	t.Setenv("RELAY_BUFFER_TTL", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("want Port=8080, got %d", cfg.Port)
	}
	if cfg.JWTSecret != "test-secret" {
		t.Errorf("want JWTSecret=test-secret, got %q", cfg.JWTSecret)
	}
	if cfg.BufferMaxMessages != 1000 {
		t.Errorf("want BufferMaxMessages=1000, got %d", cfg.BufferMaxMessages)
	}
	if cfg.BufferMaxBytes != 10*1024*1024 {
		t.Errorf("want BufferMaxBytes=10MB, got %d", cfg.BufferMaxBytes)
	}
	if cfg.BufferTTL != 24*time.Hour {
		t.Errorf("want BufferTTL=24h, got %v", cfg.BufferTTL)
	}
}

func TestEnvOverrides(t *testing.T) {
	t.Setenv("RELAY_PORT", "9090")
	t.Setenv("RELAY_JWT_SECRET", "my-secret")
	t.Setenv("RELAY_BUFFER_MAX_MESSAGES", "500")
	t.Setenv("RELAY_BUFFER_MAX_BYTES", "5242880")
	t.Setenv("RELAY_BUFFER_TTL", "12h")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 9090 {
		t.Errorf("want Port=9090, got %d", cfg.Port)
	}
	if cfg.JWTSecret != "my-secret" {
		t.Errorf("want JWTSecret=my-secret, got %q", cfg.JWTSecret)
	}
	if cfg.BufferMaxMessages != 500 {
		t.Errorf("want BufferMaxMessages=500, got %d", cfg.BufferMaxMessages)
	}
	if cfg.BufferMaxBytes != 5242880 {
		t.Errorf("want BufferMaxBytes=5242880, got %d", cfg.BufferMaxBytes)
	}
	if cfg.BufferTTL != 12*time.Hour {
		t.Errorf("want BufferTTL=12h, got %v", cfg.BufferTTL)
	}
}

func TestMissingJWTSecretReturnsError(t *testing.T) {
	t.Setenv("RELAY_JWT_SECRET", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing JWT secret, got nil")
	}
}
