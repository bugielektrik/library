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
