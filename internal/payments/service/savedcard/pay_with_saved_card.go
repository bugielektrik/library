package savedcard

import (
	"context"
	errors2 "library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"
	"time"

	"go.uber.org/zap"

	"library-service/internal/payments/domain"
	paymentops "library-service/internal/payments/service/payment"
)

// PayWithSavedCardRequest represents the input for paying with a saved card.
type PayWithSavedCardRequest struct {
	MemberID        string
	SavedCardID     string
	Amount          int64
	Currency        string
	PaymentType     domain.PaymentType
	RelatedEntityID *string
}

// PayWithSavedCardResponse represents the output of paying with a saved card.
type PayWithSavedCardResponse struct {
	PaymentID string
	InvoiceID string
	Status    domain.Status
	Amount    int64
	Currency  string
	CardMask  string
}

// PayWithSavedCardUseCase handles payment using a saved card.
type PayWithSavedCardUseCase struct {
	paymentRepo    domain.Repository
	savedCardRepo  domain.SavedCardRepository
	paymentService *domain.Service
	paymentGateway domain.Gateway
}

// NewPayWithSavedCardUseCase creates a new instance of PayWithSavedCardUseCase.
func NewPayWithSavedCardUseCase(
	paymentRepo domain.Repository,
	savedCardRepo domain.SavedCardRepository,
	paymentService *domain.Service,
	paymentGateway domain.Gateway,
) *PayWithSavedCardUseCase {
	return &PayWithSavedCardUseCase{
		paymentRepo:    paymentRepo,
		savedCardRepo:  savedCardRepo,
		paymentService: paymentService,
		paymentGateway: paymentGateway,
	}
}

// validateSavedCard retrieves and validates a saved card for domain.
// Returns the card if valid, or an error if not found, unauthorized, or unusable.
func (uc *PayWithSavedCardUseCase) validateSavedCard(
	ctx context.Context,
	cardID string,
	memberID string,
	logger *zap.Logger,
) (domain.SavedCard, error) {
	// Retrieve saved card
	savedCard, err := uc.savedCardRepo.GetByID(ctx, cardID)
	if err != nil {
		logger.Error("failed to retrieve saved card", zap.Error(err))
		return domain.SavedCard{}, errors2.ErrNotFound.
			WithDetails("saved_card_id", cardID).
			Wrap(err)
	}

	// Verify card belongs to member
	if savedCard.MemberID != memberID {
		logger.Warn("unauthorized card usage attempt",
			zap.String("card_member_id", savedCard.MemberID),
			zap.String("requesting_member_id", memberID),
		)
		return domain.SavedCard{}, errors2.ErrNotFound.
			WithDetails("saved_card_id", cardID)
	}

	// Check if card can be used
	if !savedCard.CanBeUsed() {
		logger.Warn("card cannot be used",
			zap.Bool("is_active", savedCard.IsActive),
			zap.Bool("is_expired", savedCard.IsExpired()),
		)
		return domain.SavedCard{}, errors2.NewError(errors2.CodeValidation).
			WithDetail("field", "saved_card_id").
			WithDetail("reason", "card is inactive or expired").
			Build()
	}

	return savedCard, nil
}

// createPaymentRecord creates and validates a payment entity, then saves it to the repository.
// Returns the created payment entity with ID populated.
func (uc *PayWithSavedCardUseCase) createPaymentRecord(
	ctx context.Context,
	req PayWithSavedCardRequest,
	invoiceID string,
	cardMask string,
	logger *zap.Logger,
) (domain.Payment, error) {
	// Create payment entity
	paymentEntity := domain.New(domain.Request{
		InvoiceID:       invoiceID,
		MemberID:        req.MemberID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		PaymentType:     req.PaymentType,
		RelatedEntityID: req.RelatedEntityID,
	})

	// Set card information
	paymentEntity.PaymentMethod = domain.PaymentMethodCard
	paymentEntity.CardMask = &cardMask

	// Validate payment
	if err := uc.paymentService.Validate(paymentEntity); err != nil {
		logger.Warn("validation failed", zap.Error(err))
		return domain.Payment{}, err
	}

	// Save payment to repository
	paymentID, err := uc.paymentRepo.Create(ctx, paymentEntity)
	if err != nil {
		logger.Error("failed to create payment in repository", zap.Error(err))
		return domain.Payment{}, errors2.Database("database operation", err)
	}
	paymentEntity.ID = paymentID

	return paymentEntity, nil
}

