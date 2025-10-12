package payment

import (
	"context"
	"encoding/json"

	"library-service/internal/payments/domain"
	errors2 "library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"

	"go.uber.org/zap"
)

// ================================================================================
// Save Card Use Case
// ================================================================================

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
	savedCardRepo domain.SavedCardRepository
}

// NewSaveCardUseCase creates a new instance of SaveCardUseCase.
func NewSaveCardUseCase(savedCardRepo domain.SavedCardRepository) *SaveCardUseCase {
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
	card := domain.NewSavedCard(
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
		return SaveCardResponse{}, errors2.Database("database operation", err)
	}

	if len(existingCards) == 0 {
		card.IsDefault = true
	}

	// Save to repository
	cardID, err := uc.savedCardRepo.Create(ctx, card)
	if err != nil {
		logger.Error("failed to save card", zap.Error(err))
		return SaveCardResponse{}, errors2.Database("database operation", err)
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

// ================================================================================
// Set Default Card Use Case
// ================================================================================

// SetDefaultCardRequest represents the input for setting a default card.
type SetDefaultCardRequest struct {
	CardID   string
	MemberID string
}

// SetDefaultCardResponse represents the output of setting a default card.
type SetDefaultCardResponse struct {
	Success bool
}

// SetDefaultCardUseCase handles setting a card as the default.
type SetDefaultCardUseCase struct {
	savedCardRepo domain.SavedCardRepository
}

// NewSetDefaultCardUseCase creates a new instance of SetDefaultCardUseCase.
func NewSetDefaultCardUseCase(savedCardRepo domain.SavedCardRepository) *SetDefaultCardUseCase {
	return &SetDefaultCardUseCase{
		savedCardRepo: savedCardRepo,
	}
}

// Execute sets a card as the default for a member.
func (uc *SetDefaultCardUseCase) Execute(ctx context.Context, req SetDefaultCardRequest) (SetDefaultCardResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "set_default_card")

	// Verify card belongs to member
	card, err := uc.savedCardRepo.GetByID(ctx, req.CardID)
	if err != nil {
		logger.Error("failed to retrieve card", zap.Error(err))
		return SetDefaultCardResponse{}, errors2.NotFoundWithID("card", req.CardID)
	}

	if card.MemberID != req.MemberID {
		logger.Warn("unauthorized set default attempt")
		return SetDefaultCardResponse{}, errors2.NotFoundWithID("card", req.CardID)
	}

	// Verify card can be used
	if !card.CanBeUsed() {
		logger.Warn("card cannot be set as default", zap.Bool("is_active", card.IsActive), zap.Bool("is_expired", card.IsExpired()))
		return SetDefaultCardResponse{}, errors2.NewError(errors2.CodeValidation).
			WithDetail("field", "card_id").
			WithDetail("reason", "card is inactive or expired").
			Build()
	}

	// Set as default
	if err := uc.savedCardRepo.SetAsDefault(ctx, req.MemberID, req.CardID); err != nil {
		logger.Error("failed to set card as default", zap.Error(err))
		return SetDefaultCardResponse{}, errors2.Database("database operation", err)
	}

	logger.Info("card set as default successfully")

	return SetDefaultCardResponse{
		Success: true,
	}, nil
}

// ================================================================================
// List Member Payments Use Case
// ================================================================================

// ListMemberPaymentsRequest represents the input for listing member payments.
type ListMemberPaymentsRequest struct {
	MemberID string
}

// PaymentSummary represents a summary of a payment.
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

// ListMemberPaymentsResponse represents the output of listing member payments.
type ListMemberPaymentsResponse struct {
	Payments []PaymentSummary
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

// Execute lists all payments for a member.
func (uc *ListMemberPaymentsUseCase) Execute(ctx context.Context, req ListMemberPaymentsRequest) (ListMemberPaymentsResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "list_member_payments")

	payments, err := uc.paymentRepo.ListByMemberID(ctx, req.MemberID)
	if err != nil {
		logger.Error("failed to list payments", zap.Error(err))
		return ListMemberPaymentsResponse{}, errors2.Database("list payments", err)
	}

	logger.Info("payments listed successfully", zap.Int("count", len(payments)))

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

// ================================================================================
// Helper Functions
// ================================================================================

// interfaceToMap converts an interface{} to a map[string]interface{} using JSON marshaling.
func interfaceToMap(data interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Unmarshal to map
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// UpdatePaymentFromGatewayResponse updates payment entity fields from a gateway response.
// This is a common pattern used across multiple payment use cases.
func UpdatePaymentFromGatewayResponse(
	paymentEntity *domain.Payment,
	transactionID string,
	approvalCode string,
	errorCode string,
	errorMessage string,
) {
	if transactionID != "" {
		paymentEntity.GatewayTransactionID = &transactionID
	}

	if approvalCode != "" {
		paymentEntity.ApprovalCode = &approvalCode
	}

	if errorCode != "" {
		paymentEntity.ErrorCode = &errorCode
	}

	if errorMessage != "" {
		paymentEntity.ErrorMessage = &errorMessage
	}
}

// UpdatePaymentFromCardCharge updates payment entity from a card charge response.
func UpdatePaymentFromCardCharge(
	paymentEntity *domain.Payment,
	gatewayResp *domain.CardChargeResponse,
	paymentService *domain.Service,
) {
	UpdatePaymentFromGatewayResponse(
		paymentEntity,
		gatewayResp.TransactionID,
		gatewayResp.ApprovalCode,
		gatewayResp.ErrorCode,
		gatewayResp.ErrorMessage,
	)

	paymentEntity.Status = paymentService.MapGatewayStatus(gatewayResp.Status)
}
