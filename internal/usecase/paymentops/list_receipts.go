package paymentops

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/domain/payment"
	"library-service/pkg/logutil"
)

// ListReceiptsRequest represents the input for listing receipts.
type ListReceiptsRequest struct {
	MemberID string
}

// ListReceiptsResponse represents the output of listing receipts.
type ListReceiptsResponse struct {
	Receipts []payment.Receipt
	Total    int
}

// ListReceiptsUseCase handles listing receipts for a member.
type ListReceiptsUseCase struct {
	receiptRepo payment.ReceiptRepository
}

// NewListReceiptsUseCase creates a new instance of ListReceiptsUseCase.
func NewListReceiptsUseCase(receiptRepo payment.ReceiptRepository) *ListReceiptsUseCase {
	return &ListReceiptsUseCase{
		receiptRepo: receiptRepo,
	}
}

// Execute lists all receipts for a member.
func (uc *ListReceiptsUseCase) Execute(ctx context.Context, req ListReceiptsRequest) (ListReceiptsResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "list_receipts")

	logger.Info("listing receipts for member")

	receipts, err := uc.receiptRepo.ListByMemberID(req.MemberID)
	if err != nil {
		logger.Error("failed to list receipts", zap.Error(err))
		return ListReceiptsResponse{}, err
	}

	logger.Info("receipts listed successfully", zap.Int("count", len(receipts)))

	return ListReceiptsResponse{
		Receipts: receipts,
		Total:    len(receipts),
	}, nil
}
