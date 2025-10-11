//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"library-service/internal/adapters/repository/postgres"
	"library-service/internal/domain/reservation"
	"library-service/test/fixtures"
)

func TestReservationRepository_Integration(t *testing.T) {
	db := Setup(t)
	defer db.Cleanup()

	// Clean up test data
	db.TruncateAll()

	repo := postgres.NewReservationRepository(db.DB)
	ctx := context.Background()

	t.Run("Complete CRUD workflow", func(t *testing.T) {
		db.Truncate("reservations")

		// CREATE
		testReservation := fixtures.ReservationForCreate("book-001", "member-001")
		id, err := repo.Create(ctx, testReservation)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		if id == "" {
			t.Fatal("Create returned empty ID")
		}

		// Verify created
		db.AssertExists("reservations", id)
		db.AssertRowCount("reservations", 1)

		// READ
		retrieved, err := repo.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if retrieved.ID != id {
			t.Errorf("ID mismatch: expected %s, got %s", id, retrieved.ID)
		}

		if retrieved.BookID != testReservation.BookID {
			t.Errorf("BookID mismatch: expected %s, got %s", testReservation.BookID, retrieved.BookID)
		}

		if retrieved.MemberID != testReservation.MemberID {
			t.Errorf("MemberID mismatch: expected %s, got %s", testReservation.MemberID, retrieved.MemberID)
		}

		// UPDATE
		retrieved.Status = reservation.StatusFulfilled
		fulfilledTime := time.Now()
		retrieved.FulfilledAt = &fulfilledTime

		err = repo.Update(ctx, retrieved)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// Verify update
		updated, err := repo.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("GetByID after Update failed: %v", err)
		}

		if updated.Status != reservation.StatusFulfilled {
			t.Errorf("Status not updated: expected %s, got %s", reservation.StatusFulfilled, updated.Status)
		}

		if updated.FulfilledAt == nil {
			t.Error("FulfilledAt should not be nil after update")
		}

		// DELETE
		err = repo.Delete(ctx, id)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// Verify deletion
		db.AssertNotExists("reservations", id)
		db.AssertRowCount("reservations", 0)

		// Try to get deleted reservation
		_, err = repo.GetByID(ctx, id)
		if err == nil {
			t.Error("GetByID should fail for deleted reservation")
		}
	})

	t.Run("GetByMemberID", func(t *testing.T) {
		db.Truncate("reservations")

		memberID := "member-001"

		// Create reservations for the member
		reservation1 := fixtures.ReservationForCreate("book-001", memberID)
		reservation2 := fixtures.ReservationForCreate("book-002", memberID)
		reservation3 := fixtures.ReservationForCreate("book-003", "member-002")

		_, err := repo.Create(ctx, reservation1)
		if err != nil {
			t.Fatalf("Create reservation1 failed: %v", err)
		}

		_, err = repo.Create(ctx, reservation2)
		if err != nil {
			t.Fatalf("Create reservation2 failed: %v", err)
		}

		_, err = repo.Create(ctx, reservation3)
		if err != nil {
			t.Fatalf("Create reservation3 failed: %v", err)
		}

		// Get by member ID
		reservations, err := repo.GetByMemberID(ctx, memberID)
		if err != nil {
			t.Fatalf("GetByMemberID failed: %v", err)
		}

		if len(reservations) != 2 {
			t.Errorf("expected 2 reservations for member %s, got %d", memberID, len(reservations))
		}

		// Verify all reservations belong to the member
		for _, r := range reservations {
			if r.MemberID != memberID {
				t.Errorf("reservation %s has wrong member ID: expected %s, got %s", r.ID, memberID, r.MemberID)
			}
		}
	})

	t.Run("GetByBookID", func(t *testing.T) {
		db.Truncate("reservations")

		bookID := "book-001"

		// Create reservations for the book
		reservation1 := fixtures.ReservationForCreate(bookID, "member-001")
		reservation2 := fixtures.ReservationForCreate(bookID, "member-002")
		reservation3 := fixtures.ReservationForCreate("book-002", "member-003")

		_, err := repo.Create(ctx, reservation1)
		if err != nil {
			t.Fatalf("Create reservation1 failed: %v", err)
		}

		_, err = repo.Create(ctx, reservation2)
		if err != nil {
			t.Fatalf("Create reservation2 failed: %v", err)
		}

		_, err = repo.Create(ctx, reservation3)
		if err != nil {
			t.Fatalf("Create reservation3 failed: %v", err)
		}

		// Get by book ID
		reservations, err := repo.GetByBookID(ctx, bookID)
		if err != nil {
			t.Fatalf("GetByBookID failed: %v", err)
		}

		if len(reservations) != 2 {
			t.Errorf("expected 2 reservations for book %s, got %d", bookID, len(reservations))
		}

		// Verify all reservations are for the book
		for _, r := range reservations {
			if r.BookID != bookID {
				t.Errorf("reservation %s has wrong book ID: expected %s, got %s", r.ID, bookID, r.BookID)
			}
		}
	})

	t.Run("GetActiveByMemberAndBook", func(t *testing.T) {
		db.Truncate("reservations")

		memberID := "member-001"
		bookID := "book-001"

		// Create active reservation
		activeReservation := fixtures.ReservationForCreate(bookID, memberID)
		activeReservation.Status = reservation.StatusPending

		_, err := repo.Create(ctx, activeReservation)
		if err != nil {
			t.Fatalf("Create active reservation failed: %v", err)
		}

		// Create cancelled reservation (not active)
		cancelledReservation := fixtures.ReservationForCreate(bookID, memberID)
		cancelledReservation.Status = reservation.StatusCancelled
		cancelledTime := time.Now()
		cancelledReservation.CancelledAt = &cancelledTime

		_, err = repo.Create(ctx, cancelledReservation)
		if err != nil {
			t.Fatalf("Create cancelled reservation failed: %v", err)
		}

		// Get active reservations
		activeReservations, err := repo.GetActiveByMemberAndBook(ctx, memberID, bookID)
		if err != nil {
			t.Fatalf("GetActiveByMemberAndBook failed: %v", err)
		}

		// Should only get the active (pending) reservation, not the cancelled one
		if len(activeReservations) != 1 {
			t.Errorf("expected 1 active reservation, got %d", len(activeReservations))
		}

		if len(activeReservations) > 0 && activeReservations[0].Status != reservation.StatusPending {
			t.Errorf("expected pending status, got %s", activeReservations[0].Status)
		}
	})

	t.Run("ListPending", func(t *testing.T) {
		db.Truncate("reservations")

		// Create reservations with different statuses
		pendingReservation := fixtures.PendingReservation()
		fulfilledReservation := fixtures.FulfilledReservation()
		cancelledReservation := fixtures.CancelledReservation()

		_, err := repo.Create(ctx, pendingReservation)
		if err != nil {
			t.Fatalf("Create pending reservation failed: %v", err)
		}

		_, err = repo.Create(ctx, fulfilledReservation)
		if err != nil {
			t.Fatalf("Create fulfilled reservation failed: %v", err)
		}

		_, err = repo.Create(ctx, cancelledReservation)
		if err != nil {
			t.Fatalf("Create cancelled reservation failed: %v", err)
		}

		// List pending
		pendingReservations, err := repo.ListPending(ctx)
		if err != nil {
			t.Fatalf("ListPending failed: %v", err)
		}

		if len(pendingReservations) != 1 {
			t.Errorf("expected 1 pending reservation, got %d", len(pendingReservations))
		}

		if len(pendingReservations) > 0 && pendingReservations[0].Status != reservation.StatusPending {
			t.Errorf("expected pending status, got %s", pendingReservations[0].Status)
		}
	})

	t.Run("ListExpired", func(t *testing.T) {
		db.Truncate("reservations")

		// Create expired reservation
		expiredReservation := fixtures.ExpiredReservation()

		// Create non-expired reservation
		pendingReservation := fixtures.PendingReservation()

		_, err := repo.Create(ctx, expiredReservation)
		if err != nil {
			t.Fatalf("Create expired reservation failed: %v", err)
		}

		_, err = repo.Create(ctx, pendingReservation)
		if err != nil {
			t.Fatalf("Create pending reservation failed: %v", err)
		}

		// List expired
		expiredReservations, err := repo.ListExpired(ctx)
		if err != nil {
			t.Fatalf("ListExpired failed: %v", err)
		}

		if len(expiredReservations) != 1 {
			t.Errorf("expected 1 expired reservation, got %d", len(expiredReservations))
		}

		if len(expiredReservations) > 0 && expiredReservations[0].Status != reservation.StatusExpired {
			t.Errorf("expected expired status, got %s", expiredReservations[0].Status)
		}
	})

	t.Run("Batch operations", func(t *testing.T) {
		db.Truncate("reservations")

		// Add multiple reservations
		sampleReservations := fixtures.Reservations()
		reservationIDs := make([]string, 0, len(sampleReservations))

		for i, r := range sampleReservations {
			id, err := repo.Create(ctx, r)
			if err != nil {
				t.Fatalf("failed to create reservation %d: %v", i, err)
			}
			reservationIDs = append(reservationIDs, id)
		}

		// Verify count
		db.AssertRowCount("reservations", len(sampleReservations))

		// Delete all
		for _, id := range reservationIDs {
			if err := repo.Delete(ctx, id); err != nil {
				t.Errorf("failed to delete reservation %s: %v", id, err)
			}
		}

		// Verify all deleted
		db.AssertRowCount("reservations", 0)
	})
}
