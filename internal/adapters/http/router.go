package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	_ "library-service/api/openapi" // swagger docs
	v1 "library-service/internal/adapters/http/handlers"
	httpmiddleware "library-service/internal/adapters/http/middleware"
	"library-service/internal/infrastructure/config"
	"library-service/internal/usecase"
)

// RouterConfig holds router configuration
type RouterConfig struct {
	Config       *config.Config
	Usecases     *usecase.Container
	AuthServices *usecase.AuthServices
	Logger       *zap.Logger
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

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	// Create auth middleware
	authMiddleware := httpmiddleware.NewAuthMiddleware(cfg.AuthServices.JWTService)

	// Create handlers
	authHandler := v1.NewAuthHandler(
		cfg.Usecases.RegisterMember,
		cfg.Usecases.LoginMember,
		cfg.Usecases.RefreshToken,
		cfg.Usecases.ValidateToken,
	)

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
		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.RefreshToken)

			// Protected auth routes
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.Authenticate)
				r.Get("/me", authHandler.GetCurrentMember)
			})
		})

		// Book routes (protected)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Mount("/books", bookHandler.Routes())
		})

		// TODO: Add author and member handlers
	})

	return r
}
