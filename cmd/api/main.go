package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hongminglow/go-template/internal/auth"
	"github.com/hongminglow/go-template/internal/config"
	"github.com/hongminglow/go-template/internal/infrastructure/cache"
	"github.com/hongminglow/go-template/internal/infrastructure/postgres"
	"github.com/hongminglow/go-template/internal/pkg/logger"
	"github.com/hongminglow/go-template/internal/server"
	"github.com/hongminglow/go-template/internal/system"
	"github.com/hongminglow/go-template/internal/user"
)

// @title           Go Template API
// @version         1.0
// @description     Enterprise standard Go template backend API.
// @termsOfService  http://swagger.io/terms/
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @host      localhost:8080
// @BasePath  /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	cfg := config.FromEnv()

	// Initialize structured logger
	_ = logger.Init(cfg.AppEnv)
	slog.Info("starting application", "env", cfg.AppEnv, "addr", cfg.HTTPAddr)

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbPool, err := postgres.NewPostgresPool(rootCtx, cfg.PostgresDSN())
	if err != nil {
		slog.Error("failed to initialize postgres pool", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	if err := postgres.EnsureSchema(cfg.PostgresDSN()); err != nil {
		slog.Error("failed to ensure database schema", "error", err)
		os.Exit(1)
	}

	if err := postgres.SeedDefaultData(rootCtx, dbPool, postgres.SeedConfig{
		AdminName:  cfg.SeedAdminName,
		AdminEmail: cfg.SeedAdminEmail,
	}); err != nil {
		slog.Error("failed to seed default data", "error", err)
		os.Exit(1)
	}

	redisPool, err := cache.NewRedisClient(rootCtx, cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword)
	if err != nil {
		slog.Error("failed to initialize redis pool", "error", err)
		os.Exit(1)
	}
	defer redisPool.Close()

	systemHandler := system.New(dbPool, cfg)

	userRepository := user.NewRepository(dbPool)
	userService := user.NewService(userRepository, redisPool)
	userHandler := user.NewHTTPHandler(userService)
	authHandler := auth.NewHTTPHandler(userService, cfg.JWTSecret)

	router := server.NewRouter(systemHandler, userHandler, authHandler, cfg, redisPool)

	httpServer := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("http server listening", "addr", cfg.HTTPAddr, "env", cfg.AppEnv)
		serverErr <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			slog.Error("http server failed", "error", err)
			os.Exit(1)
		}
	case <-rootCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("graceful shutdown failed", "error", err)
			if closeErr := httpServer.Close(); closeErr != nil {
				slog.Error("forced close failed", "error", closeErr)
			}
		}
	}

	slog.Info("server stopped")
}

