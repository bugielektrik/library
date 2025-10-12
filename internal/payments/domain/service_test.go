package domain

import (
	"strings"
	"testing"
	"time"
)

func TestService_Validate(t *testing.T) {
	service := NewService()

	tests := []struct {
		name    string
		payment Payment
		wantErr bool
	}{
		{
			name: "valid payment",
			payment: Payment{
				MemberID:    "member-123",
				InvoiceID:   "invoice-123",
				Amount:      10000,
				Currency:    "KZT",
				PaymentType: PaymentTypeFine,
			},
			wantErr: false,
		},
		{
			name: "invalid payment - empty member ID",
			payment: Payment{
				MemberID:    "",
				InvoiceID:   "invoice-123",
				Amount:      10000,
				Currency:    "KZT",
				PaymentType: PaymentTypeFine,
			},
			wantErr: true,
		},
		{
			name: "invalid payment - zero amount",
			payment: Payment{
				MemberID:    "member-123",
				InvoiceID:   "invoice-123",
				Amount:      0,
				Currency:    "KZT",
				PaymentType: PaymentTypeFine,
			},
			wantErr: true,
		},
		{
			name: "invalid payment - invalid currency",
			payment: Payment{
				MemberID:    "member-123",
				InvoiceID:   "invoice-123",
				Amount:      10000,
				Currency:    "TTT",
				PaymentType: PaymentTypeFine,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Validate(tt.payment)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_ValidateStatusTransition(t *testing.T) {
	service := NewService()

	tests := []struct {
		name          string
		currentStatus Status
		newStatus     Status
		wantErr       bool
	}{
		{"pending to processing", StatusPending, StatusProcessing, false},
		{"pending to cancelled", StatusPending, StatusCancelled, false},
		{"processing to completed", StatusProcessing, StatusCompleted, false},
		{"completed to refunded", StatusCompleted, StatusRefunded, false},
		{"completed to pending - invalid", StatusCompleted, StatusPending, true},
		{"cancelled to completed - invalid", StatusCancelled, StatusCompleted, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateStatusTransition(tt.currentStatus, tt.newStatus)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStatusTransition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_GenerateInvoiceID(t *testing.T) {
	service := NewService()

	memberID := "member-123"
	paymentType := PaymentTypeFine

	invoiceID := service.GenerateInvoiceID(memberID, paymentType)

	if invoiceID == "" {
		t.Error("GenerateInvoiceID() returned empty string")
	}

	// Invoice ID should start with payment type
	expectedPrefix := string(paymentType) + "-"
	if len(invoiceID) < len(expectedPrefix) {
		t.Errorf("GenerateInvoiceID() = %v, expected to start with %v", invoiceID, expectedPrefix)
	}
}

func TestService_isValidCurrency(t *testing.T) {
	service := NewService()

	tests := []struct {
		name     string
		currency string
		want     bool
	}{
		{"valid KZT", "KZT", true},
		{"valid USD", "USD", true},
		{"valid EUR", "EUR", true},
		{"valid RUB", "RUB", true},
		{"invalid TTT", "TTT", false},
		{"invalid ABC", "ABC", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := service.isValidCurrency(tt.currency); got != tt.want {
				t.Errorf("isValidCurrency() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestService_Validate_Comprehensive tests all validation scenarios
func TestService_Validate_Comprehensive(t *testing.T) {
	service := NewService()

	tests := []struct {
		name    string
		payment Payment
		wantErr bool
	}{
		{
			name: "valid payment - all fields",
			payment: Payment{
				MemberID:    "member-123",
				InvoiceID:   "invoice-123",
				Amount:      10000,
				Currency:    "KZT",
				PaymentType: PaymentTypeFine,
			},
			wantErr: false,
		},
		{
			name: "missing member ID",
			payment: Payment{
				MemberID:    "",
				InvoiceID:   "invoice-123",
				Amount:      10000,
				Currency:    "KZT",
				PaymentType: PaymentTypeFine,
			},
			wantErr: true,
		},
		{
			name: "missing invoice ID",
			payment: Payment{
				MemberID:    "member-123",
				InvoiceID:   "",
				Amount:      10000,
				Currency:    "KZT",
				PaymentType: PaymentTypeFine,
			},
			wantErr: true,
		},
		{
			name: "zero amount",
			payment: Payment{
				MemberID:    "member-123",
				InvoiceID:   "invoice-123",
				Amount:      0,
				Currency:    "KZT",
				PaymentType: PaymentTypeFine,
			},
			wantErr: true,
		},
		{
			name: "negative amount",
			payment: Payment{
				MemberID:    "member-123",
				InvoiceID:   "invoice-123",
				Amount:      -5000,
				Currency:    "KZT",
				PaymentType: PaymentTypeFine,
			},
			wantErr: true,
		},
		{
			name: "missing currency",
			payment: Payment{
				MemberID:    "member-123",
				InvoiceID:   "invoice-123",
				Amount:      10000,
				Currency:    "",
				PaymentType: PaymentTypeFine,
			},
			wantErr: true,
		},
		{
			name: "invalid currency code",
			payment: Payment{
				MemberID:    "member-123",
				InvoiceID:   "invoice-123",
				Amount:      10000,
				Currency:    "INVALID",
				PaymentType: PaymentTypeFine,
			},
			wantErr: true,
		},
		{
			name: "invalid payment type",
			payment: Payment{
				MemberID:    "member-123",
				InvoiceID:   "invoice-123",
				Amount:      10000,
				Currency:    "KZT",
				PaymentType: "invalid_type",
			},
			wantErr: true,
		},
		{
			name: "valid subscription payment",
			payment: Payment{
				MemberID:    "member-456",
				InvoiceID:   "invoice-456",
				Amount:      25000,
				Currency:    "USD",
				PaymentType: PaymentTypeSubscription,
			},
			wantErr: false,
		},
		{
			name: "valid deposit payment",
			payment: Payment{
				MemberID:    "member-789",
				InvoiceID:   "invoice-789",
				Amount:      50000,
				Currency:    "EUR",
				PaymentType: PaymentTypeDeposit,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Validate(tt.payment)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestService_ValidateStatusTransition_Comprehensive tests all status transitions
func TestService_ValidateStatusTransition_Comprehensive(t *testing.T) {
	service := NewService()

	tests := []struct {
		name          string
		currentStatus Status
		newStatus     Status
		wantErr       bool
	}{
		// Valid transitions from Pending
		{"pending → processing", StatusPending, StatusProcessing, false},
		{"pending → cancelled", StatusPending, StatusCancelled, false},
		{"pending → failed", StatusPending, StatusFailed, false},

		// Invalid transitions from Pending
		{"pending → completed (invalid)", StatusPending, StatusCompleted, true},
		{"pending → refunded (invalid)", StatusPending, StatusRefunded, true},

		// Valid transitions from Processing
		{"processing → completed", StatusProcessing, StatusCompleted, false},
		{"processing → failed", StatusProcessing, StatusFailed, false},
		{"processing → cancelled", StatusProcessing, StatusCancelled, false},

		// Invalid transitions from Processing
		{"processing → pending (invalid)", StatusProcessing, StatusPending, true},
		{"processing → refunded (invalid)", StatusProcessing, StatusRefunded, true},

		// Valid transitions from Completed
		{"completed → refunded", StatusCompleted, StatusRefunded, false},

		// Invalid transitions from Completed
		{"completed → pending (invalid)", StatusCompleted, StatusPending, true},
		{"completed → processing (invalid)", StatusCompleted, StatusProcessing, true},
		{"completed → failed (invalid)", StatusCompleted, StatusFailed, true},
		{"completed → cancelled (invalid)", StatusCompleted, StatusCancelled, true},

		// Valid transitions from Failed
		{"failed → pending (retry)", StatusFailed, StatusPending, false},

		// Invalid transitions from Failed
		{"failed → processing (invalid)", StatusFailed, StatusProcessing, true},
		{"failed → completed (invalid)", StatusFailed, StatusCompleted, true},
		{"failed → refunded (invalid)", StatusFailed, StatusRefunded, true},
		{"failed → cancelled (invalid)", StatusFailed, StatusCancelled, true},

		// No valid transitions from Cancelled (terminal state)
		{"cancelled → pending (invalid)", StatusCancelled, StatusPending, true},
		{"cancelled → processing (invalid)", StatusCancelled, StatusProcessing, true},
		{"cancelled → completed (invalid)", StatusCancelled, StatusCompleted, true},
		{"cancelled → failed (invalid)", StatusCancelled, StatusFailed, true},
		{"cancelled → refunded (invalid)", StatusCancelled, StatusRefunded, true},

		// No valid transitions from Refunded (terminal state)
		{"refunded → pending (invalid)", StatusRefunded, StatusPending, true},
		{"refunded → processing (invalid)", StatusRefunded, StatusProcessing, true},
		{"refunded → completed (invalid)", StatusRefunded, StatusCompleted, true},
		{"refunded → failed (invalid)", StatusRefunded, StatusFailed, true},
		{"refunded → cancelled (invalid)", StatusRefunded, StatusCancelled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateStatusTransition(tt.currentStatus, tt.newStatus)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStatusTransition(%s, %s) error = %v, wantErr %v",
					tt.currentStatus, tt.newStatus, err, tt.wantErr)
			}
		})
	}
}

// TestService_IsExpired tests payment expiration logic
func TestService_IsExpired(t *testing.T) {
	service := NewService()

	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	tests := []struct {
		name     string
		payment  Payment
		expected bool
	}{
		{
			name: "pending payment - not expired",
			payment: Payment{
				Status:    StatusPending,
				ExpiresAt: future,
			},
			expected: false,
		},
		{
			name: "pending payment - expired",
			payment: Payment{
				Status:    StatusPending,
				ExpiresAt: past,
			},
			expected: true,
		},
		{
			name: "processing payment - expired",
			payment: Payment{
				Status:    StatusProcessing,
				ExpiresAt: past,
			},
			expected: true,
		},
		{
			name: "completed payment - expired time passed (should not matter)",
			payment: Payment{
				Status:    StatusCompleted,
				ExpiresAt: past,
			},
			expected: false,
		},
		{
			name: "failed payment - cannot expire after failure",
			payment: Payment{
				Status:    StatusFailed,
				ExpiresAt: past,
			},
			expected: false,
		},
		{
			name: "cancelled payment - cannot expire after cancellation",
			payment: Payment{
				Status:    StatusCancelled,
				ExpiresAt: past,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.IsExpired(tt.payment)
			if result != tt.expected {
				t.Errorf("IsExpired() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestService_CanRefund tests refund eligibility logic
func TestService_CanRefund(t *testing.T) {
	service := NewService()

	now := time.Now()
	recentCompletion := now.Add(-1 * time.Hour)
	oldCompletion := now.Add(-31 * 24 * time.Hour) // 31 days ago

	refundPolicy := 30 * 24 * time.Hour // 30 days

	tests := []struct {
		name    string
		payment Payment
		wantErr bool
	}{
		{
			name: "valid refund - within policy",
			payment: Payment{
				Status:      StatusCompleted,
				CompletedAt: &recentCompletion,
			},
			wantErr: false,
		},
		{
			name: "invalid - pending payment",
			payment: Payment{
				Status:      StatusPending,
				CompletedAt: &recentCompletion,
			},
			wantErr: true,
		},
		{
			name: "invalid - processing payment",
			payment: Payment{
				Status:      StatusProcessing,
				CompletedAt: &recentCompletion,
			},
			wantErr: true,
		},
		{
			name: "invalid - failed payment",
			payment: Payment{
				Status:      StatusFailed,
				CompletedAt: nil,
			},
			wantErr: true,
		},
		{
			name: "invalid - cancelled payment",
			payment: Payment{
				Status:      StatusCancelled,
				CompletedAt: nil,
			},
			wantErr: true,
		},
		{
			name: "invalid - no completion date",
			payment: Payment{
				Status:      StatusCompleted,
				CompletedAt: nil,
			},
			wantErr: true,
		},
		{
			name: "invalid - outside refund policy",
			payment: Payment{
				Status:      StatusCompleted,
				CompletedAt: &oldCompletion,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CanRefund(tt.payment, refundPolicy)
			if (err != nil) != tt.wantErr {
				t.Errorf("CanRefund() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestService_GenerateInvoiceID_Uniqueness tests invoice ID generation uniqueness
func TestService_GenerateInvoiceID_Uniqueness(t *testing.T) {
	service := NewService()

	memberID := "member-123"
	paymentType := PaymentTypeFine

	// Generate multiple invoice IDs with 1-second delay (time.Now().Unix() has second precision)
	ids := make(map[string]bool)
	iterations := 3
	for i := 0; i < iterations; i++ {
		id := service.GenerateInvoiceID(memberID, paymentType)

		if id == "" {
			t.Error("GenerateInvoiceID() returned empty string")
		}

		if ids[id] {
			t.Errorf("GenerateInvoiceID() generated duplicate ID: %s", id)
		}
		ids[id] = true

		// Check format
		if !strings.HasPrefix(id, string(paymentType)+"-") {
			t.Errorf("GenerateInvoiceID() = %s, expected to start with %s-", id, paymentType)
		}

		// 1 second delay to ensure different Unix timestamps
		if i < iterations-1 {
			time.Sleep(1100 * time.Millisecond)
		}
	}

	if len(ids) != iterations {
		t.Errorf("Expected %d unique IDs, got %d", iterations, len(ids))
	}
}

// TestService_FormatAmount tests amount formatting for different currencies
func TestService_FormatAmount(t *testing.T) {
	service := NewService()

	tests := []struct {
		name     string
		amount   int64
		currency string
		expected string
	}{
		{"KZT formatting", 10000, "KZT", "100.00 KZT"},
		{"KZT small amount", 50, "KZT", "0.50 KZT"},
		{"KZT zero", 0, "KZT", "0.00 KZT"},
		{"USD formatting", 25000, "USD", "250.00 USD"},
		{"EUR formatting", 15000, "EUR", "150.00 EUR"},
		{"RUB formatting", 50000, "RUB", "500.00 RUB"},
		{"Unknown currency", 10000, "TTT", "10000 TTT"},
		{"Large amount KZT", 1000000, "KZT", "10000.00 KZT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.FormatAmount(tt.amount, tt.currency)
			if result != tt.expected {
				t.Errorf("FormatAmount(%d, %s) = %s, expected %s",
					tt.amount, tt.currency, result, tt.expected)
			}
		})
	}
}

// TestService_CalculateAmount tests amount calculation (placeholder implementation)
func TestService_CalculateAmount(t *testing.T) {
	service := NewService()

	tests := []struct {
		name            string
		paymentType     PaymentType
		relatedEntityID string
		wantErr         bool
	}{
		{"fine calculation - not implemented", PaymentTypeFine, "fine-123", true},
		{"subscription calculation - not implemented", PaymentTypeSubscription, "sub-456", true},
		{"deposit calculation - not implemented", PaymentTypeDeposit, "deposit-789", true},
		{"invalid payment type", "invalid", "entity-123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CalculateAmount(tt.paymentType, tt.relatedEntityID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateAmount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
