package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config aggregates application-level configuration sourced from environment variables.
type Config struct {
	AppName  string
	HTTPPort string
	GinMode  string
}

// Load reads configuration from the environment, optionally sourcing a local .env file.
func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	cfg := Config{
		AppName:  getEnv("APP_NAME", "backend-lab3"),
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		GinMode:  getEnv("GIN_MODE", "debug"),
	}

	if cfg.HTTPPort == "" {
		return Config{}, errors.New("HTTP_PORT must not be empty")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}
	return value
}
