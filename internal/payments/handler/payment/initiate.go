package payment

import (
	"library-service/internal/pkg/httputil"
	"library-service/internal/pkg/logutil"
	"net/http"

	"go.uber.org/zap"

	paymentops "library-service/internal/payments/service/payment"
	savedcardops "library-service/internal/payments/service/savedcard"
)

// This file contains payment initiation/creation handler.
// These endpoints start new payment flows.

// @Summary Initiate a payment
// @Description Initiate a payment and get payment provider details
// @Tags payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body InitiatePaymentRequest true "Payment initiation data"
// @Success 200 {object} InitiatePaymentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/initiate [post]
func (h *PaymentHandler) initiatePayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "payment_handler", "initiate")

	// Get member ID from context
	memberID, ok := h.GetMemberID(w, r)
	if !ok {
		return
	}

	// Decode request
	var req InitiatePaymentRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Validate request
	if !h.validator.ValidateStruct(w, req) {
		return
	}

	// Execute usecase
	result, err := h.useCases.Payment.InitiatePayment.Execute(ctx, paymentops.InitiatePaymentRequest{
		MemberID:        memberID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		PaymentType:     req.PaymentType,
		RelatedEntityID: req.RelatedEntityID,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTO
	response := ToInitiatePaymentResponse(result)

	logger.Info("payment initiated",
		zap.String("payment_id", response.PaymentID),
		zap.String("invoice_id", response.InvoiceID),
	)
	h.RespondJSON(w, http.StatusOK, response)
}

// @Summary Pay with saved card
// @Description Make a payment using a saved card
// @Tags payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PayWithSavedCardRequest true "Payment data"
// @Success 200 {object} PayWithSavedCardResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/pay-with-card [post]
func (h *PaymentHandler) payWithSavedCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "payment_handler", "pay_with_saved_card")

	// Get member ID from context
	memberID, ok := h.GetMemberID(w, r)
	if !ok {
		return
	}

	// Decode request
	var req PayWithSavedCardRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Validate request
	if !h.validator.ValidateStruct(w, req) {
		return
	}

	// Execute usecase
	result, err := h.useCases.Payment.PayWithSavedCard.Execute(ctx, savedcardops.PayWithSavedCardRequest{
		MemberID:        memberID,
		SavedCardID:     req.SavedCardID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		PaymentType:     req.PaymentType,
		RelatedEntityID: req.RelatedEntityID,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTO
	response := ToPayWithSavedCardResponse(result)

	logger.Info("payment with saved card initiated",
		zap.String("payment_id", result.PaymentID),
		zap.String("card_mask", result.CardMask),
	)
	h.RespondJSON(w, http.StatusOK, response)
}
