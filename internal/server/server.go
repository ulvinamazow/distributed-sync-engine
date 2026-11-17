package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// wrapper
type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

// constructor for Server
func New(addr string, mux http.Handler, logger *slog.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		},
		logger: logger,
	}
}

// Start begins listening
func (s *Server) Start() error {
	s.logger.Info("HTTP server starting", "addr", s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Graceful shutdown for server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("HTTP server shutdown starting")
	return s.httpServer.Shutdown(ctx)
}
