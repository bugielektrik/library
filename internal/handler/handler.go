package handler

import (
	chiprometheus "github.com/766b/chi-prometheus"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"library-service/docs"
	"library-service/internal/config"
	"library-service/internal/handler/http"
	"library-service/internal/service"
	"library-service/pkg/server/router"
)

type Dependencies struct {
	Configs  *config.Configs
	Services *service.Services
}

// Configuration is an alias for a function that will take in a pointer to a Handler and modify it
type Configuration func(h *Handlers) error

// Handlers is an implementation of the Handlers
type Handlers struct {
	dependencies Dependencies

	HTTP *chi.Mux
}

// New takes a variable amount of Configuration functions and returns a new Handler
// Each Configuration will be called in the order they are passed in
func New(d Dependencies, configs ...Configuration) (h *Handlers, err error) {
	// Create the handler
	h = &Handlers{
		dependencies: d,
	}

	// Apply all Configurations passed in
	for _, cfg := range configs {
		// Pass the service into the configuration function
		if err = cfg(h); err != nil {
			return
		}
	}

	return
}

// WithHTTPHandler applies a http handler to the Handler
func WithHTTPHandler() Configuration {
	return func(h *Handlers) (err error) {
		// Create the http handler, if we needed parameters, such as connection strings they could be inputted here
		h.HTTP = router.New([]string{
			"/health",
			"/metrics",
			"/swagger/{*}",
		})

		// Add some default middleware to the http handler
		h.HTTP.Use(middleware.RequestID)
		h.HTTP.Use(middleware.RealIP)
		h.HTTP.Use(middleware.Logger)
		h.HTTP.Use(middleware.Recoverer)
		h.HTTP.Use(middleware.Timeout(h.dependencies.Configs.APP.Timeout))
		h.HTTP.Use(middleware.URLFormat)
		h.HTTP.Use(middleware.StripSlashes)
		h.HTTP.Use(middleware.Heartbeat("/health"))

		// Add prometheus monitoring
		prometheus.NewRegistry().MustRegister(
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		)
		prometheusMiddleware := chiprometheus.NewMiddleware("library-service")

		h.HTTP.Use(prometheusMiddleware)

		// Swagger documentation handler setup
		docs.SwaggerInfo.BasePath = h.dependencies.Configs.APP.Path
		h.HTTP.Get("/swagger/*", httpSwagger.WrapHandler)

		// Add all routes to the http handler
		authorHandler := http.NewAuthorHandler(h.dependencies.Services.Library)
		bookHandler := http.NewBookHandler(h.dependencies.Services.Library)
		memberHandler := http.NewMemberHandler(h.dependencies.Services.Subscription)

		// Mount all the routes
		h.HTTP.Route("/", func(r chi.Router) {
			r.Mount("/authors", authorHandler.Routes())
			r.Mount("/books", bookHandler.Routes())
			r.Mount("/members", memberHandler.Routes())
		})

		return
	}
}
