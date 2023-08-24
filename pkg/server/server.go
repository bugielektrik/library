package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	http     *http.Server
	grpc     *grpc.Server
	listener net.Listener
}

// Configuration is an alias for a function that will take in a pointer to a Repository and modify it
type Configuration func(r *Server) error

// New takes a variable amount of Configuration functions and returns a new Server
// Each Configuration will be called in the order they are passed in
func New(configs ...Configuration) (r *Server, err error) {
	// Create the Server
	r = &Server{}

	// Apply all Configurations passed in
	for _, cfg := range configs {
		// Pass the service into the configuration function
		if err = cfg(r); err != nil {
			return
		}
	}
	return
}

func (s *Server) Run(logger *zap.Logger) (err error) {
	if s.http != nil {
		go func() {
			if err = s.http.ListenAndServe(); err != nil {
				logger.Error("ERR_SERVE_HTTP", zap.Error(err))
				return
			}
		}()
	}

	if s.grpc != nil {
		go func() {
			if err = s.grpc.Serve(s.listener); err != nil {
				logger.Error("ERR_SERVE_GRPC", zap.Error(err))
				return
			}
		}()
	}

	return
}

func (s *Server) Stop(ctx context.Context) (err error) {
	if s.http != nil {
		if err = s.http.Shutdown(ctx); err != nil {
			return
		}
	}

	if s.grpc != nil {
		s.grpc.GracefulStop()
	}

	return
}

func WithGRPCServer(port string) Configuration {
	return func(s *Server) (err error) {
		s.listener, err = net.Listen("tcp", fmt.Sprintf("localhost:%s", port))
		if err != nil {
			return
		}
		s.grpc = &grpc.Server{}

		return
	}
}

func WithHTTPServer(handler http.Handler, port string) Configuration {
	return func(s *Server) (err error) {
		s.http = &http.Server{
			Handler: handler,
			Addr:    ":" + port,
		}
		return
	}
}
