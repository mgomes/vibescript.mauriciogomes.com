package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Host            string
	Port            string
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Host:            getenv("HOST", "0.0.0.0"),
		Port:            getenv("PORT", "8080"),
		ShutdownTimeout: 10 * time.Second,
	}

	if value := os.Getenv("SHUTDOWN_TIMEOUT"); value != "" {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return Config{}, fmt.Errorf("parse SHUTDOWN_TIMEOUT: %w", err)
		}
		cfg.ShutdownTimeout = duration
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
