package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"library/internal/handler/rest"
	"library/internal/service"
	"library/pkg/router"
)

type Dependencies struct {
	AuthorService service.AuthorService
	BookService   service.BookService
	MemberService service.MemberService
}

type Handler struct {
	HTTP         http.Handler
	Dependencies Dependencies
}

// Configuration is an alias for a function that will take in a pointer to a Handler and modify it
type Configuration func(r *Handler) error

// New takes a variable amount of Configuration functions and returns a new Handler
// Each Configuration will be called in the order they are passed in
func New(configs ...Configuration) (r *Handler, err error) {
	// Create the Handler
	r = &Handler{}
	// Apply all Configurations passed in
	for _, cfg := range configs {
		// Pass the service into the configuration function
		if err = cfg(r); err != nil {
			return
		}
	}
	return
}

func WithDependencies(d Dependencies) Configuration {
	return func(h *Handler) (err error) {
		h.Dependencies = d
		return
	}
}

func WithHTTPHandler() Configuration {
	return func(h *Handler) (err error) {
		r := router.New()

		r.Route("/api/v1", func(r chi.Router) {
			r.Mount("/authors", rest.NewAuthorHandler(h.Dependencies.AuthorService).Routes())
			r.Mount("/books", rest.NewBookHandler(h.Dependencies.BookService).Routes())
			r.Mount("/members", rest.NewMemberHandler(h.Dependencies.MemberService).Routes())
		})

		h.HTTP = r
		return
	}
}
