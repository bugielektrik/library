package router

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
)

// New creates a router with a default application logger.
// Kept for convenience and backwards compatibility.
func New(patterns []string) *chi.Mux {
	// Init a new router instance
	r := chi.NewRouter()

	// Logger middleware with path skipping
	r.Use(LoggerWithSkips(patterns))

	// Common, lightweight middlewares applied in a clear order.
	r.Use(middleware.RequestID)                          // attach a request id to the context
	r.Use(middleware.RealIP)                             // get real IP from headers
	r.Use(middleware.Recoverer)                          // recover from panics
	r.Use(middleware.CleanPath)                          // sanitize URL path
	r.Use(middleware.Heartbeat("/"))                     // simple health check endpoint
	r.Use(middleware.Timeout(60 * time.Second))          // set request timeout
	r.Use(render.SetContentType(render.ContentTypeJSON)) // always return JSON content type

	// CORS settings - kept permissive per original code but centralized here.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	return r
}
