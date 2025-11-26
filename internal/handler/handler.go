package handler

import (
	"library-service/config"

	chiprometheus "github.com/766b/chi-prometheus"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"library-service/docs"
	"library-service/internal/handler/http"
	"library-service/internal/service"
	"library-service/pkg/server/router"
)

type Dependencies struct {
	Configs  *config.Configs
	Services *service.Services
}

type Configuration func(h *Handlers) error

type Handlers struct {
	dependencies Dependencies

	HTTP *chi.Mux
}

func New(d Dependencies, configs ...Configuration) (h *Handlers, err error) {
	h = &Handlers{
		dependencies: d,
	}

	for _, cfg := range configs {
		if err = cfg(h); err != nil {
			return
		}
	}

	return
}

func WithHTTPHandler() Configuration {
	return func(h *Handlers) (err error) {
		h.HTTP = router.New([]string{
			"/health",
			"/metrics",
			"/swagger/{*}",
		})

		h.HTTP.Use(middleware.RequestID)
		h.HTTP.Use(middleware.RealIP)
		h.HTTP.Use(middleware.Logger)
		h.HTTP.Use(middleware.Recoverer)
		h.HTTP.Use(middleware.Timeout(h.dependencies.Configs.APP.Timeout))
		h.HTTP.Use(middleware.URLFormat)
		h.HTTP.Use(middleware.StripSlashes)
		h.HTTP.Use(middleware.Heartbeat("/health"))

		prometheus.NewRegistry().MustRegister(
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		)
		prometheusMiddleware := chiprometheus.NewMiddleware("library-service")

		h.HTTP.Use(prometheusMiddleware)

		docs.SwaggerInfo.BasePath = h.dependencies.Configs.APP.Path
		h.HTTP.Get("/swagger/*", httpSwagger.WrapHandler)

		authorHandler := http.NewAuthorHandler(h.dependencies.Services.Author)
		bookHandler := http.NewBookHandler(h.dependencies.Services.Book)
		memberHandler := http.NewMemberHandler(h.dependencies.Services.Member)

		h.HTTP.Route("/api/v1", func(r chi.Router) {
			r.Mount("/authors", authorHandler.Routes())
			r.Mount("/books", bookHandler.Routes())
			r.Mount("/members", memberHandler.Routes())
		})

		return
	}
}
