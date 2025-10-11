package savedcard

import (
	"net/http"

	"go.uber.org/zap"

	"library-service/internal/adapters/http/dto"
	"library-service/internal/domain/payment"
	"library-service/internal/usecase/paymentops"
	"library-service/pkg/httputil"
	"library-service/pkg/logutil"
)

// This file contains CRUD operations for saved cards.
// Create, Read (list), and Delete operations.

// @Summary Save a payment card
// @Description Save a payment card for future use
// @Tags saved-cards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.SaveCardRequest true "Card data"
// @Success 200 {object} payment.SavedCardResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
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
	var req dto.SaveCardRequest
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
	response := payment.SavedCardResponse{
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
// @Success 200 {object} dto.ListSavedCardsResponse
// @Failure 500 {object} dto.ErrorResponse
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
	result, err := h.useCases.SavedCard.ListSavedCards.Execute(ctx, paymentops.ListSavedCardsRequest{
		MemberID: memberID,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTO
	cards := payment.ParseFromSavedCards(result.Cards)
	response := dto.ListSavedCardsResponse{
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
// @Success 200 {object} dto.DeleteSavedCardResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
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
	result, err := h.useCases.SavedCard.DeleteSavedCard.Execute(ctx, paymentops.DeleteSavedCardRequest{
		CardID:   cardID,
		MemberID: memberID,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	response := dto.DeleteSavedCardResponse{
		Success: result.Success,
	}

	logger.Info("card deleted", zap.String("card_id", cardID))
	h.RespondJSON(w, http.StatusOK, response)
}
