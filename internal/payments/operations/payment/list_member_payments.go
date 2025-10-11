package payment

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/payments/domain"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
)

// ListMemberPaymentsRequest represents the input for listing member payments.
type ListMemberPaymentsRequest struct {
	MemberID string
}

// ListMemberPaymentsResponse represents the output of listing member payments.
type ListMemberPaymentsResponse struct {
	Payments []PaymentSummary
}

// PaymentSummary represents a summary of a domain.
type PaymentSummary struct {
	ID          string
	InvoiceID   string
	Amount      int64
	Currency    string
	Status      domain.Status
	PaymentType domain.PaymentType
	CreatedAt   string
	CompletedAt *string
}

// ListMemberPaymentsUseCase handles listing all payments for a member.
type ListMemberPaymentsUseCase struct {
	paymentRepo domain.Repository
}

// NewListMemberPaymentsUseCase creates a new instance of ListMemberPaymentsUseCase.
func NewListMemberPaymentsUseCase(paymentRepo domain.Repository) *ListMemberPaymentsUseCase {
	return &ListMemberPaymentsUseCase{
		paymentRepo: paymentRepo,
	}
}

// Execute retrieves all payments for a specific member.
func (uc *ListMemberPaymentsUseCase) Execute(ctx context.Context, req ListMemberPaymentsRequest) (ListMemberPaymentsResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "list_member_payments")

	// Get payments from repository
	payments, err := uc.paymentRepo.ListByMemberID(ctx, req.MemberID)
	if err != nil {
		logger.Error("failed to get payments from repository", zap.Error(err))
		return ListMemberPaymentsResponse{}, errors.Database("database operation", err)
	}

	logger.Info("payments retrieved successfully", zap.Int("count", len(payments)))

	return ListMemberPaymentsResponse{
		Payments: uc.toSummaries(payments),
	}, nil
}

// toSummaries converts payments to payment summaries.
func (uc *ListMemberPaymentsUseCase) toSummaries(payments []domain.Payment) []PaymentSummary {
	summaries := make([]PaymentSummary, len(payments))
	for i, p := range payments {
		var completedAt *string
		if p.CompletedAt != nil {
			completed := p.CompletedAt.Format("2006-01-02T15:04:05Z")
			completedAt = &completed
		}

		summaries[i] = PaymentSummary{
			ID:          p.ID,
			InvoiceID:   p.InvoiceID,
			Amount:      p.Amount,
			Currency:    p.Currency,
			Status:      p.Status,
			PaymentType: p.PaymentType,
			CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z"),
			CompletedAt: completedAt,
		}
	}
	return summaries
}
