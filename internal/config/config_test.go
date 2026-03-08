package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("HOST", "")
	t.Setenv("PORT", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Host != "0.0.0.0" {
		t.Fatalf("expected default host 0.0.0.0, got %q", cfg.Host)
	}

	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %q", cfg.Port)
	}

	if cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("expected default shutdown timeout 10s, got %s", cfg.ShutdownTimeout)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("HOST", "127.0.0.1")
	t.Setenv("PORT", "9090")
	t.Setenv("SHUTDOWN_TIMEOUT", "3s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Host != "127.0.0.1" {
		t.Fatalf("expected host override, got %q", cfg.Host)
	}

	if cfg.Port != "9090" {
		t.Fatalf("expected port override, got %q", cfg.Port)
	}

	if cfg.ShutdownTimeout != 3*time.Second {
		t.Fatalf("expected shutdown timeout override 3s, got %s", cfg.ShutdownTimeout)
	}
}
