package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	// _ "library-service/api/openapi" // swagger docs - comment out if not generated
	bookhttp "library-service/internal/books/handler"
	authorhttp "library-service/internal/books/handler/author"
	"library-service/internal/container"
	"library-service/internal/infrastructure/config"
	memberauth "library-service/internal/members/handler/auth"
	memberprofile "library-service/internal/members/handler/profile"
	"library-service/internal/payments/handler/payment"
	"library-service/internal/payments/handler/receipt"
	"library-service/internal/payments/handler/savedcard"
	middleware2 "library-service/internal/pkg/middleware"
	reservationhttp "library-service/internal/reservations/handler"
)

// RouterConfig holds router configuration
type RouterConfig struct {
	Config       *config.Config
	Usecases     *container.Container
	AuthServices *container.AuthServices
	Logger       *zap.Logger
}

// NewRouter creates a new HTTP router with all routes configured
func NewRouter(cfg RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	// Base middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware2.RequestLogger(cfg.Logger))
	r.Use(middleware2.ErrorHandler(cfg.Logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.Config.Server.ReadTimeout))
	r.Use(middleware.Heartbeat("/health"))

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	// Create auth middleware
	authMiddleware := middleware2.NewAuthMiddleware(cfg.AuthServices.JWTService)

	// Create validator (shared across all handlers)
	validator := middleware2.NewValidator()

	// Create handlers
	authHandler := memberauth.NewAuthHandler(
		cfg.Usecases,
		validator,
	)

	bookHandler := bookhttp.NewBookHandler(
		cfg.Usecases,
		validator,
	)

	reservationHandler := reservationhttp.NewReservationHandler(
		cfg.Usecases,
		validator,
	)

	paymentHandler := payment.NewPaymentHandler(
		cfg.Usecases,
		validator,
	)

	savedCardHandler := savedcard.NewSavedCardHandler(
		cfg.Usecases,
		validator,
	)

	authorHandler := authorhttp.NewAuthorHandler(
		cfg.Usecases,
	)

	memberHandler := memberprofile.NewMemberHandler(
		cfg.Usecases,
	)

	receiptHandler := receipt.NewReceiptHandler(
		cfg.Usecases,
		validator,
	)

	// Payment page handler
	paymentPageHandler, err := payment.NewPaymentPageHandler()
	if err != nil {
		panic(err) // In production, handle this more gracefully
	}

	// Payment page route (public)
	r.Get("/payment", paymentPageHandler.ServePaymentPage)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (handles public/protected internally)
		r.Mount("/auth", authHandler.Routes(authMiddleware))

		// Book routes (protected)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Mount("/books", bookHandler.Routes())
		})

		// Reservation routes (protected)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Mount("/reservations", reservationHandler.Routes())
		})

		// Payment routes
		r.Route("/payments", func(r chi.Router) {
			// Public callback endpoint (payment gateway calls this)
			r.Post("/callback", func(w http.ResponseWriter, req *http.Request) {
				paymentHandler.Routes().ServeHTTP(w, req)
			})

			// Protected payment routes
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.Authenticate)
				r.Mount("/", paymentHandler.Routes())
			})
		})

		// Saved card routes (protected)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Mount("/saved-cards", savedCardHandler.Routes())
		})

		// Author routes (protected)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Mount("/authors", authorHandler.Routes())
		})

		// Member routes (protected)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Mount("/members", memberHandler.Routes())
		})

		// Receipt routes (protected)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Mount("/receipts", receiptHandler.Routes())
		})
	})

	return r
}
