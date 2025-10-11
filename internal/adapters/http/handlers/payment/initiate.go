package payment

import (
	"net/http"

	"go.uber.org/zap"

	"library-service/internal/adapters/http/dto"
	"library-service/internal/usecase/paymentops"
	"library-service/pkg/httputil"
	"library-service/pkg/logutil"
)

// This file contains payment initiation/creation handlers.
// These endpoints start new payment flows.

// @Summary Initiate a payment
// @Description Initiate a payment and get payment gateway details
// @Tags payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.InitiatePaymentRequest true "Payment initiation data"
// @Success 200 {object} dto.InitiatePaymentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
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
	var req dto.InitiatePaymentRequest
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
	response := dto.ToInitiatePaymentResponse(result)

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
// @Param request body dto.PayWithSavedCardRequest true "Payment data"
// @Success 200 {object} dto.PayWithSavedCardResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
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
	var req dto.PayWithSavedCardRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Validate request
	if !h.validator.ValidateStruct(w, req) {
		return
	}

	// Execute usecase
	result, err := h.useCases.Payment.PayWithSavedCard.Execute(ctx, paymentops.PayWithSavedCardRequest{
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
	response := dto.ToPayWithSavedCardResponse(result)

	logger.Info("payment with saved card initiated",
		zap.String("payment_id", result.PaymentID),
		zap.String("card_mask", result.CardMask),
	)
	h.RespondJSON(w, http.StatusOK, response)
}
