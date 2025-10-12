package payment

import (
	"context"
	errors2 "library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"

	"go.uber.org/zap"

	"library-service/internal/payments/domain"
)

// Validator defines the validation interface
type Validator interface {
	Validate(i interface{}) error
}

// InitiatePaymentRequest represents the input for initiating a domain.
type InitiatePaymentRequest struct {
	MemberID        string             `validate:"required"`
	Amount          int64              `validate:"required,min=100,max=10000000"`
	Currency        string             `validate:"required,oneof=KZT USD EUR RUB"`
	PaymentType     domain.PaymentType `validate:"required"`
	RelatedEntityID *string            `validate:"omitempty"`
}

// InitiatePaymentResponse represents the output of initiating a domain.
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

// InitiatePaymentUseCase handles the initiation of a new domain.
//
// Architecture Pattern: Complex orchestration with external provider integration.
// Demonstrates error recovery (updating payment status on provider failure).
//
// See Also:
//   - Gateway interface: internal/domain/payment/provider.go (domain.Gateway)
//   - Gateway impl: internal/adapters/payment/epayment/provider.go
//   - Domain service: internal/domain/payment/service.go (validation, invoice generation)
//   - HTTP handler: internal/adapters/http/handler/payment/initiate.go
//   - ADR: .claude/adr/005-payment-provider-interface.md (provider abstraction)
//   - ADR: .claude/adr/003-domain-service-vs-infrastructure.md (service types)
//   - Test: internal/usecase/paymentops/initiate_payment_test.go
type InitiatePaymentUseCase struct {
	paymentRepo    domain.Repository
	paymentService *domain.Service
	paymentGateway domain.Gateway
	gatewayConfig  domain.GatewayConfig
	validator      Validator
}

// NewInitiatePaymentUseCase creates a new instance of InitiatePaymentUseCase.
//
// The provider parameter must implement both domain.Gateway (operations) and
// domain.GatewayConfig (configuration) interfaces.
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

	// Get auth token from payment provider
	authToken, err := uc.paymentGateway.GetAuthToken(ctx)
	if err != nil {
		logger.Error("failed to get auth token from payment provider", zap.Error(err))

		// Update payment status to failed
		if updateErr := uc.paymentRepo.UpdateStatus(ctx, paymentID, domain.StatusFailed); updateErr != nil {
			logger.Error("failed to update payment status", zap.Error(updateErr))
		}

		return InitiatePaymentResponse{}, errors2.External("payment provider", err)
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
