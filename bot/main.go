package main

import (
	"log/slog"
	"os"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/mrralexandrov/ai-bot/bot/config"
	"github.com/mrralexandrov/ai-bot/bot/handlers"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	bot, err := tele.NewBot(tele.Settings{
		Token:  cfg.TelegramToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		slog.Error("bot init", "err", err)
		os.Exit(1)
	}

	h := handlers.New(cfg.BackendURL)

	bot.Handle("/start", h.Start)
	bot.Handle("/help", h.Help)
	bot.Handle("/reset", h.Reset)
	bot.Handle(tele.OnText, h.Message)

	slog.Info("telegram bot started")
	bot.Start()
}
