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
	"github.com/hongminglow/go-template/internal/handler"
	"github.com/hongminglow/go-template/internal/server"
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

	h := handler.New(dbPool, cfg)
	router := server.NewRouter(h)

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
