package paymentops

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/domain/payment"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
	"library-service/pkg/validation"
)

// InitiatePaymentRequest represents the input for initiating a payment.
type InitiatePaymentRequest struct {
	MemberID        string
	Amount          int64
	Currency        string
	PaymentType     payment.PaymentType
	RelatedEntityID *string
}

// Validate validates the InitiatePaymentRequest
func (r InitiatePaymentRequest) Validate() error {
	// Validate required fields
	if err := validation.RequiredString(r.MemberID, "MemberID"); err != nil {
		return err
	}

	// Validate amount range
	if err := validation.ValidateRange(r.Amount, "Amount", payment.MinPaymentAmount, payment.MaxPaymentAmount); err != nil {
		return err
	}

	// Validate currency
	if err := validation.RequiredString(r.Currency, "Currency"); err != nil {
		return err
	}

	// Validate currency is supported
	allowedCurrencies := []string{payment.CurrencyKZT, payment.CurrencyUSD, payment.CurrencyEUR, payment.CurrencyRUB}
	if err := validation.ValidateEnum(r.Currency, "Currency", allowedCurrencies); err != nil {
		return err
	}

	// Validate payment type
	if err := validation.RequiredString(string(r.PaymentType), "PaymentType"); err != nil {
		return err
	}

	return nil
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
//
// Architecture Pattern: Complex orchestration with external gateway integration.
// Demonstrates error recovery (updating payment status on gateway failure).
//
// See Also:
//   - Gateway interface: internal/domain/payment/gateway.go (payment.Gateway)
//   - Gateway impl: internal/adapters/payment/epayment/gateway.go
//   - Domain service: internal/domain/payment/service.go (validation, invoice generation)
//   - HTTP handler: internal/adapters/http/handlers/payment/initiate.go
//   - ADR: .claude/adr/005-payment-gateway-interface.md (gateway abstraction)
//   - ADR: .claude/adr/003-domain-services-vs-infrastructure.md (service types)
//   - Test: internal/usecase/paymentops/initiate_payment_test.go
type InitiatePaymentUseCase struct {
	paymentRepo    payment.Repository
	paymentService *payment.Service
	paymentGateway payment.Gateway
	gatewayConfig  payment.GatewayConfig
}

// NewInitiatePaymentUseCase creates a new instance of InitiatePaymentUseCase.
//
// The gateway parameter must implement both payment.Gateway (operations) and
// payment.GatewayConfig (configuration) interfaces.
func NewInitiatePaymentUseCase(
	paymentRepo payment.Repository,
	paymentService *payment.Service,
	gateway interface {
		payment.Gateway
		payment.GatewayConfig
	},
) *InitiatePaymentUseCase {
	return &InitiatePaymentUseCase{
		paymentRepo:    paymentRepo,
		paymentService: paymentService,
		paymentGateway: gateway,
		gatewayConfig:  gateway,
	}
}

// Execute initiates a new payment in the system.
func (uc *InitiatePaymentUseCase) Execute(ctx context.Context, req InitiatePaymentRequest) (InitiatePaymentResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "initiate")

	// Validate request
	if err := req.Validate(); err != nil {
		return InitiatePaymentResponse{}, err
	}

	// Generate unique invoice ID
	invoiceID := uc.paymentService.GenerateInvoiceID(req.MemberID, req.PaymentType)

	// Create payment entity from request
	paymentEntity := payment.New(payment.Request{
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
		return InitiatePaymentResponse{}, errors.Database("database operation", err)
	}
	paymentEntity.ID = paymentID

	// Get auth token from payment gateway
	authToken, err := uc.paymentGateway.GetAuthToken(ctx)
	if err != nil {
		logger.Error("failed to get auth token from payment gateway", zap.Error(err))

		// Update payment status to failed
		if updateErr := uc.paymentRepo.UpdateStatus(ctx, paymentID, payment.StatusFailed); updateErr != nil {
			logger.Error("failed to update payment status", zap.Error(updateErr))
		}

		return InitiatePaymentResponse{}, errors.External("payment gateway", err)
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
