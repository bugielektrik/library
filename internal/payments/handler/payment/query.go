package payment

import (
	"library-service/internal/pkg/logutil"
	"net/http"

	"go.uber.org/zap"

	paymentops "library-service/internal/payments/service/payment"
)

// This file contains payment query/read handler.
// These endpoints retrieve payment information without modifying state.

// @Summary Verify payment status
// @Description Get the current status of a payment
// @Tags payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Payment ID"
// @Success 200 {object} PaymentResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/{id} [get]
func (h *PaymentHandler) verifyPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "payment_handler", "verify")

	id, ok := h.GetURLParam(w, r, "id")
	if !ok {
		return
	}

	// Execute usecase
	result, err := h.useCases.Payment.VerifyPayment.Execute(ctx, paymentops.VerifyPaymentRequest{PaymentID: id})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTO
	response := ToPaymentResponse(result)

	logger.Info("payment verified", zap.String("payment_id", id), zap.String("status", string(result.Status)))
	h.RespondJSON(w, http.StatusOK, response)
}

// @Summary List member payments
// @Description Get all payments for a specific member
// @Tags payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param memberId path string true "Member ID"
// @Success 200 {object} ListPaymentsResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/member/{memberId} [get]
func (h *PaymentHandler) listMemberPayments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "payment_handler", "list_member_payments")

	memberID, ok := h.GetURLParam(w, r, "memberId")
	if !ok {
		return
	}

	// Execute usecase
	result, err := h.useCases.Payment.ListMemberPayments.Execute(ctx, paymentops.ListMemberPaymentsRequest{MemberID: memberID})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTOs
	payments := ToPaymentSummaryResponses(result.Payments)

	response := ListPaymentsResponse{
		Payments: payments,
	}

	logger.Info("member payments listed", zap.String("member_id", memberID), zap.Int("count", len(payments)))
	h.RespondJSON(w, http.StatusOK, response)
}
