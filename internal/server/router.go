package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hongminglow/go-template/internal/auth"
	"github.com/hongminglow/go-template/internal/config"
	"github.com/hongminglow/go-template/internal/pkg/httpx"
	"github.com/hongminglow/go-template/internal/system"
	"github.com/hongminglow/go-template/internal/user"
	"github.com/go-chi/cors"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	_ "github.com/hongminglow/go-template/docs"
)

func NewRouter(systemHandler *system.Handler, userHandler *user.HTTPHandler, authHandler *auth.HTTPHandler, cfg config.Config, redisClient *redis.Client) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httpx.LoggerMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Global Rate Limiter
	// Token Bucket algorithm restricts burst traffic from a single IP.
	// - RequestsPerSecond: 10 (Replenishes 10 tokens per second).
	// - BurstSize: 30 (Allows up to 30 immediate requests simultaneously before rate limiting).
	// This helps block abuse/DDoS while keeping normal user interactions completely unrestrained.
	r.Use(httpx.TokenBucketRateLimiter(redisClient, httpx.RateLimiterConfig{
		RequestsPerSecond: 10,
		BurstSize:         30,
	}))

	r.Get("/", systemHandler.Root)
	r.Get("/healthz", systemHandler.Health)
	r.Get("/readyz", systemHandler.Ready)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", systemHandler.Health)
		r.Get("/ready", systemHandler.Ready)
		r.Get("/hello", systemHandler.Hello)

		// Swagger Documentation
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/api/v1/swagger/doc.json"), //The url pointing to API definition
		))

		// Public Routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/register", authHandler.Register)
		})

		// Protected Routes
		r.Group(func(r chi.Router) {
			r.Use(httpx.JWTMiddleware(cfg.JWTSecret))

			r.Route("/users", func(r chi.Router) {
				r.Get("/me", userHandler.GetMe)
				r.Get("/", userHandler.List)
				r.Get("/{userID}", userHandler.GetByID)
				r.Put("/{userID}", userHandler.Update)
				r.Delete("/{userID}", userHandler.Delete)
			})
		})
	})

	return r
}
