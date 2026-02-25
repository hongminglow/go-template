package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/hongminglow/go-template/internal/config"
	"github.com/hongminglow/go-template/internal/database"
	"github.com/hongminglow/go-template/internal/server"
	"github.com/hongminglow/go-template/internal/system"
	"github.com/hongminglow/go-template/internal/user"
)

func main() {
	cfg := config.FromEnv()

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbPool, err := database.NewPostgresPool(rootCtx, cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("failed to initialize postgres pool: %v", err)
	}
	defer dbPool.Close()

	if err := database.EnsureSchema(rootCtx, dbPool); err != nil {
		log.Fatalf("failed to ensure database schema: %v", err)
	}

	systemHandler := system.New(dbPool, cfg)

	userRepository := user.NewRepository(dbPool)
	userService := user.NewService(userRepository)
	userHandler := user.NewHTTPHandler(userService)

	router := server.NewRouter(systemHandler, userHandler)

	httpServer := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("http server listening on %s (env=%s)", cfg.HTTPAddr, cfg.AppEnv)
		serverErr <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server failed: %v", err)
		}
	case <-rootCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
			if closeErr := httpServer.Close(); closeErr != nil {
				log.Printf("forced close failed: %v", closeErr)
			}
		}
	}

	log.Println("server stopped")
}
