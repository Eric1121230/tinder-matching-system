package main

import (
	"example.com/tinder/internal/config"
	"example.com/tinder/internal/container"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	cfg := config.NewConfig()
	if envPort := os.Getenv("PORT"); envPort != "" {
		cfg.Port = envPort
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	c := container.New(logger)

	addr := ":" + cfg.Port
	logger.Info("starting tinder server", "addr", addr)
	if err := http.ListenAndServe(addr, c.Handler()); err != nil {
		logger.Error("server stopped", "error", err.Error())
		os.Exit(1)
	}
}
