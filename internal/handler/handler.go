package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/oauth"
	"github.com/swaggo/http-swagger/v2"
	"library-service/docs"
	_ "library-service/docs"
	"library-service/internal/config"
	"library-service/internal/handler/http"
	"library-service/internal/service/auth"
	"library-service/internal/service/library"
	"library-service/internal/service/subscription"
	"library-service/pkg/server/router"
	"net/url"
)

type Dependencies struct {
	Configs             config.Configs
	AuthService         *auth.Service
	LibraryService      *library.Service
	SubscriptionService *subscription.Service
}

// Configuration is an alias for a function that will take in a pointer to a Handler and modify it
type Configuration func(h *Handler) error

// Handler is an implementation of the Handler
type Handler struct {
	dependencies Dependencies

	HTTP *chi.Mux
}

// New takes a variable amount of Configuration functions and returns a new Handler
// Each Configuration will be called in the order they are passed in
func New(d Dependencies, configs ...Configuration) (h *Handler, err error) {
	// Create the handler
	h = &Handler{
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
	return func(h *Handler) (err error) {
		// Create the http handler, if we needed parameters, such as connection strings they could be inputted here
		h.HTTP = router.New()

		// Init swagger handler
		docs.SwaggerInfo.BasePath = "/api/v1"
		docs.SwaggerInfo.Host = h.dependencies.Configs.HTTP.Host
		docs.SwaggerInfo.Schemes = []string{h.dependencies.Configs.HTTP.Schema}

		swaggerURL := url.URL{
			Scheme: h.dependencies.Configs.HTTP.Schema,
			Host:   h.dependencies.Configs.HTTP.Host,
			Path:   "swagger/doc.json",
		}

		h.HTTP.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL(swaggerURL.String()),
		))

		// Init auth handler
		authHandler := oauth.NewBearerServer(
			h.dependencies.Configs.OAUTH.Secret,
			h.dependencies.Configs.OAUTH.Expires,
			h.dependencies.AuthService, nil)

		h.HTTP.Post("/token", authHandler.UserCredentials)

		// Init service handlers
		authorHandler := http.NewAuthorHandler(h.dependencies.LibraryService)
		bookHandler := http.NewBookHandler(h.dependencies.LibraryService)
		memberHandler := http.NewMemberHandler(h.dependencies.SubscriptionService)

		h.HTTP.Route("/api/v1", func(r chi.Router) {
			// use the Bearer Authentication middleware
			r.Use(oauth.Authorize(h.dependencies.Configs.OAUTH.Secret, nil))

			r.Mount("/authors", authorHandler.Routes())
			r.Mount("/books", bookHandler.Routes())
			r.Mount("/members", memberHandler.Routes())
		})

		return
	}
}
