package savedcard

import (
	"library-service/internal/pkg/logutil"
	"net/http"

	"go.uber.org/zap"

	paymentops "library-service/internal/payments/service/payment"
)

// This file contains saved card management service.
// Operations that modify card settings rather than the card itself.

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
