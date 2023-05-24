package server

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

type Server struct {
	http *http.Server
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
	go func() {
		if err = s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("ERR_INIT_REST", zap.Error(err))
			return
		}
	}()
	logger.Info("server started on http://localhost" + s.http.Addr)

	return
}

func (s *Server) Stop(ctx context.Context) (err error) {
	if err = s.http.Shutdown(ctx); err != nil {
		return
	}

	return
}

func WithHTTPServer(handler http.Handler, port string) Configuration {
	return func(r *Server) (err error) {
		r.http = &http.Server{
			Handler: handler,
			Addr:    ":" + port,
		}
		return
	}
}
