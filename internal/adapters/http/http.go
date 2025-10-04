package http

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"library-service/internal/infrastructure/config"
	"library-service/internal/usecase"
)

// Server represents an HTTP server
type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(cfg *config.Config, usecases *usecase.Container, logger *zap.Logger) (*Server, error) {
	// Create router
	router := NewRouter(RouterConfig{
		Config:   cfg,
		Usecases: usecases,
		Logger:   logger,
	})

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    cfg.App.Port,
		Handler: router,
	}

	return &Server{
		httpServer: httpServer,
		logger:     logger,
	}, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	go func() {
		s.logger.Info("starting HTTP server", zap.String("addr", s.httpServer.Addr))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", zap.Error(err))
		}
	}()
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("HTTP server shutdown error: %w", err)
	}
	s.logger.Info("HTTP server stopped")
	return nil
}
