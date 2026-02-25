package handler

import (
	"net/http"
	"time"
)

func (h *Handler) Root(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"app":     h.Config.AppName,
		"env":     h.Config.AppEnv,
		"message": "go backend is running",
	})
}

func (h *Handler) Hello(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "hello from /api/v1/hello",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}
