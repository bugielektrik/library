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

	"library-service/internal/infrastructure/pkg/repository/postgres"
	"library-service/internal/payments/domain"
	"library-service/internal/payments/service/payment"
)

// TestPaymentSimpleFlow tests a simple payment flow
func TestPaymentSimpleFlow(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	paymentRepo := postgres.NewPaymentRepository(db)
	paymentService := domain.NewService()
	mockGateway := &MockPaymentGateway{
		terminal:  "test-terminal",
		backLink:  "http://localhost:8080/payment",
		postLink:  "http://localhost:8080/api/v1/payments/callback",
		widgetURL: "https://test.edomain.kz/widget",
	}

	initiateUC := paymentops.NewInitiatePaymentUseCase(paymentRepo, paymentService, mockGateway)
	handleCallbackUC := paymentops.NewHandleCallbackUseCase(paymentRepo, paymentService)

	ctx := context.Background()
	memberID := uuid.New().String()

	t.Run("Complete Payment Flow", func(t *testing.T) {
		// Step 1: Initiate payment
		req := paymentops.InitiatePaymentRequest{
			MemberID:    memberID,
			Amount:      5000,
			Currency:    "KZT",
			PaymentType: domain.PaymentTypeFine,
		}

		resp, err := initiateUC.Execute(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.PaymentID)
		assert.NotEmpty(t, resp.InvoiceID)
		assert.Equal(t, int64(5000), resp.Amount)

		// Step 2: Get payment from DB
		paymentEntity, err := paymentRepo.GetByID(ctx, resp.PaymentID)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusPending, paymentEntity.Status)

		// Step 3: Simulate successful callback
		callbackReq := paymentops.PaymentCallbackRequest{
			InvoiceID:     resp.InvoiceID,
			TransactionID: "txn-" + uuid.New().String(),
			Amount:        5000,
			Currency:      "KZT",
			Status:        "success",
		}

		callbackResp, err := handleCallbackUC.Execute(ctx, callbackReq)
		require.NoError(t, err)
		assert.Equal(t, resp.PaymentID, callbackResp.PaymentID)
		assert.Equal(t, domain.StatusCompleted, callbackResp.Status)

		// Step 4: Verify payment status updated
		updatedPayment, err := paymentRepo.GetByID(ctx, resp.PaymentID)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusCompleted, updatedPayment.Status)
		assert.NotNil(t, updatedPayment.CompletedAt)
	})
}

// TestPaymentIdempotency tests that duplicate callbacks are handled correctly
func TestPaymentIdempotency(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	paymentRepo := postgres.NewPaymentRepository(db)
	paymentService := domain.NewService()
	handleCallbackUC := paymentops.NewHandleCallbackUseCase(paymentRepo, paymentService)

	ctx := context.Background()

	// Create a completed payment
	testPayment := domain.Payment{
		ID:            uuid.New().String(),
		MemberID:      uuid.New().String(),
		InvoiceID:     "idempotency-test-" + uuid.New().String(),
		Amount:        3000,
		Currency:      "KZT",
		PaymentType:   domain.PaymentTypeFine,
		Status:        domain.StatusCompleted,
		PaymentMethod: domain.PaymentMethodCard,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(30 * time.Minute),
	}

	paymentID, err := paymentRepo.Create(ctx, testPayment)
	require.NoError(t, err)
	testPayment.ID = paymentID

	t.Run("Duplicate Callback Ignored", func(t *testing.T) {
		// Send duplicate callback for already completed payment
		callbackReq := paymentops.PaymentCallbackRequest{
			InvoiceID:     testPayment.InvoiceID,
			TransactionID: "txn-" + uuid.New().String(),
			Amount:        3000,
			Currency:      "KZT",
			Status:        "success",
		}

		resp, err := handleCallbackUC.Execute(ctx, callbackReq)
		require.NoError(t, err)
		assert.Equal(t, testPayment.ID, resp.PaymentID)
		assert.Equal(t, domain.StatusCompleted, resp.Status)
		assert.False(t, resp.Processed) // Should indicate no processing was needed
	})
}

// TestPaymentExpiry tests payment expiration
func TestPaymentExpiry(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	paymentRepo := postgres.NewPaymentRepository(db)
	paymentService := domain.NewService()
	expireUC := paymentops.NewExpirePaymentsUseCase(paymentRepo, paymentService)

	ctx := context.Background()

	// Create expired payment
	expiredPayment := domain.Payment{
		ID:            uuid.New().String(),
		MemberID:      uuid.New().String(),
		InvoiceID:     "expired-" + uuid.New().String(),
		Amount:        2000,
		Currency:      "KZT",
		PaymentType:   domain.PaymentTypeFine,
		Status:        domain.StatusPending,
		PaymentMethod: domain.PaymentMethodCard,
		CreatedAt:     time.Now().Add(-2 * time.Hour),
		UpdatedAt:     time.Now().Add(-2 * time.Hour),
		ExpiresAt:     time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}

	paymentID, err := paymentRepo.Create(ctx, expiredPayment)
	require.NoError(t, err)
	expiredPayment.ID = paymentID

	t.Run("Expire Pending Payment", func(t *testing.T) {
		req := paymentops.ExpirePaymentsRequest{
			BatchSize: 100,
		}

		resp, err := expireUC.Execute(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, 1, resp.ExpiredCount)
		assert.Equal(t, 0, resp.FailedCount)

		// Verify payment is now failed
		updatedPayment, err := paymentRepo.GetByID(ctx, expiredPayment.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusFailed, updatedPayment.Status)
	})
}

