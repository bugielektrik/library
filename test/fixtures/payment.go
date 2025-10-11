package fixtures

import (
	"time"

	"library-service/internal/domain/payment"
)

// stringPtr is a helper to create string pointers
func stringPtr(s string) *string {
	return &s
}

// timePtr is a helper to create time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}

// PendingPayment returns a payment in pending status (just initiated)
func PendingPayment() payment.Payment {
	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)

	return payment.Payment{
		ID:                   "pay_550e8400-e29b-41d4-a716-446655440000",
		InvoiceID:            "INV_2024_001",
		MemberID:             "b4101570-0a35-4dd3-b8f7-745d56013263",
		Amount:               5000, // 50.00 KZT
		Currency:             "KZT",
		Status:               payment.StatusPending,
		PaymentMethod:        payment.PaymentMethodCard,
		PaymentType:          payment.PaymentTypeFine,
		RelatedEntityID:      stringPtr("fine_123"),
		GatewayTransactionID: nil,
		GatewayResponse:      nil,
		CardMask:             nil,
		ApprovalCode:         nil,
		ErrorCode:            nil,
		ErrorMessage:         nil,
		CreatedAt:            now,
		UpdatedAt:            now,
		CompletedAt:          nil,
		ExpiresAt:            expiresAt,
	}
}

// ProcessingPayment returns a payment in processing status (gateway is processing)
func ProcessingPayment() payment.Payment {
	now := time.Now()
	createdAt := now.Add(-5 * time.Minute)
	expiresAt := createdAt.Add(30 * time.Minute)

	return payment.Payment{
		ID:                   "pay_550e8400-e29b-41d4-a716-446655440001",
		InvoiceID:            "INV_2024_002",
		MemberID:             "b4101570-0a35-4dd3-b8f7-745d56013263",
		Amount:               15000, // 150.00 KZT
		Currency:             "KZT",
		Status:               payment.StatusProcessing,
		PaymentMethod:        payment.PaymentMethodCard,
		PaymentType:          payment.PaymentTypeSubscription,
		RelatedEntityID:      stringPtr("sub_456"),
		GatewayTransactionID: stringPtr("GW_TX_789ABC"),
		GatewayResponse:      stringPtr(`{"status":"processing","transaction_id":"GW_TX_789ABC"}`),
		CardMask:             stringPtr("4405-62**-****-1448"),
		ApprovalCode:         nil,
		ErrorCode:            nil,
		ErrorMessage:         nil,
		CreatedAt:            createdAt,
		UpdatedAt:            now,
		CompletedAt:          nil,
		ExpiresAt:            expiresAt,
	}
}

// CompletedPayment returns a successfully completed payment
func CompletedPayment() payment.Payment {
	now := time.Now()
	createdAt := now.Add(-1 * time.Hour)
	completedAt := now.Add(-55 * time.Minute)
	expiresAt := createdAt.Add(30 * time.Minute)

	return payment.Payment{
		ID:                   "pay_550e8400-e29b-41d4-a716-446655440002",
		InvoiceID:            "INV_2024_003",
		MemberID:             "b4101570-0a35-4dd3-b8f7-745d56013263",
		Amount:               10000, // 100.00 KZT
		Currency:             "KZT",
		Status:               payment.StatusCompleted,
		PaymentMethod:        payment.PaymentMethodCard,
		PaymentType:          payment.PaymentTypeFine,
		RelatedEntityID:      stringPtr("fine_789"),
		GatewayTransactionID: stringPtr("GW_TX_COMPLETED_123"),
		GatewayResponse:      stringPtr(`{"status":"success","transaction_id":"GW_TX_COMPLETED_123","approval_code":"ABC123"}`),
		CardMask:             stringPtr("4405-62**-****-1448"),
		ApprovalCode:         stringPtr("ABC123"),
		ErrorCode:            nil,
		ErrorMessage:         nil,
		CreatedAt:            createdAt,
		UpdatedAt:            completedAt,
		CompletedAt:          timePtr(completedAt),
		ExpiresAt:            expiresAt,
	}
}

