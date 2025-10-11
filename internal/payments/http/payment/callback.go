package payment

import (
	"net/http"

	"go.uber.org/zap"

	"library-service/internal/adapters/http/dto"
	paymentops "library-service/internal/payments/operations/payment"
	"library-service/pkg/httputil"
	"library-service/pkg/logutil"
)

// This file contains payment gateway webhook/callback handlers.
// These endpoints are called by external payment gateway (edomain.kz).

// @Summary Handle payment callback
// @Description Process callback from payment gateway
// @Tags payments
// @Accept json
// @Produce json
// @Param request body dto.PaymentCallbackRequest true "Payment callback data"
// @Success 200 {object} dto.PaymentCallbackResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /payments/callback [post]
func (h *PaymentHandler) handleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "payment_handler", "callback")

	// Decode request
	var req dto.PaymentCallbackRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Map DTO to use case request
	// Status is derived from code and reason fields using constants
	status := dto.GetPaymentStatus(req.Code, req.Reason)

	transactionID := ""
	if req.TransactionID != nil {
		transactionID = *req.TransactionID
	}

	// Execute usecase
	result, err := h.useCases.Payment.HandleCallback.Execute(ctx, paymentops.PaymentCallbackRequest{
		InvoiceID:       req.InvoiceID,
		TransactionID:   transactionID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Status:          status,
		CardMask:        req.CardMask,
		ApprovalCode:    req.ApprovalCode,
		ErrorCode:       req.ReasonCode,
		ErrorMessage:    &req.Reason,
		GatewayResponse: req.Extra,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTO
	response := dto.ToPaymentCallbackResponse(result)

	logger.Info("payment callback processed",
		zap.String("payment_id", result.PaymentID),
		zap.String("invoice_id", req.InvoiceID),
		zap.String("status", string(result.Status)),
	)
	h.RespondJSON(w, http.StatusOK, response)
}
