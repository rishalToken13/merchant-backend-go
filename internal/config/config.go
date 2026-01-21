package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppEnv   string
	HTTPPort int

	DBDSN string

	RabbitURL      string
	RabbitExchange string

	JWTSecret string
	JWTIssuer string

	TronAPIBase string
	TronAPIKey  string
}

func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:   getEnv("APP_ENV", "dev"),
		HTTPPort: getEnvInt("HTTP_PORT", 8080),
		DBDSN:    os.Getenv("DB_DSN"),

		RabbitURL:      os.Getenv("RABBIT_URL"),
		RabbitExchange: getEnv("RABBIT_EXCHANGE", "token13.events"),

		JWTSecret: os.Getenv("JWT_SECRET"),
		JWTIssuer: getEnv("JWT_ISSUER", "merchant-backend"),

		TronAPIBase: getEnv("TRON_API_BASE", "https://api.trongrid.io"),
		TronAPIKey:  os.Getenv("TRON_API_KEY"),
	}

	// Minimal validation
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.DBDSN == "" {
		return nil, fmt.Errorf("DB_DSN is required")
	}
	if cfg.RabbitURL == "" {
		return nil, fmt.Errorf("RABBIT_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
