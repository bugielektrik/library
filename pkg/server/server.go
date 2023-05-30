package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"go.uber.org/zap"

	"google.golang.org/grpc"
)

type Dependencies struct {
	HTTPHandler http.Handler
	HTTPPort    string
}

type Server struct {
	dependencies Dependencies

	http     *http.Server
	grpc     *grpc.Server
	listener net.Listener
}

// Configuration is an alias for a function that will take in a pointer to a Repository and modify it
type Configuration func(r *Server) error

// New takes a variable amount of Configuration functions and returns a new Server
// Each Configuration will be called in the order they are passed in
func New(d Dependencies, configs ...Configuration) (r *Server, err error) {
	// Create the Server
	r = &Server{
		dependencies: d,
	}

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
			if err = s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("ERR_INIT_REST", zap.Error(err))
				return
			}
		}()
		logger.Info("http server started on http://localhost" + s.http.Addr)
	}

	if s.grpc != nil {
		if err = s.grpc.Serve(s.listener); err != nil {
			return
		}
		logger.Info("grpc server started on http://localhost" + s.listener.Addr().String())
	}

	return
}

func (s *Server) Stop(ctx context.Context) (err error) {
	if s.http != nil {
		if err = s.http.Shutdown(ctx); err != nil {
			return
		}
	}

	return
}

func WithGRPCServer() Configuration {
	return func(s *Server) (err error) {
		s.listener, err = net.Listen("tcp", fmt.Sprintf("localhost:%s", s.dependencies.HTTPPort))
		if err != nil {
			return
		}
		s.grpc = &grpc.Server{}

		return
	}
}

func WithHTTPServer() Configuration {
	return func(s *Server) (err error) {
		s.http = &http.Server{
			Handler: s.dependencies.HTTPHandler,
			Addr:    ":" + s.dependencies.HTTPPort,
		}
		return
	}
}
