package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hongminglow/go-template/internal/system"
	"github.com/hongminglow/go-template/internal/user"
)

func NewRouter(systemHandler *system.Handler, userHandler *user.HTTPHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", systemHandler.Root)
	r.Get("/healthz", systemHandler.Health)
	r.Get("/readyz", systemHandler.Ready)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", systemHandler.Health)
		r.Get("/ready", systemHandler.Ready)
		r.Get("/hello", systemHandler.Hello)

		r.Route("/users", func(r chi.Router) {
			r.Post("/", userHandler.Create)
			r.Get("/", userHandler.List)
			r.Get("/{userID}", userHandler.GetByID)
			r.Put("/{userID}", userHandler.Update)
			r.Delete("/{userID}", userHandler.Delete)
		})
	})

	return r
}
