package payment

import (
	"net/http"

	"go.uber.org/zap"

	"library-service/internal/adapters/http/dto"
	paymentops "library-service/internal/payments/operations/payment"
	"library-service/pkg/httputil"
	"library-service/pkg/logutil"
)

// This file contains payment management handlers.
// These endpoints modify existing payment state (cancel, refund).

// @Summary Cancel a payment
// @Description Cancel a pending or processing payment
// @Tags payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Payment ID"
// @Param request body dto.CancelPaymentRequest true "Cancellation reason"
// @Success 200 {object} dto.CancelPaymentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /payments/{id}/cancel [post]
func (h *PaymentHandler) cancelPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "payment_handler", "cancel")

	// Get member ID from context
	memberID, ok := h.GetMemberID(w, r)
	if !ok {
		return
	}

	// Get payment ID from URL
	paymentID, ok := h.GetURLParam(w, r, "id")
	if !ok {
		return
	}

	// Decode request
	var req dto.CancelPaymentRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Execute usecase
	result, err := h.useCases.Payment.CancelPayment.Execute(ctx, paymentops.CancelPaymentRequest{
		PaymentID: paymentID,
		MemberID:  memberID,
		Reason:    req.Reason,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTO
	response := dto.ToCancelPaymentResponse(result)

	logger.Info("payment cancelled", zap.String("payment_id", paymentID))
	h.RespondJSON(w, http.StatusOK, response)
}

// @Summary Refund a payment
// @Description Refund a completed payment
// @Tags payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Payment ID"
// @Param request body dto.RefundPaymentRequest true "Refund reason"
// @Success 200 {object} dto.RefundPaymentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /payments/{id}/refund [post]
func (h *PaymentHandler) refundPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "payment_handler", "refund")

	// Get member ID from context
	memberID, ok := h.GetMemberID(w, r)
	if !ok {
		return
	}

	// Get payment ID from URL
	paymentID, ok := h.GetURLParam(w, r, "id")
	if !ok {
		return
	}

	// Decode request
	var req dto.RefundPaymentRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Check if user is admin (from role in context)
	role, _ := ctx.Value("role").(string)
	isAdmin := role == "admin"

	// Execute usecase
	result, err := h.useCases.Payment.RefundPayment.Execute(ctx, paymentops.RefundPaymentRequest{
		PaymentID:    paymentID,
		MemberID:     memberID,
		Reason:       req.Reason,
		IsAdmin:      isAdmin,
		RefundAmount: req.RefundAmount, // Optional: nil for full refund, specified for partial
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTO
	response := dto.ToRefundPaymentResponse(result)

	logger.Info("payment refunded", zap.String("payment_id", paymentID))
	h.RespondJSON(w, http.StatusOK, response)
}
