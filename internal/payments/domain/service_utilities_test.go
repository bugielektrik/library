package domain

import (
	"strings"
	"testing"
	"time"
)

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
