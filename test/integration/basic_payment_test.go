//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"library-service/internal/adapters/repository/postgres"
	"library-service/internal/domain/payment"
	"library-service/internal/usecase/paymentops"
)

// TestBasicPaymentOperations tests basic payment CRUD operations
func TestBasicPaymentOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	paymentRepo := postgres.NewPaymentRepository(db)

	ctx := context.Background()
	memberID := uuid.New().String()

	// Create a payment directly
	testPayment := payment.Payment{
		ID:            uuid.New().String(),
		MemberID:      memberID,
		InvoiceID:     "test-invoice-" + uuid.New().String(),
		Amount:        5000,
		Currency:      "KZT",
		PaymentType:   payment.PaymentTypeFine,
		Status:        payment.StatusPending,
		PaymentMethod: payment.PaymentMethodCard,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(30 * time.Minute),
	}

	var paymentID string

	t.Run("Create Payment", func(t *testing.T) {
		var err error
		paymentID, err = paymentRepo.Create(ctx, testPayment)
		require.NoError(t, err)
		testPayment.ID = paymentID
	})

	t.Run("Get Payment By ID", func(t *testing.T) {
		retrieved, err := paymentRepo.GetByID(ctx, testPayment.ID)
		require.NoError(t, err)
		assert.Equal(t, testPayment.ID, retrieved.ID)
		assert.Equal(t, testPayment.MemberID, retrieved.MemberID)
		assert.Equal(t, testPayment.Amount, retrieved.Amount)
		assert.Equal(t, testPayment.Currency, retrieved.Currency)
		assert.Equal(t, payment.StatusPending, retrieved.Status)
	})

	t.Run("Get Payment By Invoice ID", func(t *testing.T) {
		retrieved, err := paymentRepo.GetByInvoiceID(ctx, testPayment.InvoiceID)
		require.NoError(t, err)
		assert.Equal(t, testPayment.ID, retrieved.ID)
		assert.Equal(t, testPayment.InvoiceID, retrieved.InvoiceID)
	})

	t.Run("List Payments By Member", func(t *testing.T) {
		payments, err := paymentRepo.ListByMemberID(ctx, memberID)
		require.NoError(t, err)
		assert.Len(t, payments, 1)
		assert.Equal(t, testPayment.ID, payments[0].ID)
	})

	t.Run("Update Payment Status", func(t *testing.T) {
		err := paymentRepo.UpdateStatus(ctx, testPayment.ID, payment.StatusCompleted)
		require.NoError(t, err)

		// Verify status updated
		retrieved, err := paymentRepo.GetByID(ctx, testPayment.ID)
		require.NoError(t, err)
		assert.Equal(t, payment.StatusCompleted, retrieved.Status)
	})
}

