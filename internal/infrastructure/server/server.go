package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// GitHub Copilot: refactored server implementation using options pattern,
// clearer field names, proper error handling, and structured logging.

// Option configures a Server.
type Option func(*Servers) error

// Servers manages optional HTTP and gRPC servers and their listeners.
type Servers struct {
	httpServer   *http.Server
	httpListener net.Listener

	grpcServer *grpc.Server
	grpcListen net.Listener
}

// NewServer creates a Server and applies provided options in order.
// Any option error aborts creation and is returned.
func NewServer(opts ...Option) (*Servers, error) {
	s := &Servers{}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("apply option: %w", err)
		}
	}

	return s, nil
}

// Run starts configured servers in background goroutines and returns immediately.
// It logs server runtime errors using the provided logger.
// If logger is nil, a no-op logger is used.
func (s *Servers) Run(logger *zap.Logger) error {
	if logger == nil {
		logger = zap.NewNop()
	}

	if s.httpServer == nil && s.grpcServer == nil {
		return fmt.Errorf("no servers configured to run")
	}

	// Start HTTP server if configured. Use Serve on an already-created listener
	// so option creation fails early on address conflicts.
	if s.httpServer != nil && s.httpListener != nil {
		addr := s.httpListener.Addr().String()
		go func() {
			logger.Info("starting http server", zap.String("addr", addr))
			if err := s.httpServer.Serve(s.httpListener); err != nil && err != http.ErrServerClosed {
				// http.ErrServerClosed is expected when Shutdown is called.
				logger.Error("http serve failed", zap.String("addr", addr), zap.Error(err))
			}
			logger.Info("http server stopped", zap.String("addr", addr))
		}()
	}

	// Start gRPC server if configured.
	if s.grpcServer != nil && s.grpcListen != nil {
		addr := s.grpcListen.Addr().String()
		go func() {
			logger.Info("starting grpc server", zap.String("addr", addr))
			if err := s.grpcServer.Serve(s.grpcListen); err != nil {
				// grpc.Server.Serve returns non-nil on normal stop as well, so always log with context.
				logger.Error("grpc serve failed", zap.String("addr", addr), zap.Error(err))
			}
			logger.Info("grpc server stopped", zap.String("addr", addr))
		}()
	}

	return nil
}

// Stop gracefully stops any running servers.
// HTTP server is shutdown with the provided context; gRPC receives a GracefulStop.
func (s *Servers) Stop(ctx context.Context) error {
	// Shutdown HTTP server first so it stops accepting new requests.
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("http shutdown: %w", err)
		}
	}

	// GracefulStop will stop accepting new connections and block until RPCs finish.
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	return nil
}

// WithGRPC sets up a gRPC server listening on the given address (e.g., ":50051").
// It creates the listener during configuration so address conflicts are reported early.
func WithGRPC(addr string, serverOpts ...grpc.ServerOption) Option {
	return func(s *Servers) error {
		if s.grpcServer != nil {
			return fmt.Errorf("grpc server already configured")
		}
		l, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("listen grpc %s: %w", addr, err)
		}
		s.grpcListen = l
		s.grpcServer = grpc.NewServer(serverOpts...)
		return nil
	}
}

// WithHTTP sets up an HTTP server with the provided handler listening on addr (e.g., ":8080").
// It creates the listener during configuration so address conflicts are reported early.
func WithHTTP(handler http.Handler, addr string) Option {
	return func(s *Servers) error {
		if s.httpServer != nil {
			return fmt.Errorf("http server already configured")
		}
		l, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("listen http %s: %w", addr, err)
		}
		s.httpListener = l
		s.httpServer = &http.Server{
			Handler: handler,
		}
		return nil
	}
}
