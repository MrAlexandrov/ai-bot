package config

import "os"

type Config struct {
	Port        string
	OllamaURL   string
	OllamaModel string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		OllamaURL:   getEnv("OLLAMA_URL", "http://ollama:11434"),
		OllamaModel: getEnv("OLLAMA_MODEL", "qwen2.5:7b"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
