package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/tinder/internal/config"
	"example.com/tinder/internal/container"
)

func main() {
	cfg := config.NewConfig()
	if envPort := os.Getenv("PORT"); envPort != "" {
		cfg.Port = envPort
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	c := container.New(logger)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: c.Handler(),
	}

	go func() {
		logger.Info("starting tinder server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}

	logger.Info("server exited gracefully")
}
