package domain

import (
	"testing"
	"time"
)

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
