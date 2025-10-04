package grpc

import (
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Server represents a gRPC server
type Server struct {
	grpcServer *grpc.Server
	logger     *zap.Logger
	port       string
}

// NewServer creates a new gRPC server
func NewServer(port string, logger *zap.Logger) *Server {
	return &Server{
		grpcServer: grpc.NewServer(),
		logger:     logger,
		port:       port,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.logger.Info("starting gRPC server", zap.String("port", s.port))

	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	s.logger.Info("stopping gRPC server")
	s.grpcServer.GracefulStop()
}
