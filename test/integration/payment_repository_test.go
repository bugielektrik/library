//go:build integration

package integration

import (
	"context"
	"testing"

	"library-service/internal/infrastructure/pkg/repository/postgres"
	"library-service/internal/payments/domain"
	"library-service/test/fixtures"
)

func TestPaymentRepository_Integration(t *testing.T) {
	db := Setup(t)
	defer db.Cleanup()

	// Clean up test data
	db.TruncateAll()

	repo := postgres.NewPaymentRepository(db.DB)
	ctx := context.Background()

	t.Run("Complete CRUD workflow", func(t *testing.T) {
		db.Truncate("payments")

		// CREATE
		testPayment := fixtures.PaymentForCreate()
		id, err := repo.Create(ctx, testPayment)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		if id == "" {
			t.Fatal("Create returned empty ID")
		}

		// Verify created
		db.AssertExists("payments", id)
		db.AssertRowCount("payments", 1)

		// READ by ID
		retrieved, err := repo.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if retrieved.ID != id {
			t.Errorf("ID mismatch: expected %s, got %s", id, retrieved.ID)
		}

		if retrieved.InvoiceID != testPayment.InvoiceID {
			t.Errorf("InvoiceID mismatch: expected %s, got %s", testPayment.InvoiceID, retrieved.InvoiceID)
		}

		// UPDATE
		retrieved.Amount = 15000
		err = repo.Update(ctx, id, retrieved)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// Verify update
		updated, err := repo.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("GetByID after Update failed: %v", err)
		}

		if updated.Amount != 15000 {
			t.Errorf("Amount not updated: expected 15000, got %d", updated.Amount)
		}
	})

	t.Run("GetByInvoiceID", func(t *testing.T) {
		db.Truncate("payments")

		// Create payment
		testPayment := fixtures.PaymentForCreate()
		id, err := repo.Create(ctx, testPayment)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		// Get by invoice ID
		retrieved, err := repo.GetByInvoiceID(ctx, testPayment.InvoiceID)
		if err != nil {
			t.Fatalf("GetByInvoiceID failed: %v", err)
		}

		if retrieved.ID != id {
			t.Errorf("GetByInvoiceID returned wrong payment: expected %s, got %s", id, retrieved.ID)
		}

		// Try non-existent invoice ID
		_, err = repo.GetByInvoiceID(ctx, "non-existent-invoice")
		if err == nil {
			t.Error("GetByInvoiceID should fail for non-existent invoice")
		}
	})

	t.Run("ListByMemberID", func(t *testing.T) {
		db.Truncate("payments")

		// Create payments for different members
		memberID := "member-001"
		payment1 := fixtures.PaymentForCreate()
		payment1.MemberID = memberID

		payment2 := fixtures.PaymentForCreate()
		payment2.MemberID = memberID
		payment2.InvoiceID = "invoice-002"

		payment3 := fixtures.PaymentForCreate()
		payment3.MemberID = "member-002"
		payment3.InvoiceID = "invoice-003"

		_, err := repo.Create(ctx, payment1)
		if err != nil {
			t.Fatalf("Create payment1 failed: %v", err)
		}

		_, err = repo.Create(ctx, payment2)
		if err != nil {
			t.Fatalf("Create payment2 failed: %v", err)
		}

		_, err = repo.Create(ctx, payment3)
		if err != nil {
			t.Fatalf("Create payment3 failed: %v", err)
		}

		// List by member ID
		payments, err := repo.ListByMemberID(ctx, memberID)
		if err != nil {
			t.Fatalf("ListByMemberID failed: %v", err)
		}

		if len(payments) != 2 {
			t.Errorf("expected 2 payments for member %s, got %d", memberID, len(payments))
		}

		// Verify all payments belong to the member
		for _, p := range payments {
			if p.MemberID != memberID {
				t.Errorf("payment %s has wrong member ID: expected %s, got %s", p.ID, memberID, p.MemberID)
			}
		}
	})

	t.Run("ListByStatus", func(t *testing.T) {
		db.Truncate("payments")

		// Create payments with different statuses
		completedPayment := fixtures.CompletedPayment()
		pendingPayment := fixtures.PendingPayment()
		failedPayment := fixtures.FailedPayment()

		_, err := repo.Create(ctx, completedPayment)
		if err != nil {
			t.Fatalf("Create completed payment failed: %v", err)
		}

		_, err = repo.Create(ctx, pendingPayment)
		if err != nil {
			t.Fatalf("Create pending payment failed: %v", err)
		}

		_, err = repo.Create(ctx, failedPayment)
		if err != nil {
			t.Fatalf("Create failed payment failed: %v", err)
		}

		// List by status
		completedPayments, err := repo.ListByStatus(ctx, domain.StatusCompleted)
		if err != nil {
			t.Fatalf("ListByStatus(completed) failed: %v", err)
		}

		if len(completedPayments) != 1 {
			t.Errorf("expected 1 completed payment, got %d", len(completedPayments))
		}

		pendingPayments, err := repo.ListByStatus(ctx, domain.StatusPending)
		if err != nil {
			t.Fatalf("ListByStatus(pending) failed: %v", err)
		}

		if len(pendingPayments) != 1 {
			t.Errorf("expected 1 pending payment, got %d", len(pendingPayments))
		}
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		db.Truncate("payments")

		// Create payment
		testPayment := fixtures.PendingPayment()
		id, err := repo.Create(ctx, testPayment)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		// Update status
		err = repo.UpdateStatus(ctx, id, domain.StatusCompleted)
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		// Verify status update
		retrieved, err := repo.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if retrieved.Status != domain.StatusCompleted {
			t.Errorf("Status not updated: expected %s, got %s", domain.StatusCompleted, retrieved.Status)
		}
	})

	t.Run("Batch operations", func(t *testing.T) {
		db.Truncate("payments")

		// Add multiple payments
		samplePayments := fixtures.Payments()
		paymentIDs := make([]string, 0, len(samplePayments))

		for i, p := range samplePayments {
			id, err := repo.Create(ctx, p)
			if err != nil {
				t.Fatalf("failed to create payment %d: %v", i, err)
			}
			paymentIDs = append(paymentIDs, id)
		}

		// Verify count
		db.AssertRowCount("payments", len(samplePayments))

		// Verify each payment exists
		for _, id := range paymentIDs {
			db.AssertExists("payments", id)
		}
	})
}
