package payment

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/payments/domain"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
)

// VerifyPaymentRequest represents the input for verifying a domain.
type VerifyPaymentRequest struct {
	PaymentID string
}

// VerifyPaymentResponse represents the output of verifying a domain.
type VerifyPaymentResponse struct {
	PaymentID            string
	InvoiceID            string
	Status               domain.Status
	Amount               int64
	Currency             string
	GatewayTransactionID *string
	CardMask             *string
	ApprovalCode         *string
	ErrorCode            *string
	ErrorMessage         *string
}

// VerifyPaymentUseCase handles the verification of a payment status.
type VerifyPaymentUseCase struct {
	paymentRepo    domain.Repository
	paymentService *domain.Service
	paymentGateway domain.Gateway
}

// NewVerifyPaymentUseCase creates a new instance of VerifyPaymentUseCase.
func NewVerifyPaymentUseCase(
	paymentRepo domain.Repository,
	paymentService *domain.Service,
	paymentGateway domain.Gateway,
) *VerifyPaymentUseCase {
	return &VerifyPaymentUseCase{
		paymentRepo:    paymentRepo,
		paymentService: paymentService,
		paymentGateway: paymentGateway,
	}
}

// Execute verifies the status of a payment by checking with the gateway.
func (uc *VerifyPaymentUseCase) Execute(ctx context.Context, req VerifyPaymentRequest) (VerifyPaymentResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "verify_payment")

	// Get payment from repository
	paymentEntity, err := uc.paymentRepo.GetByID(ctx, req.PaymentID)
	if err != nil {
		logger.Error("failed to get payment from repository", zap.Error(err))
		return VerifyPaymentResponse{}, errors.NotFoundWithID("payment", req.PaymentID)
	}

	// Check if payment has expired
	if uc.paymentService.IsExpired(paymentEntity) && paymentEntity.Status == domain.StatusPending {
		logger.Warn("payment has expired", zap.String("expires_at", paymentEntity.ExpiresAt.String()))
		handleExpiredPayment(ctx, uc.paymentRepo, &paymentEntity, logger)
	}

	// If payment is not in final state, check status with gateway
	if isPaymentUpdatable(paymentEntity.Status) {
		logger.Info("checking payment status with gateway", zap.String("invoice_id", paymentEntity.InvoiceID))

		// Call gateway API to check status
		statusResp, err := uc.paymentGateway.CheckPaymentStatus(ctx, paymentEntity.InvoiceID)
		if err != nil {
			logger.Warn("failed to check payment status with gateway", zap.Error(err))
			// Don't fail the request, just return current status
		} else {
			logger.Info("gateway status check successful",
				zap.String("result_code", statusResp.ResultCode),
				zap.String("transaction_id", statusResp.Transaction.ID),
			)

			// Update payment based on gateway response
			updated := uc.updatePaymentFromGatewayResponse(ctx, &paymentEntity, statusResp, logger)
			if updated {
				// Reload payment entity from repository to get updated values
				if reloadedPayment, err := uc.paymentRepo.GetByID(ctx, req.PaymentID); err == nil {
					paymentEntity = reloadedPayment
				}
			}
		}
	}

	logger.Info("payment verified")

	return VerifyPaymentResponse{
		PaymentID:            paymentEntity.ID,
		InvoiceID:            paymentEntity.InvoiceID,
		Status:               paymentEntity.Status,
		Amount:               paymentEntity.Amount,
		Currency:             paymentEntity.Currency,
		GatewayTransactionID: paymentEntity.GatewayTransactionID,
		CardMask:             paymentEntity.CardMask,
		ApprovalCode:         paymentEntity.ApprovalCode,
		ErrorCode:            paymentEntity.ErrorCode,
		ErrorMessage:         paymentEntity.ErrorMessage,
	}, nil
}

// updatePaymentFromGatewayResponse updates payment based on gateway status response.
// Returns true if payment was updated.
func (uc *VerifyPaymentUseCase) updatePaymentFromGatewayResponse(
	ctx context.Context,
	paymentEntity *domain.Payment,
	gatewayResp *domain.GatewayStatusResponse,
	logger *zap.Logger,
) bool {
	transaction := gatewayResp.Transaction

	logger.Info("processing gateway response")
	newStatus := uc.paymentService.MapGatewayStatus(transaction.Status)

	// Update status if changed
	if newStatus != paymentEntity.Status {
		logger.Info("payment status changed",
			zap.String("old_status", string(paymentEntity.Status)),
			zap.String("new_status", string(newStatus)),
		)

		if err := uc.paymentRepo.UpdateStatus(ctx, paymentEntity.ID, newStatus); err != nil {
			logger.Error("failed to update payment status", zap.Error(err))
			return false
		}
	}

	// Update additional payment fields
	if updatePaymentFields(paymentEntity, transaction) {
		if err := uc.paymentRepo.Update(ctx, paymentEntity.ID, *paymentEntity); err != nil {
			logger.Error("failed to update payment details", zap.Error(err))
			return false
		}
	}

	return true
}
