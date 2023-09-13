package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/oauth"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"library-service/docs"
	"library-service/internal/config"
	"library-service/internal/handler/http"
	"library-service/internal/service/auth"
	"library-service/internal/service/library"
	"library-service/internal/service/subscription"
	"library-service/pkg/server/router"
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

		h.HTTP.Use(middleware.Timeout(h.dependencies.Configs.APP.Timeout))

		// Init swagger handler
		docs.SwaggerInfo.BasePath = h.dependencies.Configs.APP.Path
		h.HTTP.Get("/swagger/*", httpSwagger.WrapHandler)

		// Init auth handler
		authHandler := oauth.NewBearerServer(
			h.dependencies.Configs.TOKEN.Salt,
			h.dependencies.Configs.TOKEN.Expires,
			h.dependencies.AuthService, nil)

		h.HTTP.Post("/token", authHandler.UserCredentials)
		h.HTTP.Post("/auth", authHandler.ClientCredentials)

		// Init service handlers
		authorHandler := http.NewAuthorHandler(h.dependencies.LibraryService)
		bookHandler := http.NewBookHandler(h.dependencies.LibraryService)
		memberHandler := http.NewMemberHandler(h.dependencies.SubscriptionService)

		h.HTTP.Route("/", func(r chi.Router) {
			// use the Bearer Authentication middleware
			r.Use(oauth.Authorize(h.dependencies.Configs.TOKEN.Salt, nil))

			r.Mount("/authors", authorHandler.Routes())
			r.Mount("/books", bookHandler.Routes())
			r.Mount("/members", memberHandler.Routes())
		})

		return
	}
}