// chargeCardViaGateway processes the payment through the provider and updates the payment entity.
// Returns the updated payment entity or an error if the charge fails.
func (uc *PayWithSavedCardUseCase) chargeCardViaGateway(
	ctx context.Context,
	paymentEntity domain.Payment,
	req PayWithSavedCardRequest,
	cardToken string,
	logger *zap.Logger,
) (domain.Payment, error) {
	logger.Info("charging card with token")

	// Prepare card charge request
	chargeReq := &domain.CardChargeRequest{
		InvoiceID:   paymentEntity.InvoiceID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		CardID:      cardToken,
		Description: string(req.PaymentType),
	}

	// Call provider
	gatewayResp, err := uc.paymentGateway.ChargeCard(ctx, chargeReq)
	if err != nil {
		logger.Error("failed to charge card via provider", zap.Error(err))
		// Update payment status to failed
		_ = uc.paymentRepo.UpdateStatus(ctx, paymentEntity.ID, domain.StatusFailed)
		return domain.Payment{}, errors2.External("payment provider", err)
	}

	// Update payment with provider response
	paymentops.UpdatePaymentFromCardCharge(&paymentEntity, gatewayResp, uc.paymentService)

	logger.Info("card charged successfully via provider",
		zap.String("payment_id", paymentEntity.ID),
		zap.String("status", string(paymentEntity.Status)),
	)

	// Update payment in repository
	if err := uc.paymentRepo.Update(ctx, paymentEntity.ID, paymentEntity); err != nil {
		logger.Error("failed to update payment", zap.Error(err))
		return domain.Payment{}, errors2.Database("database operation", err)
	}

	return paymentEntity, nil
}

// updateCardLastUsed updates the last used timestamp for the saved card.
// This is a best-effort operation - failures are logged but don't affect the domain.
func (uc *PayWithSavedCardUseCase) updateCardLastUsed(
	ctx context.Context,
	card domain.SavedCard,
	logger *zap.Logger,
) {
	now := time.Now()
	card.LastUsedAt = &now
	if err := uc.savedCardRepo.Update(ctx, card.ID, card); err != nil {
		logger.Warn("failed to update card last used timestamp", zap.Error(err))
	}
}

// Execute processes a payment using a saved card.
func (uc *PayWithSavedCardUseCase) Execute(ctx context.Context, req PayWithSavedCardRequest) (PayWithSavedCardResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "pay_with_saved_card")

	// Step 1: Validate and retrieve saved card
	savedCard, err := uc.validateSavedCard(ctx, req.SavedCardID, req.MemberID, logger)
	if err != nil {
		return PayWithSavedCardResponse{}, err
	}

	// Step 2: Generate invoice ID and create payment record
	invoiceID := uc.paymentService.GenerateInvoiceID(req.MemberID, req.PaymentType)
	paymentEntity, err := uc.createPaymentRecord(ctx, req, invoiceID, savedCard.CardMask, logger)
	if err != nil {
		return PayWithSavedCardResponse{}, err
	}

	// Step 3: Charge card via provider
	paymentEntity, err = uc.chargeCardViaGateway(ctx, paymentEntity, req, savedCard.CardToken, logger)
	if err != nil {
		return PayWithSavedCardResponse{}, err
	}

	// Step 4: Update card last used timestamp (best effort)
	uc.updateCardLastUsed(ctx, savedCard, logger)

	logger.Info("payment initiated with saved card",
		zap.String("payment_id", paymentEntity.ID),
		zap.String("saved_card_id", req.SavedCardID),
	)

	return PayWithSavedCardResponse{
		PaymentID: paymentEntity.ID,
		InvoiceID: invoiceID,
		Status:    paymentEntity.Status,
		Amount:    req.Amount,
		Currency:  req.Currency,
		CardMask:  savedCard.CardMask,
	}, nil
}
