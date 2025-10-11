package savedcard

import (
	"github.com/go-chi/chi/v5"

	"library-service/internal/adapters/http/handlers"
	"library-service/internal/adapters/http/middleware"
	"library-service/internal/usecase"
)

// SavedCardHandler handles HTTP requests for saved cards.
//
// ORGANIZATION:
// This file contains only the handler struct, constructor, and route definitions.
// Handler methods are split across files by feature area:
//   - saved_card_crud.go: CRUD operations (SaveCard, ListSavedCards, DeleteSavedCard)
//   - saved_card_manage.go: Card management (SetDefaultCard)
//
// RATIONALE:
// Splitting by operation type makes it easier to:
//   - Find CRUD operations in one place
//   - Separate management logic from basic operations
//   - Maintain focused, readable files (~70-100 lines each)
//
// BUSINESS CONTEXT:
// Saved cards allow members to:
//   - Store payment methods securely (via payment gateway tokenization)
//   - Make faster payments without re-entering card details
//   - Set a default card for automatic selection
//   - Manage multiple payment methods
type SavedCardHandler struct {
	handlers.BaseHandler
	useCases  *usecase.Container
	validator *middleware.Validator
}

// NewSavedCardHandler creates a new saved card handler.
func NewSavedCardHandler(
	useCases *usecase.Container,
	validator *middleware.Validator,
) *SavedCardHandler {
	return &SavedCardHandler{
		useCases:  useCases,
		validator: validator,
	}
}

// Routes returns the router for saved card endpoints.
//
// ROUTE STRUCTURE:
//
//	POST   /                    → saveCard (saved_card_crud.go)
//	GET    /                    → listSavedCards (saved_card_crud.go)
//	DELETE /{id}                → deleteSavedCard (saved_card_crud.go)
//	POST   /{id}/set-default    → setDefaultCard (saved_card_manage.go)
//
// SECURITY:
//   - All routes require authentication
//   - Cards are scoped to member (can only access own cards)
//   - Card tokens stored securely (never full card numbers)
func (h *SavedCardHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// CRUD operations
	r.Post("/", h.saveCard)
	r.Get("/", h.listSavedCards)

	// Card-specific operations
	r.Route("/{id}", func(r chi.Router) {
		r.Delete("/", h.deleteSavedCard)
		r.Post("/set-default", h.setDefaultCard)
	})

	return r
}