// TestPaymentStatusTransitions tests payment status validation
func TestPaymentStatusTransitions(t *testing.T) {
	service := payment.NewService()

	tests := []struct {
		name          string
		currentStatus payment.Status
		newStatus     payment.Status
		wantErr       bool
	}{
		{"pending to processing", payment.StatusPending, payment.StatusProcessing, false},
		{"pending to completed", payment.StatusPending, payment.StatusCompleted, false},
		{"pending to cancelled", payment.StatusPending, payment.StatusCancelled, false},
		{"processing to completed", payment.StatusProcessing, payment.StatusCompleted, false},
		{"processing to failed", payment.StatusProcessing, payment.StatusFailed, false},
		{"completed to refunded", payment.StatusCompleted, payment.StatusRefunded, false},
		{"completed to pending - invalid", payment.StatusCompleted, payment.StatusPending, true},
		{"cancelled to completed - invalid", payment.StatusCancelled, payment.StatusCompleted, true},
		{"refunded to completed - invalid", payment.StatusRefunded, payment.StatusCompleted, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateStatusTransition(tt.currentStatus, tt.newStatus)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCallbackRetryBasicOperations tests callback retry CRUD
func TestCallbackRetryBasicOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	callbackRetryRepo := postgres.NewCallbackRetryRepository(db)
	paymentRepo := postgres.NewPaymentRepository(db)

	ctx := context.Background()

	// Create a payment first
	testPayment := payment.Payment{
		ID:            uuid.New().String(),
		MemberID:      uuid.New().String(),
		InvoiceID:     "callback-test-" + uuid.New().String(),
		Amount:        3000,
		Currency:      "KZT",
		PaymentType:   payment.PaymentTypeFine,
		Status:        payment.StatusPending,
		PaymentMethod: payment.PaymentMethodCard,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(30 * time.Minute),
	}

	paymentID, err := paymentRepo.Create(ctx, testPayment)
	require.NoError(t, err)
	testPayment.ID = paymentID

	// Create a callback retry
	now := time.Now()
	callbackRetry := &payment.CallbackRetry{
		ID:           uuid.New().String(),
		PaymentID:    testPayment.ID,
		CallbackData: []byte(`{"invoice_id":"test","amount":3000}`),
		RetryCount:   0,
		MaxRetries:   5,
		NextRetryAt:  &now,
		Status:       payment.CallbackRetryStatusPending,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	t.Run("Create Callback Retry", func(t *testing.T) {
		err := callbackRetryRepo.Create(callbackRetry)
		require.NoError(t, err)
	})

	t.Run("Get Callback Retry By ID", func(t *testing.T) {
		retrieved, err := callbackRetryRepo.GetByID(callbackRetry.ID)
		require.NoError(t, err)
		assert.Equal(t, callbackRetry.ID, retrieved.ID)
		assert.Equal(t, callbackRetry.PaymentID, retrieved.PaymentID)
		assert.Equal(t, payment.CallbackRetryStatusPending, retrieved.Status)
	})

	t.Run("Get Pending Retries", func(t *testing.T) {
		retries, err := callbackRetryRepo.GetPendingRetries(10)
		require.NoError(t, err)
		assert.Len(t, retries, 1)
		assert.Equal(t, callbackRetry.ID, retries[0].ID)
	})

	t.Run("Update Callback Retry", func(t *testing.T) {
		callbackRetry.RetryCount = 1
		callbackRetry.Status = payment.CallbackRetryStatusCompleted
		callbackRetry.UpdatedAt = time.Now()

		err := callbackRetryRepo.Update(callbackRetry)
		require.NoError(t, err)

		// Verify update
		retrieved, err := callbackRetryRepo.GetByID(callbackRetry.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, retrieved.RetryCount)
		assert.Equal(t, payment.CallbackRetryStatusCompleted, retrieved.Status)
	})
}

// TestSavedCardBasicOperations tests saved card CRUD
func TestSavedCardBasicOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	savedCardRepo := postgres.NewSavedCardRepository(db)
	ctx := context.Background()
	memberID := uuid.New().String()

	// Create a saved card
	savedCard := payment.SavedCard{
		ID:          uuid.New().String(),
		MemberID:    memberID,
		CardToken:   "card-token-" + uuid.New().String(),
		CardMask:    "4111 11** **** 1111",
		CardType:    "visa",
		ExpiryMonth: 12,
		ExpiryYear:  2025,
		IsDefault:   true,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	var cardID string

	t.Run("Create Saved Card", func(t *testing.T) {
		var err error
		cardID, err = savedCardRepo.Create(ctx, savedCard)
		require.NoError(t, err)
		savedCard.ID = cardID
	})

	t.Run("Get Saved Card By ID", func(t *testing.T) {
		retrieved, err := savedCardRepo.GetByID(ctx, savedCard.ID)
		require.NoError(t, err)
		assert.Equal(t, savedCard.ID, retrieved.ID)
		assert.Equal(t, savedCard.MemberID, retrieved.MemberID)
		assert.Equal(t, savedCard.CardMask, retrieved.CardMask)
		assert.True(t, retrieved.IsDefault)
	})

	t.Run("List Saved Cards By Member", func(t *testing.T) {
		cards, err := savedCardRepo.ListByMemberID(ctx, memberID)
		require.NoError(t, err)
		assert.Len(t, cards, 1)
		assert.Equal(t, savedCard.ID, cards[0].ID)
	})

	t.Run("Delete Saved Card", func(t *testing.T) {
		err := savedCardRepo.Delete(ctx, savedCard.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = savedCardRepo.GetByID(ctx, savedCard.ID)
		assert.Error(t, err)
	})
}

// TestExpirePaymentsUseCase tests the payment expiry use case
func TestExpirePaymentsUseCase(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	paymentRepo := postgres.NewPaymentRepository(db)
	paymentService := payment.NewService()
	expireUC := paymentops.NewExpirePaymentsUseCase(paymentRepo, paymentService)

	ctx := context.Background()

	// Create an expired pending payment
	expiredPayment := payment.Payment{
		ID:            uuid.New().String(),
		MemberID:      uuid.New().String(),
		InvoiceID:     "expired-" + uuid.New().String(),
		Amount:        2000,
		Currency:      "KZT",
		PaymentType:   payment.PaymentTypeFine,
		Status:        payment.StatusPending,
		PaymentMethod: payment.PaymentMethodCard,
		CreatedAt:     time.Now().Add(-2 * time.Hour),
		UpdatedAt:     time.Now().Add(-2 * time.Hour),
		ExpiresAt:     time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}

	expiredPaymentID, err := paymentRepo.Create(ctx, expiredPayment)
	require.NoError(t, err)
	expiredPayment.ID = expiredPaymentID

	t.Run("Expire Old Payments", func(t *testing.T) {
		req := paymentops.ExpirePaymentsRequest{
			BatchSize: 100,
		}

		resp, err := expireUC.Execute(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, 1, resp.ExpiredCount)
		assert.Equal(t, 0, resp.FailedCount)

		// Verify payment status updated to failed
		updatedPayment, err := paymentRepo.GetByID(ctx, expiredPayment.ID)
		require.NoError(t, err)
		assert.Equal(t, payment.StatusFailed, updatedPayment.Status)
	})
}
