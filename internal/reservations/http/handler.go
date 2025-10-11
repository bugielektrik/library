package http

import (
	"github.com/go-chi/chi/v5"

	"library-service/internal/adapters/http/handlers"
	"library-service/internal/adapters/http/middleware"
	"library-service/internal/usecase"
)

// ReservationHandler handles HTTP requests for reservations
type ReservationHandler struct {
	handlers.BaseHandler
	useCases  *usecase.Container
	validator *middleware.Validator
}

// NewReservationHandler creates a new reservation handler
func NewReservationHandler(
	useCases *usecase.Container,
	validator *middleware.Validator,
) *ReservationHandler {
	return &ReservationHandler{
		useCases:  useCases,
		validator: validator,
	}
}

// Routes returns the router for reservation endpoints
func (h *ReservationHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.listMyReservations)
	r.Post("/", h.create)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Delete("/", h.cancel)
	})

	return r
}
