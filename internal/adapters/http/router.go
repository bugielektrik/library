package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	v1 "library-service/internal/adapters/http/handlers"
	httpmiddleware "library-service/internal/adapters/http/middleware"
	"library-service/internal/infrastructure/config"
	"library-service/internal/usecase"
)

// RouterConfig holds router configuration
type RouterConfig struct {
	Config   *config.Config
	Usecases *usecase.Container
	Logger   *zap.Logger
}

// NewRouter creates a new HTTP router with all routes configured
func NewRouter(cfg RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	// Base middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httpmiddleware.RequestLogger(cfg.Logger))
	r.Use(httpmiddleware.ErrorHandler(cfg.Logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.Config.App.Timeout))
	r.Use(middleware.Heartbeat("/health"))

	// Create handlers
	bookHandler := v1.NewBookHandler(
		cfg.Usecases.CreateBook,
		cfg.Usecases.GetBook,
		cfg.Usecases.ListBooks,
		cfg.Usecases.UpdateBook,
		cfg.Usecases.DeleteBook,
		cfg.Usecases.ListBookAuthors,
	)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/books", bookHandler.Routes())
		// TODO: Add author and member handlers
	})

	return r
}
