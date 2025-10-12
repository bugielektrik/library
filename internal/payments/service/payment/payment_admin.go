package payment

import (
	"context"

	"library-service/internal/payments/domain"
	errors2 "library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"

	"go.uber.org/zap"
)

// ================================================================================
// Cancel Payment Use Case
// ================================================================================

// CancelPaymentRequest represents the input for canceling a payment.
type CancelPaymentRequest struct {
	PaymentID string
	MemberID  string // For authorization check
	Reason    string
}

// CancelPaymentResponse represents the output of canceling a payment.
type CancelPaymentResponse struct {
	PaymentID   string
	Status      domain.Status
	CancelledAt string
}

// CancelPaymentUseCase handles the cancellation of a pending payment.
type CancelPaymentUseCase struct {
	paymentRepo    domain.Repository
	paymentService *domain.Service
}

// NewCancelPaymentUseCase creates a new instance of CancelPaymentUseCase.
func NewCancelPaymentUseCase(
	paymentRepo domain.Repository,
	paymentService *domain.Service,
) *CancelPaymentUseCase {
	return &CancelPaymentUseCase{
		paymentRepo:    paymentRepo,
		paymentService: paymentService,
	}
}

// Execute cancels a payment if it's in a cancellable state.
func (uc *CancelPaymentUseCase) Execute(ctx context.Context, req CancelPaymentRequest) (CancelPaymentResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "cancel")

	// Retrieve payment
	payment, err := uc.paymentRepo.GetByID(ctx, req.PaymentID)
	if err != nil {
		logger.Error("failed to retrieve payment", zap.Error(err))
		return CancelPaymentResponse{}, errors2.NotFound("payment")
	}

	// Verify payment belongs to member
	if payment.MemberID != req.MemberID {
		logger.Warn("unauthorized cancellation attempt")
		return CancelPaymentResponse{}, errors2.NotFoundWithID("payment", req.PaymentID)
	}

	// Check if payment can be cancelled
	if payment.Status == domain.StatusCompleted {
		logger.Warn("cannot cancel completed payment")
		return CancelPaymentResponse{}, errors2.ErrInvalidPaymentStatus.
			WithDetails("status", string(payment.Status)).
			WithDetails("reason", "completed payments cannot be cancelled, use refund instead")
	}

	if payment.Status == domain.StatusCancelled {
		logger.Warn("payment already cancelled")
		return CancelPaymentResponse{}, errors2.ErrPaymentAlreadyProcessed.
			WithDetails("status", string(payment.Status))
	}

	if payment.Status == domain.StatusRefunded {
		logger.Warn("cannot cancel refunded payment")
		return CancelPaymentResponse{}, errors2.ErrInvalidPaymentStatus.
			WithDetails("status", string(payment.Status)).
			WithDetails("reason", "refunded payments cannot be cancelled")
	}

	// Validate status transition
	if err := uc.paymentService.ValidateStatusTransition(payment.Status, domain.StatusCancelled); err != nil {
		logger.Warn("invalid status transition", zap.Error(err))
		return CancelPaymentResponse{}, err
	}

	// Update payment status to cancelled
	if err := uc.paymentRepo.UpdateStatus(ctx, req.PaymentID, domain.StatusCancelled); err != nil {
		logger.Error("failed to update payment status", zap.Error(err))
		return CancelPaymentResponse{}, errors2.Database("database operation", err)
	}

	logger.Info("payment cancelled successfully")

	return CancelPaymentResponse{
		PaymentID:   req.PaymentID,
		Status:      domain.StatusCancelled,
		CancelledAt: "now",
	}, nil
}

// ================================================================================
// Refund Payment Use Case
// ================================================================================

// RefundPaymentRequest represents the input for refunding a payment.
type RefundPaymentRequest struct {
	PaymentID    string
	MemberID     string // For authorization check (admin can refund any, member only their own)
	Reason       string
	IsAdmin      bool
	RefundAmount *int64 // Optional: if nil, full refund; if specified, partial refund
}

// RefundPaymentResponse represents the output of refunding a payment.
type RefundPaymentResponse struct {
	PaymentID  string
	Status     domain.Status
	RefundedAt string
	Amount     int64
	Currency   string
}

// RefundPaymentUseCase handles the refund of a completed payment.
type RefundPaymentUseCase struct {
	paymentRepo    domain.Repository
	paymentService *domain.Service
	paymentGateway domain.Gateway
}

