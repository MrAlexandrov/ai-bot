package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/mrralexandrov/ai-bot/backend/ollama"
)

type ChatHandler struct {
	ollama *ollama.Client
	model  string
}

func NewChatHandler(client *ollama.Client, model string) *ChatHandler {
	return &ChatHandler{ollama: client, model: model}
}

type chatRequest struct {
	Message string           `json:"message"`
	History []ollama.Message `json:"history,omitempty"`
}

type chatResponse struct {
	Reply string `json:"reply"`
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *ChatHandler) Chat(w http.ResponseWriter, r *http.Request) {
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	messages := append(req.History, ollama.Message{Role: "user", Content: req.Message})

	resp, err := h.ollama.Chat(r.Context(), h.model, messages)
	if err != nil {
		slog.ErrorContext(r.Context(), "ollama chat", "err", err)
		http.Error(w, "model unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chatResponse{Reply: resp.Message.Content})
}

func (h *ChatHandler) Models(w http.ResponseWriter, r *http.Request) {
	models, err := h.ollama.ListModels(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "list models", "err", err)
		http.Error(w, "model unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}
