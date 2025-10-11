package paymentops

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/domain/payment"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
)

// ListSavedCardsRequest represents the input for listing saved cards.
type ListSavedCardsRequest struct {
	MemberID string
}

// ListSavedCardsResponse represents the output of listing saved cards.
type ListSavedCardsResponse struct {
	Cards []payment.SavedCard
}

// ListSavedCardsUseCase handles listing all saved cards for a member.
type ListSavedCardsUseCase struct {
	savedCardRepo payment.SavedCardRepository
}

// NewListSavedCardsUseCase creates a new instance of ListSavedCardsUseCase.
func NewListSavedCardsUseCase(savedCardRepo payment.SavedCardRepository) *ListSavedCardsUseCase {
	return &ListSavedCardsUseCase{
		savedCardRepo: savedCardRepo,
	}
}

// Execute lists all saved cards for a member.
func (uc *ListSavedCardsUseCase) Execute(ctx context.Context, req ListSavedCardsRequest) (ListSavedCardsResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "list_saved_cards")

	cards, err := uc.savedCardRepo.ListByMemberID(ctx, req.MemberID)
	if err != nil {
		logger.Error("failed to list saved cards", zap.Error(err))
		return ListSavedCardsResponse{}, errors.Database("database operation", err)
	}

	logger.Info("saved cards listed", zap.Int("count", len(cards)))

	return ListSavedCardsResponse{
		Cards: cards,
	}, nil
}
