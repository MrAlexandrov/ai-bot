package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimid "github.com/go-chi/chi/v5/middleware"

	"github.com/mrralexandrov/ai-bot/backend/config"
	"github.com/mrralexandrov/ai-bot/backend/handlers"
	"github.com/mrralexandrov/ai-bot/backend/ollama"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg := config.Load()
	client := ollama.NewClient(cfg.OllamaURL)
	chat := handlers.NewChatHandler(client, cfg.OllamaModel)

	r := chi.NewRouter()
	r.Use(chimid.Logger)
	r.Use(chimid.Recoverer)

	r.Get("/health", handlers.Health)
	r.Post("/chat", chat.Chat)
	r.Get("/models", chat.Models)

	slog.Info("backend started", "port", cfg.Port, "model", cfg.OllamaModel)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}