// FailedPayment returns a failed payment with error details
func FailedPayment() payment.Payment {
	now := time.Now()
	createdAt := now.Add(-2 * time.Hour)
	completedAt := now.Add(-115 * time.Minute)
	expiresAt := createdAt.Add(30 * time.Minute)

	return payment.Payment{
		ID:                   "pay_550e8400-e29b-41d4-a716-446655440003",
		InvoiceID:            "INV_2024_004",
		MemberID:             "b4101570-0a35-4dd3-b8f7-745d56013263",
		Amount:               7500, // 75.00 KZT
		Currency:             "KZT",
		Status:               payment.StatusFailed,
		PaymentMethod:        payment.PaymentMethodCard,
		PaymentType:          payment.PaymentTypeDeposit,
		RelatedEntityID:      nil,
		GatewayTransactionID: stringPtr("GW_TX_FAILED_456"),
		GatewayResponse:      stringPtr(`{"status":"failed","error_code":"INSUFFICIENT_FUNDS","message":"Insufficient funds"}`),
		CardMask:             stringPtr("4405-62**-****-1448"),
		ApprovalCode:         nil,
		ErrorCode:            stringPtr("INSUFFICIENT_FUNDS"),
		ErrorMessage:         stringPtr("Insufficient funds"),
		CreatedAt:            createdAt,
		UpdatedAt:            completedAt,
		CompletedAt:          timePtr(completedAt),
		ExpiresAt:            expiresAt,
	}
}

// CancelledPayment returns a cancelled payment
func CancelledPayment() payment.Payment {
	now := time.Now()
	createdAt := now.Add(-3 * time.Hour)
	cancelledAt := now.Add(-175 * time.Minute)
	expiresAt := createdAt.Add(30 * time.Minute)

	return payment.Payment{
		ID:                   "pay_550e8400-e29b-41d4-a716-446655440004",
		InvoiceID:            "INV_2024_005",
		MemberID:             "b4101570-0a35-4dd3-b8f7-745d56013263",
		Amount:               20000, // 200.00 KZT
		Currency:             "KZT",
		Status:               payment.StatusCancelled,
		PaymentMethod:        payment.PaymentMethodCard,
		PaymentType:          payment.PaymentTypeSubscription,
		RelatedEntityID:      stringPtr("sub_999"),
		GatewayTransactionID: stringPtr("GW_TX_CANCELLED_789"),
		GatewayResponse:      stringPtr(`{"status":"cancelled","reason":"user_cancelled"}`),
		CardMask:             nil,
		ApprovalCode:         nil,
		ErrorCode:            stringPtr("USER_CANCELLED"),
		ErrorMessage:         stringPtr("Payment cancelled by user"),
		CreatedAt:            createdAt,
		UpdatedAt:            cancelledAt,
		CompletedAt:          timePtr(cancelledAt),
		ExpiresAt:            expiresAt,
	}
}

// RefundedPayment returns a refunded payment
func RefundedPayment() payment.Payment {
	now := time.Now()
	createdAt := now.Add(-24 * time.Hour)
	completedAt := now.Add(-23 * time.Hour)
	refundedAt := now.Add(-1 * time.Hour)
	expiresAt := createdAt.Add(30 * time.Minute)

	return payment.Payment{
		ID:                   "pay_550e8400-e29b-41d4-a716-446655440005",
		InvoiceID:            "INV_2024_006",
		MemberID:             "b4101570-0a35-4dd3-b8f7-745d56013263",
		Amount:               12500, // 125.00 KZT
		Currency:             "KZT",
		Status:               payment.StatusRefunded,
		PaymentMethod:        payment.PaymentMethodCard,
		PaymentType:          payment.PaymentTypeFine,
		RelatedEntityID:      stringPtr("fine_111"),
		GatewayTransactionID: stringPtr("GW_TX_REFUNDED_321"),
		GatewayResponse:      stringPtr(`{"status":"refunded","refund_transaction_id":"GW_REFUND_654"}`),
		CardMask:             stringPtr("5536-91**-****-2847"),
		ApprovalCode:         stringPtr("XYZ789"),
		ErrorCode:            nil,
		ErrorMessage:         nil,
		CreatedAt:            createdAt,
		UpdatedAt:            refundedAt,
		CompletedAt:          timePtr(completedAt),
		ExpiresAt:            expiresAt,
	}
}

// ExpiredPayment returns an expired payment (not completed within time limit)
func ExpiredPayment() payment.Payment {
	now := time.Now()
	createdAt := now.Add(-2 * time.Hour)
	expiresAt := now.Add(-90 * time.Minute)

	return payment.Payment{
		ID:                   "pay_550e8400-e29b-41d4-a716-446655440006",
		InvoiceID:            "INV_2024_007",
		MemberID:             "b4101570-0a35-4dd3-b8f7-745d56013263",
		Amount:               3000, // 30.00 KZT
		Currency:             "KZT",
		Status:               payment.StatusPending,
		PaymentMethod:        payment.PaymentMethodCard,
		PaymentType:          payment.PaymentTypeDeposit,
		RelatedEntityID:      nil,
		GatewayTransactionID: nil,
		GatewayResponse:      nil,
		CardMask:             nil,
		ApprovalCode:         nil,
		ErrorCode:            nil,
		ErrorMessage:         nil,
		CreatedAt:            createdAt,
		UpdatedAt:            createdAt,
		CompletedAt:          nil,
		ExpiresAt:            expiresAt,
	}
}

