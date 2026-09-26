package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	tele "gopkg.in/telebot.v3"
)

const updateContextKey = "otel-update-context"

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type backendRequest struct {
	Message string    `json:"message"`
	History []message `json:"history,omitempty"`
}

type backendResponse struct {
	Reply string `json:"reply"`
}

type Handler struct {
	backendURL string
	httpClient *http.Client
	rootID     int64

	mu       sync.RWMutex
	sessions map[int64][]message // per-user conversation history
}

func New(backendURL string, rootID int64) *Handler {
	return &Handler{
		backendURL: backendURL,
		httpClient: &http.Client{Timeout: 10 * time.Minute, Transport: otelhttp.NewTransport(http.DefaultTransport)},
		rootID:     rootID,
		sessions:   make(map[int64][]message),
	}
}

// Auth — middleware, пропускает только ROOT_ID.
func (h *Handler) Auth(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		ctx, span := otel.Tracer("ai-bot").Start(context.Background(), "telegram.update", trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()
		c.Set(updateContextKey, ctx)
		defer c.Set(updateContextKey, nil)
		if c.Sender().ID != h.rootID {
			slog.WarnContext(updateContext(c), "unauthorized access", "user_id", c.Sender().ID, "username", c.Sender().Username)
			return nil // молча игнорируем
		}
		return next(c)
	}
}

func (h *Handler) Start(c tele.Context) error {
	return c.Send(
		"Привет! Я AI-ассистент на локальной модели 🦙\n\n" +
			"Просто напиши мне — и я отвечу.\n\n" +
			"/reset — сбросить историю диалога\n" +
			"/help  — помощь",
	)
}

func (h *Handler) Help(c tele.Context) error {
	return c.Send(
		"Как пользоваться:\n\n" +
			"• Напиши любое сообщение — я отвечу с помощью локальной AI-модели\n" +
			"• Я помню контекст разговора (до 20 сообщений)\n" +
			"• /reset — начать диалог заново",
	)
}

func (h *Handler) Reset(c tele.Context) error {
	h.mu.Lock()
	delete(h.sessions, c.Sender().ID)
	h.mu.Unlock()
	return c.Send("Контекст сброшен. Начнём заново!")
}

func (h *Handler) Message(c tele.Context) error {
	userID := c.Sender().ID
	text := c.Text()

	h.mu.RLock()
	hist := make([]message, len(h.sessions[userID]))
	copy(hist, h.sessions[userID])
	h.mu.RUnlock()

	_ = c.Notify(tele.Typing)

	ctx, cancel := context.WithTimeout(updateContext(c), 10*time.Minute)
	defer cancel()

	reply, err := h.callBackend(ctx, text, hist)
	if err != nil {
		slog.ErrorContext(updateContext(c), "backend call failed", "err", err, "user_id", userID)
		return c.Send("Не удалось получить ответ от модели. Попробуй позже.")
	}

	h.mu.Lock()
	h.sessions[userID] = append(h.sessions[userID],
		message{Role: "user", Content: text},
		message{Role: "assistant", Content: reply},
	)
	// Keep last 20 messages (10 exchanges)
	if len(h.sessions[userID]) > 20 {
		h.sessions[userID] = h.sessions[userID][len(h.sessions[userID])-20:]
	}
	h.mu.Unlock()

	return c.Send(reply)
}

func updateContext(c tele.Context) context.Context {
	if ctx, ok := c.Get(updateContextKey).(context.Context); ok {
		return ctx
	}
	return context.Background()
}

func (h *Handler) callBackend(ctx context.Context, text string, history []message) (string, error) {
	body, err := json.Marshal(backendRequest{Message: text, History: history})
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.backendURL+"/chat", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("backend status %d", resp.StatusCode)
	}

	var out backendResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	return out.Reply, nil
}
