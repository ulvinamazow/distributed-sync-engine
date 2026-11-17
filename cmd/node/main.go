package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/config"
	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/handler"
	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("configuration was not loaded", "err", err)
		os.Exit(1)
	}

	logger.Info("node starting", "node_id", cfg.NodeID, "addr", cfg.HTTPAddr)

	// blocking
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.Health(cfg.NodeID))

	srv := server.New(cfg.HTTPAddr, mux, logger)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errCh <- err
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()

	select {
	case err := <-errCh:
		logger.Error("server error", "err", err)
		os.Exit(1)
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)

	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "err", err)
		os.Exit(1)
	}
	logger.Info("node stopped")
}
