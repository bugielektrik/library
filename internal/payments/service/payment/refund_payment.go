package payment

import (
	"context"
	"library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"library-service/internal/payments/domain"
)

// RefundPaymentRequest represents the input for refunding a domain.
type RefundPaymentRequest struct {
	PaymentID    string
	MemberID     string // For authorization check (admin can refund any, member only their own)
	Reason       string
	IsAdmin      bool
	RefundAmount *int64 // Optional: if nil, full refund; if specified, partial refund
}

// RefundPaymentResponse represents the output of refunding a domain.
type RefundPaymentResponse struct {
	PaymentID  string
	Status     domain.Status
	RefundedAt time.Time
	Amount     int64
	Currency   string
}

// RefundPaymentUseCase handles the refund of a completed domain.
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

// Execute processes a refund for a completed domain.
func (uc *RefundPaymentUseCase) Execute(ctx context.Context, req RefundPaymentRequest) (RefundPaymentResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "refund_payment")

	// Retrieve payment
	paymentEntity, err := uc.paymentRepo.GetByID(ctx, req.PaymentID)
	if err != nil {
		logger.Error("failed to retrieve payment", zap.Error(err))
		return RefundPaymentResponse{}, errors.NotFound("payment")
	}

	// Verify authorization
	if err := validateRefundAuthorization(req, paymentEntity, logger); err != nil {
		return RefundPaymentResponse{}, err
	}

	// Check refund eligibility
	if err := validateRefundEligibility(paymentEntity, uc.paymentService, logger); err != nil {
		return RefundPaymentResponse{}, err
	}

	// Determine and validate refund amount
	refundAmount, isPartialRefund, err := validateRefundAmount(req, paymentEntity, logger)
	if err != nil {
		return RefundPaymentResponse{}, err
	}

	// Call payment provider refund API
	if paymentEntity.GatewayTransactionID != nil && *paymentEntity.GatewayTransactionID != "" {
		logger.Info("calling provider refund API",
			zap.String("transaction_id", *paymentEntity.GatewayTransactionID),
			zap.Int64("refund_amount", refundAmount),
			zap.Bool("is_partial", isPartialRefund),
		)

		// Prepare refund parameters
		var gatewayAmount *float64
		if isPartialRefund {
			// Convert from smallest currency unit (cents/tenge) to decimal amount
			// Use decimal for precision to avoid floating-point errors
			amountDecimal := decimal.NewFromInt(refundAmount).Div(decimal.NewFromInt(100))
			amount, _ := amountDecimal.Float64() // Convert to float64 for provider API
			gatewayAmount = &amount

			logger.Debug("converted refund amount for provider",
				zap.Int64("amount_cents", refundAmount),
				zap.Float64("amount_decimal", amount),
			)
		}
		// If not partial refund, leave nil for full refund

		// Call provider using domain interface
		if err := uc.paymentGateway.RefundPayment(
			ctx,
			*paymentEntity.GatewayTransactionID,
			gatewayAmount,
			req.PaymentID, // externalID for tracking
		); err != nil {
			logger.Error("provider refund failed", zap.Error(err))
			return RefundPaymentResponse{}, errors.External("payment provider", err)
		}
		logger.Info("provider refund successful")
	} else {
		logger.Warn("no provider transaction ID, updating status only")
	}

	// Update payment status to refunded
	if err := uc.paymentRepo.UpdateStatus(ctx, req.PaymentID, domain.StatusRefunded); err != nil {
		logger.Error("failed to update payment status", zap.Error(err))
		return RefundPaymentResponse{}, errors.Database("database operation", err)
	}

	now := time.Now()

	logger.Info("payment refunded successfully")

	return RefundPaymentResponse{
		PaymentID:  req.PaymentID,
		Status:     domain.StatusRefunded,
		RefundedAt: now,
		Amount:     refundAmount, // Return actual refunded amount (may be partial)
		Currency:   paymentEntity.Currency,
	}, nil
}
