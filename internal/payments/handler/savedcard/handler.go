package savedcard

import (
	"library-service/internal/container"
	"library-service/internal/payments/domain"
	paymentops "library-service/internal/payments/service/payment"
	savedcardops "library-service/internal/payments/service/savedcard"
	"library-service/internal/pkg/handlers"
	"library-service/internal/pkg/httputil"
	"library-service/internal/pkg/logutil"
	"library-service/internal/pkg/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// SavedCardHandler handles HTTP requests for saved cards.
//
// BUSINESS CONTEXT:
// Saved cards allow members to:
//   - Store payment methods securely (via payment provider tokenization)
//   - Make faster payments without re-entering card details
//   - Set a default card for automatic selection
//   - Manage multiple payment methods
//
// This handler provides 4 endpoints:
//   - SaveCard: Tokenize and store a new card
//   - ListSavedCards: View all saved cards
//   - DeleteSavedCard: Remove a saved card
//   - SetDefaultCard: Set a card as the default payment method
type SavedCardHandler struct {
	handlers.BaseHandler
	useCases  *container.Container
	validator *middleware.Validator
}

// NewSavedCardHandler creates a new saved card handler.
func NewSavedCardHandler(
	useCases *container.Container,
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
//	POST   /                    → saveCard
//	GET    /                    → listSavedCards
//	DELETE /{id}                → deleteSavedCard
//	POST   /{id}/set-default    → setDefaultCard
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

// ================================================================================
// Handler Methods
// ================================================================================

// @Summary Save a payment card
// @Description Save a payment card for future use
// @Tags saved-cards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SaveCardRequest true "Card data"
// @Success 200 {object} domain.SavedCardResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /saved-cards [post]
func (h *SavedCardHandler) saveCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "saved_card_handler", "save")

	// Get member ID from context
	memberID, ok := h.GetMemberID(w, r)
	if !ok {
		return
	}

	// Decode request
	var req SaveCardRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Validate request
	if !h.validator.ValidateStruct(w, req) {
		return
	}

	// Execute usecase
	result, err := h.useCases.SavedCard.SaveCard.Execute(ctx, paymentops.SaveCardRequest{
		MemberID:    memberID,
		CardToken:   req.CardToken,
		CardMask:    req.CardMask,
		CardType:    req.CardType,
		ExpiryMonth: req.ExpiryMonth,
		ExpiryYear:  req.ExpiryYear,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTO
	response := domain.SavedCardResponse{
		ID:          result.CardID,
		CardMask:    result.CardMask,
		CardType:    result.CardType,
		ExpiryMonth: result.ExpiryMonth,
		ExpiryYear:  result.ExpiryYear,
		IsDefault:   result.IsDefault,
	}

	logger.Info("card saved", zap.String("card_id", result.CardID))
	h.RespondJSON(w, http.StatusOK, response)
}

// @Summary List saved cards
// @Description Get all saved cards for the authenticated member
// @Tags saved-cards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ListSavedCardsResponse
// @Failure 500 {object} ErrorResponse
// @Router /saved-cards [get]
func (h *SavedCardHandler) listSavedCards(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "saved_card_handler", "list")

	// Get member ID from context
	memberID, ok := h.GetMemberID(w, r)
	if !ok {
		return
	}

	// Execute usecase
	result, err := h.useCases.SavedCard.ListSavedCards.Execute(ctx, savedcardops.ListSavedCardsRequest{
		MemberID: memberID,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTO
	cards := domain.ParseFromSavedCards(result.Cards)
	response := ListSavedCardsResponse{
		Cards: cards,
	}

	logger.Info("saved cards listed", zap.Int("count", len(cards)))
	h.RespondJSON(w, http.StatusOK, response)
}

// @Summary Delete a saved card
// @Description Delete a saved payment card
// @Tags saved-cards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Card ID"
// @Success 200 {object} DeleteSavedCardResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /saved-cards/{id} [delete]
func (h *SavedCardHandler) deleteSavedCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "saved_card_handler", "delete")

	// Get member ID from context
	memberID, ok := h.GetMemberID(w, r)
	if !ok {
		return
	}

	// Get card ID from URL
	cardID, ok := h.GetURLParam(w, r, "id")
	if !ok {
		return
	}

	// Execute usecase
	result, err := h.useCases.SavedCard.DeleteSavedCard.Execute(ctx, savedcardops.DeleteSavedCardRequest{
		CardID:   cardID,
		MemberID: memberID,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	response := DeleteSavedCardResponse{
		Success: result.Success,
	}

	logger.Info("card deleted", zap.String("card_id", cardID))
	h.RespondJSON(w, http.StatusOK, response)
}

// @Summary Set default card
// @Description Set a saved card as the default payment method
// @Tags saved-cards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Card ID"
// @Success 200 {object} SetDefaultCardResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /saved-cards/{id}/set-default [post]
func (h *SavedCardHandler) setDefaultCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "saved_card_handler", "set_default")

	// Get member ID from context
	memberID, ok := h.GetMemberID(w, r)
	if !ok {
		return
	}

	// Get card ID from URL
	cardID, ok := h.GetURLParam(w, r, "id")
	if !ok {
		return
	}

	// Execute usecase
	result, err := h.useCases.SavedCard.SetDefaultCard.Execute(ctx, paymentops.SetDefaultCardRequest{
		CardID:   cardID,
		MemberID: memberID,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	response := SetDefaultCardResponse{
		Success: result.Success,
	}

	logger.Info("card set as default", zap.String("card_id", cardID))
	h.RespondJSON(w, http.StatusOK, response)
}
