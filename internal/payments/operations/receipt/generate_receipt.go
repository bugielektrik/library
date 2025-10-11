package receipt

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	memberdomain "library-service/internal/members/domain"
	"library-service/internal/payments/domain"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
)

// GenerateReceiptRequest represents the input for generating a receipt.
type GenerateReceiptRequest struct {
	PaymentID string
	MemberID  string // For authorization
	Notes     string // Optional notes to add to receipt
}

// GenerateReceiptResponse represents the output of generating a receipt.
type GenerateReceiptResponse struct {
	ReceiptID     string
	ReceiptNumber string
	PaymentID     string
	Amount        int64
	Currency      string
	ReceiptDate   string
}

// GenerateReceiptUseCase handles receipt generation for completed payments.
type GenerateReceiptUseCase struct {
	paymentRepo domain.Repository
	receiptRepo domain.ReceiptRepository
	memberRepo  memberdomain.Repository
}

// NewGenerateReceiptUseCase creates a new instance of GenerateReceiptUseCase.
func NewGenerateReceiptUseCase(
	paymentRepo domain.Repository,
	receiptRepo domain.ReceiptRepository,
	memberRepo memberdomain.Repository,
) *GenerateReceiptUseCase {
	return &GenerateReceiptUseCase{
		paymentRepo: paymentRepo,
		receiptRepo: receiptRepo,
		memberRepo:  memberRepo,
	}
}

// Execute generates a receipt for a completed domain.
func (uc *GenerateReceiptUseCase) Execute(ctx context.Context, req GenerateReceiptRequest) (GenerateReceiptResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "generate_receipt")

	logger.Info("generating receipt for payment")

	// Get payment
	paymentEntity, err := uc.paymentRepo.GetByID(ctx, req.PaymentID)
	if err != nil {
		logger.Error("failed to get payment", zap.Error(err))
		return GenerateReceiptResponse{}, errors.NotFoundWithID("payment", req.PaymentID)
	}

	// Verify payment belongs to member
	if paymentEntity.MemberID != req.MemberID {
		logger.Warn("unauthorized receipt generation attempt",
			zap.String("payment_member_id", paymentEntity.MemberID),
			zap.String("requesting_member_id", req.MemberID),
		)
		return GenerateReceiptResponse{}, errors.Unauthorized("invalid credentials")
	}

	// Verify payment is completed
	if paymentEntity.Status != domain.StatusCompleted {
		logger.Warn("cannot generate receipt for non-completed payment",
			zap.String("payment_status", string(paymentEntity.Status)),
		)
		return GenerateReceiptResponse{}, errors.NewError(errors.CodeValidation).
			WithDetail("field", "payment_status").
			WithDetail("reason", "payment must be completed to generate receipt").
			Build()
	}

	// Check if receipt already exists for this payment
	existingReceipt, err := uc.receiptRepo.GetByPaymentID(req.PaymentID)
	if err == nil {
		// Receipt already exists, return it
		logger.Info("receipt already exists for payment",
			zap.String("receipt_id", existingReceipt.ID),
			zap.String("receipt_number", existingReceipt.ReceiptNumber),
		)
		return GenerateReceiptResponse{
			ReceiptID:     existingReceipt.ID,
			ReceiptNumber: existingReceipt.ReceiptNumber,
			PaymentID:     existingReceipt.PaymentID,
			Amount:        existingReceipt.Amount,
			Currency:      existingReceipt.Currency,
			ReceiptDate:   existingReceipt.ReceiptDate.Format("2006-01-02 15:04:05"),
		}, nil
	}

	// Get member details
	memberEntity, err := uc.memberRepo.Get(ctx, req.MemberID)
	if err != nil {
		logger.Error("failed to get member", zap.Error(err))
		return GenerateReceiptResponse{}, errors.NotFoundWithID("member", req.MemberID)
	}

	// Create receipt items based on payment type
	items := createReceiptItems(paymentEntity)

	// Get member name (handle pointer)
	memberName := ""
	if memberEntity.FullName != nil {
		memberName = *memberEntity.FullName
	}

	// Create receipt
	receiptData := domain.ReceiptData{
		Payment:     paymentEntity,
		MemberName:  memberName,
		MemberEmail: memberEntity.Email,
		Items:       items,
		Notes:       req.Notes,
	}

	receipt := domain.CreateReceiptFromPayment(receiptData)
	receipt.ID = uuid.New().String()

	// Save receipt
	receiptID, err := uc.receiptRepo.Create(receipt)
	if err != nil {
		logger.Error("failed to create receipt", zap.Error(err))
		return GenerateReceiptResponse{}, errors.Database("database operation", err)
	}

	logger.Info("receipt generated successfully",
		zap.String("receipt_id", receiptID),
		zap.String("receipt_number", receipt.ReceiptNumber),
	)

	return GenerateReceiptResponse{
		ReceiptID:     receiptID,
		ReceiptNumber: receipt.ReceiptNumber,
		PaymentID:     receipt.PaymentID,
		Amount:        receipt.Amount,
		Currency:      receipt.Currency,
		ReceiptDate:   receipt.ReceiptDate.Format("2006-01-02 15:04:05"),
	}, nil
}

// createReceiptItems creates receipt line items based on payment type
func createReceiptItems(paymentEntity domain.Payment) []domain.ReceiptItem {
	var description string
	switch paymentEntity.PaymentType {
	case domain.PaymentTypeFine:
		description = "Library Fine"
	case domain.PaymentTypeSubscription:
		description = "Library Subscription"
	case domain.PaymentTypeDeposit:
		description = "Library Deposit"
	default:
		description = "Library Service"
	}

	return []domain.ReceiptItem{
		{
			Description: description,
			Quantity:    1,
			UnitPrice:   paymentEntity.Amount,
			Amount:      paymentEntity.Amount,
		},
	}
}
