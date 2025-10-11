package book

import (
	"github.com/go-chi/chi/v5"

	"library-service/internal/adapters/http/handlers"
	"library-service/internal/adapters/http/middleware"
	"library-service/internal/usecase"
)

// BookHandler handles HTTP requests for books
type BookHandler struct {
	handlers.BaseHandler
	useCases  *usecase.Container
	validator *middleware.Validator
}

// NewBookHandler creates a new book handler
func NewBookHandler(
	useCases *usecase.Container,
	validator *middleware.Validator,
) *BookHandler {
	return &BookHandler{
		useCases:  useCases,
		validator: validator,
	}
}

// Routes returns the router for book endpoints
func (h *BookHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.create)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Put("/", h.update)
		r.Delete("/", h.delete)
		r.Get("/authors", h.listAuthors)
	})

	return r
}
