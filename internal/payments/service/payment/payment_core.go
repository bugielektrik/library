package payment

import (
	"context"

	"library-service/internal/payments/domain"
	errors2 "library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"

	"go.uber.org/zap"
)

// ================================================================================
// Common Types & Interfaces
// ================================================================================

// Validator defines the validation interface
type Validator interface {
	Validate(i interface{}) error
}

// ================================================================================
// Initiate Payment Use Case
// ================================================================================

// InitiatePaymentRequest represents the input for initiating a payment.
type InitiatePaymentRequest struct {
	MemberID        string             `validate:"required"`
	Amount          int64              `validate:"required,min=100,max=10000000"`
	Currency        string             `validate:"required,oneof=KZT USD EUR RUB"`
	PaymentType     domain.PaymentType `validate:"required"`
	RelatedEntityID *string            `validate:"omitempty"`
}

// InitiatePaymentResponse represents the output of initiating a payment.
type InitiatePaymentResponse struct {
	PaymentID string
	InvoiceID string
	AuthToken string
	Terminal  string
	Amount    int64
	Currency  string
	BackLink  string
	PostLink  string
	WidgetURL string
}

// InitiatePaymentUseCase handles the initiation of a new payment.
type InitiatePaymentUseCase struct {
	paymentRepo    domain.Repository
	paymentService *domain.Service
	paymentGateway domain.Gateway
	gatewayConfig  domain.GatewayConfig
	validator      Validator
}

// NewInitiatePaymentUseCase creates a new instance of InitiatePaymentUseCase.
func NewInitiatePaymentUseCase(
	paymentRepo domain.Repository,
	paymentService *domain.Service,
	gateway interface {
		domain.Gateway
		domain.GatewayConfig
	},
	validator Validator,
) *InitiatePaymentUseCase {
	return &InitiatePaymentUseCase{
		paymentRepo:    paymentRepo,
		paymentService: paymentService,
		paymentGateway: gateway,
		gatewayConfig:  gateway,
		validator:      validator,
	}
}

// Execute initiates a new payment in the system.
func (uc *InitiatePaymentUseCase) Execute(ctx context.Context, req InitiatePaymentRequest) (InitiatePaymentResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "initiate")

	// Validate request using go-playground/validator
	if err := uc.validator.Validate(req); err != nil {
		return InitiatePaymentResponse{}, errors2.ErrValidation.Wrap(err)
	}

	// Generate unique invoice ID
	invoiceID := uc.paymentService.GenerateInvoiceID(req.MemberID, req.PaymentType)

	// Create payment entity from request
	paymentEntity := domain.New(domain.Request{
		InvoiceID:       invoiceID,
		MemberID:        req.MemberID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		PaymentType:     req.PaymentType,
		RelatedEntityID: req.RelatedEntityID,
	})

	// Validate payment using domain service
	if err := uc.paymentService.Validate(paymentEntity); err != nil {
		logger.Warn("validation failed", zap.Error(err))
		return InitiatePaymentResponse{}, err
	}

	// Save to repository
	paymentID, err := uc.paymentRepo.Create(ctx, paymentEntity)
	if err != nil {
		logger.Error("failed to create payment in repository", zap.Error(err))
		return InitiatePaymentResponse{}, errors2.Database("database operation", err)
	}
	paymentEntity.ID = paymentID

	// Get auth token from payment gateway
	authToken, err := uc.paymentGateway.GetAuthToken(ctx)
	if err != nil {
		logger.Error("failed to get auth token from payment gateway", zap.Error(err))

		// Update payment status to failed
		if updateErr := uc.paymentRepo.UpdateStatus(ctx, paymentID, domain.StatusFailed); updateErr != nil {
			logger.Error("failed to update payment status", zap.Error(updateErr))
		}

		return InitiatePaymentResponse{}, errors2.External("payment gateway", err)
	}

	logger.Info("payment initiated successfully",
		zap.String("payment_id", paymentID),
		zap.String("invoice_id", invoiceID),
	)

	return InitiatePaymentResponse{
		PaymentID: paymentID,
		InvoiceID: invoiceID,
		AuthToken: authToken,
		Terminal:  uc.gatewayConfig.GetTerminal(),
		Amount:    req.Amount,
		Currency:  req.Currency,
		BackLink:  uc.gatewayConfig.GetBackLink(),
		PostLink:  uc.gatewayConfig.GetPostLink(),
		WidgetURL: uc.gatewayConfig.GetWidgetURL(),
	}, nil
}

// ================================================================================
// Verify Payment Use Case
// ================================================================================

// VerifyPaymentRequest represents the input for verifying a payment.
type VerifyPaymentRequest struct {
	PaymentID string
}

// VerifyPaymentResponse represents the output of verifying a payment.
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
		return VerifyPaymentResponse{}, errors2.NotFoundWithID("payment", req.PaymentID)
	}

	// Check if payment has expired
	if uc.paymentService.IsExpired(paymentEntity) && paymentEntity.Status == domain.StatusPending {
		logger.Warn("payment has expired", zap.String("expires_at", paymentEntity.ExpiresAt.String()))
		uc.handleExpiredPayment(ctx, &paymentEntity, logger)
	}

	// If payment is not in final state, check status with gateway
	if uc.isPaymentUpdatable(paymentEntity.Status) {
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

// handleExpiredPayment updates an expired payment to failed status
func (uc *VerifyPaymentUseCase) handleExpiredPayment(ctx context.Context, payment *domain.Payment, logger *zap.Logger) {
	if err := uc.paymentRepo.UpdateStatus(ctx, payment.ID, domain.StatusFailed); err != nil {
		logger.Error("failed to update expired payment status", zap.Error(err))
		return
	}
	payment.Status = domain.StatusFailed
	logger.Info("expired payment marked as failed")
}

// isPaymentUpdatable checks if payment status can be updated
func (uc *VerifyPaymentUseCase) isPaymentUpdatable(status domain.Status) bool {
	return status == domain.StatusPending || status == domain.StatusProcessing
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
	if uc.updatePaymentFields(paymentEntity, transaction) {
		if err := uc.paymentRepo.Update(ctx, paymentEntity.ID, *paymentEntity); err != nil {
			logger.Error("failed to update payment details", zap.Error(err))
			return false
		}
	}

	return true
}

// updatePaymentFields updates payment fields from gateway transaction details.
// Returns true if any fields were updated.
func (uc *VerifyPaymentUseCase) updatePaymentFields(payment *domain.Payment, transaction domain.GatewayTransactionDetails) bool {
	updated := false

	if transaction.ID != "" && payment.GatewayTransactionID == nil {
		payment.GatewayTransactionID = &transaction.ID
		updated = true
	}

	if transaction.CardMask != "" && payment.CardMask == nil {
		payment.CardMask = &transaction.CardMask
		updated = true
	}

	if transaction.ApprovalCode != "" && payment.ApprovalCode == nil {
		payment.ApprovalCode = &transaction.ApprovalCode
		updated = true
	}

	return updated
}
