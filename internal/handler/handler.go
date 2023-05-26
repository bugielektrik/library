package handler

import (
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "library/docs"
	"library/internal/handler/http"
	"library/internal/service"
	"library/pkg/router"
)

type Dependencies struct {
	AuthorService service.AuthorService
	BookService   service.BookService
	MemberService service.MemberService
}

type Handler struct {
	dependencies Dependencies

	HTTP *chi.Mux
}

// Configuration is an alias for a function that will take in a pointer to a Handler and modify it
type Configuration func(r *Handler) error

// New takes a variable amount of Configuration functions and returns a new Handler
// Each Configuration will be called in the order they are passed in
func New(d Dependencies, configs ...Configuration) (r *Handler, err error) {
	// Add the Handler
	r = &Handler{
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

//	@title			Swagger Example API
//	@version		1.0
//	@description	This is a sample server celler server.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost
//	@BasePath	/api/v1

func WithHTTPHandler() Configuration {
	return func(h *Handler) (err error) {
		h.HTTP = router.New()

		h.HTTP.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("http://localhost/swagger/doc.json"), //The url pointing to API definition
		))

		h.HTTP.Route("/api/v1", func(r chi.Router) {
			r.Mount("/authors", http.NewAuthorHandler(h.dependencies.AuthorService).Routes())
			r.Mount("/books", http.NewBookHandler(h.dependencies.BookService).Routes())
			r.Mount("/members", http.NewMemberHandler(h.dependencies.MemberService).Routes())
		})

		return
	}
}
