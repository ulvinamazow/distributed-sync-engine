package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/config"
	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/handler"
	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/repository/postgres"
	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/server"
	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/service"
	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/sync"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("configuration was not loaded", "err", err)
		os.Exit(1)
	}

	dbCxt, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	pool, err := pgxpool.New(dbCxt, cfg.DatabaseURL)
	cancel()
	if err != nil {
		logger.Error("DB pool nor created", "err", err)
		os.Exit(1)
	}

	defer pool.Close()

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = pool.Ping(pingCtx)
	cancel()
	if err != nil {
		logger.Error("DB unreachable", "err", err)
		os.Exit(1)
	}

	itemRepo := postgres.NewItemRepo(pool)
	syncSvc := sync.New(itemRepo, cfg.PeerURLs, logger)
	itemService := service.NewItemService(itemRepo)
	itemHandler := handler.NewItemHandler(itemService)
	syncHandler := handler.NewSyncHandler(syncSvc)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.Health(cfg.NodeID))

	logger.Info("DB connection created")

	logger.Info("node starting", "node_id", cfg.NodeID, "addr", cfg.HTTPAddr)

	// blocking
	mux = http.NewServeMux()

	mux.HandleFunc("GET /health", handler.Health(cfg.NodeID))
	mux.HandleFunc("POST /items", itemHandler.Create)
	mux.HandleFunc("GET /items", itemHandler.List)
	mux.HandleFunc("GET /items/{id}", itemHandler.Get)
	mux.HandleFunc("PUT /items/{id}", itemHandler.Update)
	mux.HandleFunc("DELETE /items/{id}", itemHandler.Delete)

	mux.HandleFunc("GET /internal/items", itemHandler.List)
	mux.HandleFunc("POST /sync/pull", syncHandler.Pull)

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
