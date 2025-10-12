package domain

import (
	"testing"
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
