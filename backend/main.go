package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimid "github.com/go-chi/chi/v5/middleware"

	"github.com/mrralexandrov/ai-bot/backend/config"
	"github.com/mrralexandrov/ai-bot/backend/handlers"
	"github.com/mrralexandrov/ai-bot/backend/ollama"
	"github.com/mrralexandrov/ai-bot/logx"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	logx.Setup("ai-bot-backend")
	shutdown := logx.Init(context.Background(), "ai-bot-backend")
	defer shutdown(context.Background())

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
	if err := http.ListenAndServe(":"+cfg.Port, otelhttp.NewHandler(r, "http.server")); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}
