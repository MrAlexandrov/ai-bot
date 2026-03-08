package config

import (
	"fmt"
	"os"
)

type Config struct {
	TelegramToken string
	BackendURL    string
}

func Load() (*Config, error) {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_TOKEN is required")
	}
	return &Config{
		TelegramToken: token,
		BackendURL:    getEnv("BACKEND_URL", "http://backend:8080"),
	}, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
