package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	TelegramToken    string
	TelegramLocalAPI string
	BackendURL       string
	RootID           int64
}

func Load() (*Config, error) {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_TOKEN is required")
	}

	rootRaw := os.Getenv("ROOT_ID")
	if rootRaw == "" {
		return nil, fmt.Errorf("ROOT_ID is required")
	}
	rootID, err := strconv.ParseInt(rootRaw, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("ROOT_ID must be a number: %w", err)
	}

	return &Config{
		TelegramToken:    token,
		TelegramLocalAPI: getEnv("TELEGRAM_LOCAL_API_URL", ""),
		BackendURL:       getEnv("BACKEND_URL", "http://backend:8080"),
		RootID:           rootID,
	}, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
