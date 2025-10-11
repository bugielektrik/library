package payment

import (
	"fmt"
	"time"
)

// Receipt represents a payment receipt
type Receipt struct {
	ID            string
	PaymentID     string
	ReceiptNumber string // Human-readable receipt number (e.g., RCP-2024-00001)
	MemberID      string
	Amount        int64
	Currency      string
	PaymentType   PaymentType
	PaymentMethod PaymentMethod
	TransactionID string
	PaymentDate   time.Time
	ReceiptDate   time.Time
	Status        Status // Reflects payment status at receipt generation
	Description   string
	MemberName    string
	MemberEmail   string
	CardMask      *string
	Items         []ReceiptItem // Line items for the receipt
	TaxAmount     int64         // Tax amount if applicable
	TotalAmount   int64         // Total including tax
	Notes         string        // Additional notes
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ReceiptItem represents a line item on a receipt
type ReceiptItem struct {
	Description string
	Quantity    int
	UnitPrice   int64
	Amount      int64
}

// ReceiptRepository defines the interface for receipt persistence
type ReceiptRepository interface {
	Create(receipt Receipt) (string, error)
	GetByID(id string) (Receipt, error)
	GetByPaymentID(paymentID string) (Receipt, error)
	GetByReceiptNumber(receiptNumber string) (Receipt, error)
	ListByMemberID(memberID string) ([]Receipt, error)
	Update(receipt Receipt) error
}

// GenerateReceiptNumber generates a unique receipt number.
//
// Format: RCP-YYYY-NNNNN (e.g., RCP-2025-00001)
//
// The sequence number (NNNNN) is derived from the current timestamp in milliseconds,
// modulo 100000 to ensure it fits in 5 digits. This provides uniqueness within
// a reasonable time window while keeping the format readable.
func GenerateReceiptNumber() string {
	now := time.Now()
	// Use millisecond timestamp modulo 100000 for sequence (00000-99999)
	sequence := (now.UnixNano() / 1000000) % 100000
	// Format: RCP-YYYY-NNNNN
	return fmt.Sprintf("RCP-%s-%05d", now.Format("2006"), sequence)
}

// ReceiptData represents the data needed to generate a receipt
type ReceiptData struct {
	Payment     Payment
	MemberName  string
	MemberEmail string
	Items       []ReceiptItem
	Notes       string
}

// CreateReceiptFromPayment creates a receipt from payment data
func CreateReceiptFromPayment(data ReceiptData) Receipt {
	now := time.Now()

	// Calculate total (in this implementation, amount = total)
	totalAmount := data.Payment.Amount

	return Receipt{
		PaymentID:     data.Payment.ID,
		ReceiptNumber: GenerateReceiptNumber(),
		MemberID:      data.Payment.MemberID,
		Amount:        data.Payment.Amount,
		Currency:      data.Payment.Currency,
		PaymentType:   data.Payment.PaymentType,
		PaymentMethod: data.Payment.PaymentMethod,
		TransactionID: getStringValue(data.Payment.GatewayTransactionID),
		PaymentDate:   getTimeValue(data.Payment.CompletedAt, data.Payment.CreatedAt),
		ReceiptDate:   now,
		Status:        data.Payment.Status,
		Description:   getPaymentDescription(data.Payment.PaymentType),
		MemberName:    data.MemberName,
		MemberEmail:   data.MemberEmail,
		CardMask:      data.Payment.CardMask,
		Items:         data.Items,
		TaxAmount:     0, // No tax in current implementation
		TotalAmount:   totalAmount,
		Notes:         data.Notes,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func getTimeValue(ptr *time.Time, fallback time.Time) time.Time {
	if ptr == nil {
		return fallback
	}
	return *ptr
}

func getPaymentDescription(paymentType PaymentType) string {
	switch paymentType {
	case PaymentTypeFine:
		return "Library Fine Payment"
	case PaymentTypeSubscription:
		return "Library Subscription Payment"
	case PaymentTypeDeposit:
		return "Library Deposit Payment"
	default:
		return "Library Payment"
	}
}
