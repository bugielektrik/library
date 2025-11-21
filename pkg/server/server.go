package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Option func(*Servers) error

type Servers struct {
	httpServer   *http.Server
	httpListener net.Listener

	grpcServer *grpc.Server
	grpcListen net.Listener
}

func NewServer(opts ...Option) (*Servers, error) {
	s := &Servers{}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("apply option: %w", err)
		}
	}

	return s, nil
}

func (s *Servers) Run(logger *zap.Logger) error {
	if logger == nil {
		logger = zap.NewNop()
	}

	if s.httpServer == nil && s.grpcServer == nil {
		return fmt.Errorf("no servers configured to run")
	}

	if s.httpServer != nil && s.httpListener != nil {
		addr := s.httpListener.Addr().String()
		go func() {
			logger.Info("starting http server", zap.String("addr", addr))
			if err := s.httpServer.Serve(s.httpListener); err != nil && err != http.ErrServerClosed {
				logger.Error("http serve failed", zap.String("addr", addr), zap.Error(err))
			}
			logger.Info("http server stopped", zap.String("addr", addr))
		}()
	}

	if s.grpcServer != nil && s.grpcListen != nil {
		addr := s.grpcListen.Addr().String()
		go func() {
			logger.Info("starting grpc server", zap.String("addr", addr))
			if err := s.grpcServer.Serve(s.grpcListen); err != nil {
				logger.Error("grpc serve failed", zap.String("addr", addr), zap.Error(err))
			}
			logger.Info("grpc server stopped", zap.String("addr", addr))
		}()
	}

	return nil
}

func (s *Servers) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("http shutdown: %w", err)
		}
	}

	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	return nil
}

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
