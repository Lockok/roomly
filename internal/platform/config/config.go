package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	HTTPPort    string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env file: %w", err)
	}

	jwtTTL, err := time.ParseDuration(getEnv("JWT_TTL", "24h"))
	if err != nil {
		return Config{}, fmt.Errorf("parse JWT_TTL: %w", err)
	}

	if jwtTTL <= 0 {
		return Config{}, fmt.Errorf("JWT_TTL must be greater than zero")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	cfg := Config{
		AppEnv:      getEnv("APP_ENV", "local"),
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   jwtSecret,
		JWTTTL:      jwtTTL,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