// NewRefundPaymentUseCase creates a new instance of RefundPaymentUseCase.
func NewRefundPaymentUseCase(
	paymentRepo domain.Repository,
	paymentService *domain.Service,
	paymentGateway domain.Gateway,
) *RefundPaymentUseCase {
	return &RefundPaymentUseCase{
		paymentRepo:    paymentRepo,
		paymentService: paymentService,
		paymentGateway: paymentGateway,
	}
}

// Execute processes a refund for a completed payment.
func (uc *RefundPaymentUseCase) Execute(ctx context.Context, req RefundPaymentRequest) (RefundPaymentResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "refund_payment")

	// Retrieve payment
	paymentEntity, err := uc.paymentRepo.GetByID(ctx, req.PaymentID)
	if err != nil {
		logger.Error("failed to retrieve payment", zap.Error(err))
		return RefundPaymentResponse{}, errors2.NotFound("payment")
	}

	// Verify authorization
	if err := uc.validateRefundAuthorization(req, paymentEntity, logger); err != nil {
		return RefundPaymentResponse{}, err
	}

	// Check refund eligibility
	if err := uc.validateRefundEligibility(paymentEntity, logger); err != nil {
		return RefundPaymentResponse{}, err
	}

	// Determine and validate refund amount
	refundAmount, isPartialRefund, err := uc.validateRefundAmount(req, paymentEntity, logger)
	if err != nil {
		return RefundPaymentResponse{}, err
	}

	// Call payment gateway refund API if transaction ID exists
	if paymentEntity.GatewayTransactionID != nil && *paymentEntity.GatewayTransactionID != "" {
		var gatewayAmount *float64
		if isPartialRefund {
			// Convert from smallest currency unit to decimal amount
			amount := float64(refundAmount) / 100.0
			gatewayAmount = &amount
		}

		if err := uc.paymentGateway.RefundPayment(ctx, *paymentEntity.GatewayTransactionID, gatewayAmount, req.PaymentID); err != nil {
			logger.Error("gateway refund failed", zap.Error(err))
			return RefundPaymentResponse{}, errors2.External("payment provider", err)
		}
	}

	// Update payment status to refunded
	if err := uc.paymentRepo.UpdateStatus(ctx, req.PaymentID, domain.StatusRefunded); err != nil {
		logger.Error("failed to update payment status", zap.Error(err))
		return RefundPaymentResponse{}, errors2.Database("database operation", err)
	}

	logger.Info("payment refunded successfully")

	return RefundPaymentResponse{
		PaymentID:  req.PaymentID,
		Status:     domain.StatusRefunded,
		RefundedAt: "now", // Will be set to actual time
		Amount:     refundAmount,
		Currency:   paymentEntity.Currency,
	}, nil
}

// validateRefundAuthorization checks if the member is authorized to refund the payment
func (uc *RefundPaymentUseCase) validateRefundAuthorization(req RefundPaymentRequest, paymentEntity domain.Payment, logger *zap.Logger) error {
	if req.IsAdmin {
		return nil
	}

	if paymentEntity.MemberID != req.MemberID {
		logger.Warn("unauthorized refund attempt")
		return errors2.NotFoundWithID("payment", req.PaymentID)
	}

	return nil
}

// validateRefundAmount validates and returns the refund amount
func (uc *RefundPaymentUseCase) validateRefundAmount(req RefundPaymentRequest, paymentEntity domain.Payment, logger *zap.Logger) (int64, bool, error) {
	if req.RefundAmount == nil {
		return paymentEntity.Amount, false, nil
	}

	requestedAmount := *req.RefundAmount

	if requestedAmount <= 0 {
		return 0, false, errors2.NewError(errors2.CodeValidation).
			WithDetail("field", "refund_amount").
			WithDetail("reason", "refund amount must be positive").
			Build()
	}

	if requestedAmount > paymentEntity.Amount {
		return 0, false, errors2.NewError(errors2.CodeValidation).
			WithDetail("field", "refund_amount").
			WithDetail("reason", "refund amount cannot exceed payment amount").
			Build()
	}

	return requestedAmount, true, nil
}

// validateRefundEligibility checks if the payment can be refunded
func (uc *RefundPaymentUseCase) validateRefundEligibility(paymentEntity domain.Payment, logger *zap.Logger) error {
	if !paymentEntity.CanBeRefunded() {
		logger.Warn("payment cannot be refunded")
		return errors2.ErrInvalidPaymentStatus.
			WithDetails("status", string(paymentEntity.Status)).
			WithDetails("reason", "only completed payments within 180 days can be refunded")
	}

	if err := uc.paymentService.ValidateStatusTransition(paymentEntity.Status, domain.StatusRefunded); err != nil {
		logger.Warn("invalid status transition", zap.Error(err))
		return err
	}

	return nil
}
