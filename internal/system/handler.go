package system

import (
	"context"
	"net/http"
	"time"

	"github.com/hongminglow/go-template/internal/config"
	"github.com/hongminglow/go-template/internal/httpx"
)

type dbPinger interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	db     dbPinger
	config config.Config
}

func New(db dbPinger, cfg config.Config) *Handler {
	return &Handler{
		db:     db,
		config: cfg,
	}
}

func (h *Handler) Root(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"app":     h.config.AppName,
		"env":     h.config.AppEnv,
		"message": "go backend is running",
	})
}

func (h *Handler) Hello(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "hello from /api/v1/hello",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, "database pool is not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, "database ping failed")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}
