package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config aggregates application-level configuration sourced from environment variables.
type Config struct {
	AppName              string
	HTTPPort             string
	GinMode              string
	AllowNegativeBalance bool
	Database             DatabaseConfig
}

// DatabaseConfig captures connection-related settings for the relational database.
type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// Load reads configuration from the environment, optionally sourcing a local .env file.
func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = os.Getenv("PORT")
	}
	if httpPort == "" {
		httpPort = "8080"
	}

	cfg := Config{
		AppName:              getEnv("APP_NAME", "backend-lab3"),
		HTTPPort:             httpPort,
		GinMode:              getEnv("GIN_MODE", "debug"),
		AllowNegativeBalance: getEnvBool("ALLOW_NEGATIVE_BALANCE", false),
	}

	if cfg.HTTPPort == "" {
		return Config{}, errors.New("HTTP_PORT must not be empty")
	}

	dbCfg, err := loadDatabaseConfig()
	if err != nil {
		return Config{}, err
	}
	cfg.Database = dbCfg

	return cfg, nil
}

func loadDatabaseConfig() (DatabaseConfig, error) {
	dsn := getEnv("DATABASE_DSN", "")
	if dsn == "" {
		dsn = getEnv("DATABASE_URL", "")
	}
	if dsn == "" {
		dsn = "postgres://backend:backend@localhost:5432/backend_lab3?sslmode=disable"
	}
	if dsn == "" {
		return DatabaseConfig{}, errors.New("DATABASE_DSN or DATABASE_URL must not be empty")
	}

	maxOpen, err := getEnvInt("DB_MAX_OPEN_CONNS", 25)
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("parse DB_MAX_OPEN_CONNS: %w", err)
	}

	maxIdle, err := getEnvInt("DB_MAX_IDLE_CONNS", 25)
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("parse DB_MAX_IDLE_CONNS: %w", err)
	}

	connLifetime, err := getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute)
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("parse DB_CONN_MAX_LIFETIME: %w", err)
	}

	return DatabaseConfig{
		DSN:             dsn,
		MaxOpenConns:    maxOpen,
		MaxIdleConns:    maxIdle,
		ConnMaxLifetime: connLifetime,
	}, nil
}

func getEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) (int, error) {
	valueStr, ok := os.LookupEnv(key)
	if !ok || valueStr == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	valueStr, ok := os.LookupEnv(key)
	if !ok || valueStr == "" {
		return fallback, nil
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func getEnvBool(key string, fallback bool) bool {
	valueStr, ok := os.LookupEnv(key)
	if !ok || valueStr == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(valueStr)
	if err != nil {
		return fallback
	}
	return parsed
}
