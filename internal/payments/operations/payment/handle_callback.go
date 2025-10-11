package payment

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"library-service/internal/payments/domain"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
)

// PaymentCallbackRequest represents the callback data from payment gateway.
type PaymentCallbackRequest struct {
	InvoiceID       string
	TransactionID   string
	Amount          int64
	Currency        string
	Status          string // "success", "failed", "cancelled"
	CardMask        *string
	ApprovalCode    *string
	ErrorCode       *string
	ErrorMessage    *string
	GatewayResponse map[string]interface{}
}

// HandleCallbackResponse represents the output of processing a payment callback.
type HandleCallbackResponse struct {
	PaymentID string
	Status    domain.Status
	Processed bool
}

// HandleCallbackUseCase handles callbacks from the payment gateway.
type HandleCallbackUseCase struct {
	paymentRepo    domain.Repository
	paymentService *domain.Service
}

// NewHandleCallbackUseCase creates a new instance of HandleCallbackUseCase.
func NewHandleCallbackUseCase(
	paymentRepo domain.Repository,
	paymentService *domain.Service,
) *HandleCallbackUseCase {
	return &HandleCallbackUseCase{
		paymentRepo:    paymentRepo,
		paymentService: paymentService,
	}
}

// Execute processes a callback from the payment gateway.
func (uc *HandleCallbackUseCase) Execute(ctx context.Context, req PaymentCallbackRequest) (HandleCallbackResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "handle_callback")

	logger.Info("processing payment callback")

	// Get payment by invoice ID
	paymentEntity, err := uc.paymentRepo.GetByInvoiceID(ctx, req.InvoiceID)
	if err != nil {
		logger.Error("failed to get payment by invoice ID", zap.Error(err))
		return HandleCallbackResponse{}, errors.ErrNotFound.WithDetails("invoice_id", req.InvoiceID)
	}

	// Security check: Validate amount matches
	if req.Amount != paymentEntity.Amount {
		logger.Error("callback amount mismatch",
			zap.Int64("expected_amount", paymentEntity.Amount),
			zap.Int64("callback_amount", req.Amount),
		)
		return HandleCallbackResponse{}, errors.NewError(errors.CodeValidation).
			WithDetail("field", "amount").
			WithDetail("reason", "callback amount does not match payment amount").
			Build()
	}

	// Security check: Validate currency matches
	if req.Currency != paymentEntity.Currency {
		logger.Error("callback currency mismatch",
			zap.String("expected_currency", paymentEntity.Currency),
			zap.String("callback_currency", req.Currency),
		)
		return HandleCallbackResponse{}, errors.NewError(errors.CodeValidation).
			WithDetail("field", "currency").
			WithDetail("reason", "callback currency does not match payment currency").
			Build()
	}

	// Idempotency check: If payment is already in a final state, don't process again
	if uc.paymentService.IsFinalStatus(paymentEntity.Status) {
		logger.Warn("payment already in final state, skipping callback",
			zap.String("payment_id", paymentEntity.ID),
			zap.String("current_status", string(paymentEntity.Status)),
		)
		return HandleCallbackResponse{
			PaymentID: paymentEntity.ID,
			Status:    paymentEntity.Status,
			Processed: false, // Not processed, already final
		}, nil
	}

	// Determine new status based on gateway response
	newStatus := uc.paymentService.MapGatewayStatus(req.Status)

	// Validate status transition
	if err := uc.paymentService.ValidateStatusTransition(paymentEntity.Status, newStatus); err != nil {
		logger.Warn("invalid status transition", zap.Error(err),
			zap.String("current_status", string(paymentEntity.Status)),
			zap.String("new_status", string(newStatus)),
		)
		return HandleCallbackResponse{}, err
	}

	// Store gateway response as JSON
	var gatewayResponseStr *string
	if req.GatewayResponse != nil {
		responseJSON, err := json.Marshal(req.GatewayResponse)
		if err != nil {
			logger.Warn("failed to marshal gateway response", zap.Error(err))
		} else {
			str := string(responseJSON)
			gatewayResponseStr = &str
		}
	}

	// Update payment entity using domain service
	uc.paymentService.UpdateStatusFromCallback(&paymentEntity, domain.CallbackData{
		TransactionID:   req.TransactionID,
		CardMask:        req.CardMask,
		ApprovalCode:    req.ApprovalCode,
		ErrorCode:       req.ErrorCode,
		ErrorMessage:    req.ErrorMessage,
		GatewayResponse: gatewayResponseStr,
		NewStatus:       newStatus,
	})

	// Update payment in repository
	if err := uc.paymentRepo.Update(ctx, paymentEntity.ID, paymentEntity); err != nil {
		logger.Error("failed to update payment in repository", zap.Error(err))
		return HandleCallbackResponse{}, errors.Database("database operation", err)
	}

	logger.Info("payment callback processed successfully",
		zap.String("payment_id", paymentEntity.ID),
		zap.String("new_status", string(newStatus)),
	)

	return HandleCallbackResponse{
		PaymentID: paymentEntity.ID,
		Status:    newStatus,
		Processed: true,
	}, nil
}

// ValidateCallbackSecurity performs additional security checks on callbacks.
// Recommended additional security measures for production:
// 1. IP Whitelisting: Only accept callbacks from known payment gateway IPs
// 2. HTTPS Only: Ensure callback endpoint only accepts HTTPS requests
// 3. Secret Token: Generate and validate a unique secret token per payment
// 4. Request Signing: Implement HMAC signature verification if gateway supports it
// 5. Replay Attack Protection: Store processed callback IDs with timestamps
// 6. Rate Limiting: Limit callback requests per IP/payment to prevent abuse
//
// Current implementation includes:
// - InvoiceID validation (payment must exist in database)
// - Amount and currency validation
// - Status transition validation
// - Idempotency protection (don't process final states twice)
// - Detailed security event logging
func (uc *HandleCallbackUseCase) ValidateCallbackSecurity(ctx context.Context, invoiceID string, amount int64, currency string) error {
	// This is a placeholder for additional security validations
	// In production, implement IP whitelisting, signature verification, etc.
	return nil
}