// TestRefundFlow tests the refund flow
func TestRefundFlow(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	paymentRepo := postgres.NewPaymentRepository(db)
	paymentService := domain.NewService()
	mockGateway := &MockPaymentGateway{
		terminal:  "test-terminal",
		backLink:  "http://localhost:8080/payment",
		postLink:  "http://localhost:8080/api/v1/payments/callback",
		widgetURL: "https://test.edomain.kz/widget",
	}
	refundUC := paymentops.NewRefundPaymentUseCase(paymentRepo, paymentService, mockGateway)

	ctx := context.Background()
	memberID := uuid.New().String()

	// Create a completed payment
	testPayment := domain.Payment{
		ID:            uuid.New().String(),
		MemberID:      memberID,
		InvoiceID:     "refund-test-" + uuid.New().String(),
		Amount:        10000,
		Currency:      "KZT",
		PaymentType:   domain.PaymentTypeFine,
		Status:        domain.StatusCompleted,
		PaymentMethod: domain.PaymentMethodCard,
		CreatedAt:     time.Now().Add(-1 * time.Hour),
		UpdatedAt:     time.Now().Add(-1 * time.Hour),
		ExpiresAt:     time.Now().Add(30 * time.Minute),
	}

	paymentID, err := paymentRepo.Create(ctx, testPayment)
	require.NoError(t, err)
	testPayment.ID = paymentID

	t.Run("Full Refund", func(t *testing.T) {
		refundAmount := int64(10000)
		req := paymentops.RefundPaymentRequest{
			PaymentID:    testPayment.ID,
			MemberID:     memberID,
			Reason:       "Test refund",
			IsAdmin:      true,
			RefundAmount: &refundAmount,
		}

		resp, err := refundUC.Execute(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, testPayment.ID, resp.PaymentID)
		assert.Equal(t, int64(10000), resp.Amount)

		// Verify payment status updated
		updatedPayment, err := paymentRepo.GetByID(ctx, testPayment.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusRefunded, updatedPayment.Status)
	})
}

// TestReceiptGeneration tests receipt generation
func TestReceiptGeneration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	paymentRepo := postgres.NewPaymentRepository(db)
	receiptRepo := postgres.NewReceiptRepository(db)
	memberRepo := postgres.NewMemberRepository(db)

	ctx := context.Background()
	memberID := uuid.New().String()

	// Create a completed payment
	testPayment := domain.Payment{
		ID:            uuid.New().String(),
		MemberID:      memberID,
		InvoiceID:     "receipt-test-" + uuid.New().String(),
		Amount:        5000,
		Currency:      "KZT",
		PaymentType:   domain.PaymentTypeFine,
		Status:        domain.StatusCompleted,
		PaymentMethod: domain.PaymentMethodCard,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(30 * time.Minute),
	}

	paymentID, err := paymentRepo.Create(ctx, testPayment)
	require.NoError(t, err)
	testPayment.ID = paymentID

	generateUC := paymentops.NewGenerateReceiptUseCase(paymentRepo, receiptRepo, memberRepo)

	t.Run("Generate Receipt", func(t *testing.T) {
		req := paymentops.GenerateReceiptRequest{
			PaymentID: testPayment.ID,
			MemberID:  memberID,
		}

		resp, err := generateUC.Execute(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.ReceiptID)
		assert.NotEmpty(t, resp.ReceiptNumber)
		assert.Contains(t, resp.ReceiptNumber, "RCP-")

		// Verify receipt in database
		receipt, err := receiptRepo.GetByID(resp.ReceiptID)
		require.NoError(t, err)
		assert.Equal(t, testPayment.ID, receipt.PaymentID)
		assert.Equal(t, testPayment.Amount, receipt.Amount)
	})

	t.Run("Receipt Idempotency", func(t *testing.T) {
		// Try to generate receipt again
		req := paymentops.GenerateReceiptRequest{
			PaymentID: testPayment.ID,
			MemberID:  memberID,
		}

		resp, err := generateUC.Execute(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.ReceiptID)

		// Should return existing receipt
		receipts, err := receiptRepo.ListByMemberID(memberID)
		require.NoError(t, err)
		assert.Len(t, receipts, 1) // Only one receipt should exist
	})
}
