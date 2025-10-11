package payment

import (
	"context"
	"time"

	"go.uber.org/zap"

	"library-service/internal/payments/domain"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
)

// CancelPaymentRequest represents the input for cancelling a domain.
type CancelPaymentRequest struct {
	PaymentID string
	MemberID  string // For authorization check
	Reason    string
}

// CancelPaymentResponse represents the output of cancelling a domain.
type CancelPaymentResponse struct {
	PaymentID   string
	Status      domain.Status
	CancelledAt time.Time
}

// CancelPaymentUseCase handles the cancellation of a domain.
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
	logger := logutil.UseCaseLogger(ctx, "payment", "cancel_payment")

	// Retrieve payment
	paymentEntity, err := uc.paymentRepo.GetByID(ctx, req.PaymentID)
	if err != nil {
		logger.Error("failed to retrieve payment", zap.Error(err))
		return CancelPaymentResponse{}, errors.NotFound("payment")
	}

	// Verify payment belongs to member
	if paymentEntity.MemberID != req.MemberID {
		logger.Warn("unauthorized cancellation attempt",
			zap.String("payment_member_id", paymentEntity.MemberID),
			zap.String("requesting_member_id", req.MemberID),
		)
		return CancelPaymentResponse{}, errors.NotFoundWithID("payment", req.PaymentID)
	}

	// Check if payment can be cancelled
	if paymentEntity.Status == domain.StatusCompleted {
		logger.Warn("cannot cancel completed payment")
		return CancelPaymentResponse{}, errors.ErrInvalidPaymentStatus.
			WithDetails("status", string(paymentEntity.Status)).
			WithDetails("reason", "completed payments cannot be cancelled, use refund instead")
	}

	if paymentEntity.Status == domain.StatusCancelled {
		logger.Warn("payment already cancelled")
		return CancelPaymentResponse{}, errors.ErrPaymentAlreadyProcessed.
			WithDetails("status", string(paymentEntity.Status))
	}

	if paymentEntity.Status == domain.StatusRefunded {
		logger.Warn("cannot cancel refunded payment")
		return CancelPaymentResponse{}, errors.ErrInvalidPaymentStatus.
			WithDetails("status", string(paymentEntity.Status)).
			WithDetails("reason", "refunded payments cannot be cancelled")
	}

	// Validate status transition
	if err := uc.paymentService.ValidateStatusTransition(paymentEntity.Status, domain.StatusCancelled); err != nil {
		logger.Warn("invalid status transition", zap.Error(err))
		return CancelPaymentResponse{}, err
	}

	// Update payment status to cancelled
	if err := uc.paymentRepo.UpdateStatus(ctx, req.PaymentID, domain.StatusCancelled); err != nil {
		logger.Error("failed to update payment status", zap.Error(err))
		return CancelPaymentResponse{}, errors.Database("database operation", err)
	}

	now := time.Now()

	logger.Info("payment cancelled successfully")

	return CancelPaymentResponse{
		PaymentID:   req.PaymentID,
		Status:      domain.StatusCancelled,
		CancelledAt: now,
	}, nil
}
