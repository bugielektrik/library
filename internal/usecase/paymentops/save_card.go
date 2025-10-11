package paymentops

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/domain/payment"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
)

// SaveCardRequest represents the input for saving a card.
type SaveCardRequest struct {
	MemberID    string
	CardToken   string
	CardMask    string
	CardType    string
	ExpiryMonth int
	ExpiryYear  int
}

// SaveCardResponse represents the output of saving a card.
type SaveCardResponse struct {
	CardID      string
	CardMask    string
	CardType    string
	ExpiryMonth int
	ExpiryYear  int
	IsDefault   bool
}

// SaveCardUseCase handles saving a new card for a member.
type SaveCardUseCase struct {
	savedCardRepo payment.SavedCardRepository
}

// NewSaveCardUseCase creates a new instance of SaveCardUseCase.
func NewSaveCardUseCase(savedCardRepo payment.SavedCardRepository) *SaveCardUseCase {
	return &SaveCardUseCase{
		savedCardRepo: savedCardRepo,
	}
}

// Execute saves a new card for the member.
func (uc *SaveCardUseCase) Execute(ctx context.Context, req SaveCardRequest) (SaveCardResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "save_card")

	// Check if card with same token already exists
	existingCard, err := uc.savedCardRepo.GetByCardToken(ctx, req.CardToken)
	if err == nil && existingCard.ID != "" {
		logger.Warn("card already saved", zap.String("card_id", existingCard.ID))
		return SaveCardResponse{
			CardID:      existingCard.ID,
			CardMask:    existingCard.CardMask,
			CardType:    existingCard.CardType,
			ExpiryMonth: existingCard.ExpiryMonth,
			ExpiryYear:  existingCard.ExpiryYear,
			IsDefault:   existingCard.IsDefault,
		}, nil
	}

	// Create new saved card
	card := payment.NewSavedCard(
		req.MemberID,
		req.CardToken,
		req.CardMask,
		req.CardType,
		req.ExpiryMonth,
		req.ExpiryYear,
	)

	// Check if this is the first card (make it default)
	existingCards, err := uc.savedCardRepo.ListByMemberID(ctx, req.MemberID)
	if err != nil {
		logger.Error("failed to check existing cards", zap.Error(err))
		return SaveCardResponse{}, errors.Database("database operation", err)
	}

	if len(existingCards) == 0 {
		card.IsDefault = true
	}

	// Save to repository
	cardID, err := uc.savedCardRepo.Create(ctx, card)
	if err != nil {
		logger.Error("failed to save card", zap.Error(err))
		return SaveCardResponse{}, errors.Database("database operation", err)
	}

	logger.Info("card saved successfully", zap.String("card_id", cardID))

	return SaveCardResponse{
		CardID:      cardID,
		CardMask:    card.CardMask,
		CardType:    card.CardType,
		ExpiryMonth: card.ExpiryMonth,
		ExpiryYear:  card.ExpiryYear,
		IsDefault:   card.IsDefault,
	}, nil
}