// WalletPayment returns a payment made with an e-wallet
func WalletPayment() payment.Payment {
	now := time.Now()
	createdAt := now.Add(-10 * time.Minute)
	completedAt := now.Add(-8 * time.Minute)
	expiresAt := createdAt.Add(30 * time.Minute)

	return payment.Payment{
		ID:                   "pay_550e8400-e29b-41d4-a716-446655440007",
		InvoiceID:            "INV_2024_008",
		MemberID:             "c4101570-0a35-4dd3-b8f7-745d56013265",
		Amount:               25000, // 250.00 KZT
		Currency:             "KZT",
		Status:               payment.StatusCompleted,
		PaymentMethod:        payment.PaymentMethodWallet,
		PaymentType:          payment.PaymentTypeSubscription,
		RelatedEntityID:      stringPtr("sub_777"),
		GatewayTransactionID: stringPtr("GW_TX_WALLET_555"),
		GatewayResponse:      stringPtr(`{"status":"success","wallet":"kaspi","transaction_id":"GW_TX_WALLET_555"}`),
		CardMask:             nil,
		ApprovalCode:         stringPtr("WALLET_APPROVED"),
		ErrorCode:            nil,
		ErrorMessage:         nil,
		CreatedAt:            createdAt,
		UpdatedAt:            completedAt,
		CompletedAt:          timePtr(completedAt),
		ExpiresAt:            expiresAt,
	}
}

// HighValuePayment returns a payment with a high amount
func HighValuePayment() payment.Payment {
	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)

	return payment.Payment{
		ID:                   "pay_550e8400-e29b-41d4-a716-446655440008",
		InvoiceID:            "INV_2024_009",
		MemberID:             "a4101570-0a35-4dd3-b8f7-745d56013264",
		Amount:               500000, // 5000.00 KZT
		Currency:             "KZT",
		Status:               payment.StatusPending,
		PaymentMethod:        payment.PaymentMethodCard,
		PaymentType:          payment.PaymentTypeSubscription,
		RelatedEntityID:      stringPtr("sub_premium_001"),
		GatewayTransactionID: nil,
		GatewayResponse:      nil,
		CardMask:             nil,
		ApprovalCode:         nil,
		ErrorCode:            nil,
		ErrorMessage:         nil,
		CreatedAt:            now,
		UpdatedAt:            now,
		CompletedAt:          nil,
		ExpiresAt:            expiresAt,
	}
}

// PaymentsList returns a slice of payments for testing list operations
func PaymentsList() []payment.Payment {
	return []payment.Payment{
		CompletedPayment(),
		PendingPayment(),
		FailedPayment(),
		RefundedPayment(),
	}
}

// PaymentWithMinimalData returns a payment with only required fields
func PaymentWithMinimalData() payment.Payment {
	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)

	return payment.Payment{
		ID:                   "pay_550e8400-e29b-41d4-a716-446655440009",
		InvoiceID:            "INV_MINIMAL_001",
		MemberID:             "b4101570-0a35-4dd3-b8f7-745d56013263",
		Amount:               1000,
		Currency:             "KZT",
		Status:               payment.StatusPending,
		PaymentMethod:        payment.PaymentMethodCard,
		PaymentType:          payment.PaymentTypeFine,
		RelatedEntityID:      nil,
		GatewayTransactionID: nil,
		GatewayResponse:      nil,
		CardMask:             nil,
		ApprovalCode:         nil,
		ErrorCode:            nil,
		ErrorMessage:         nil,
		CreatedAt:            now,
		UpdatedAt:            now,
		CompletedAt:          nil,
		ExpiresAt:            expiresAt,
	}
}

// PaymentForCreate returns a payment entity suitable for repository creation (no ID)
func PaymentForCreate() payment.Payment {
	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)

	return payment.Payment{
		InvoiceID:     "INV_TEST_NEW",
		MemberID:      "b4101570-0a35-4dd3-b8f7-745d56013263",
		Amount:        10000,
		Currency:      "KZT",
		Status:        payment.StatusPending,
		PaymentMethod: payment.PaymentMethodCard,
		PaymentType:   payment.PaymentTypeFine,
		CreatedAt:     now,
		UpdatedAt:     now,
		ExpiresAt:     expiresAt,
	}
}

// Payments is an alias for PaymentsList for integration tests
func Payments() []payment.Payment {
	return PaymentsList()
}
