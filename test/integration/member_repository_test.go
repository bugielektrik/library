//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"library-service/internal/infrastructure/pkg/repository/postgres"
	"library-service/test/fixtures"
)

func TestMemberRepository_Integration(t *testing.T) {
	db := Setup(t)
	defer db.Cleanup()

	// Clean up test data
	db.TruncateAll()

	repo := postgres.NewMemberRepository(db.DB)
	ctx := context.Background()

	t.Run("Complete CRUD workflow", func(t *testing.T) {
		db.Truncate("members")

		// CREATE
		testMember := fixtures.MemberForCreate()
		id, err := repo.Add(ctx, testMember)
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}

		if id == "" {
			t.Fatal("Add returned empty ID")
		}

		// Verify created
		db.AssertExists("members", id)
		db.AssertRowCount("members", 1)

		// READ
		retrieved, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if retrieved.ID != id {
			t.Errorf("ID mismatch: expected %s, got %s", id, retrieved.ID)
		}

		if retrieved.Email != testMember.Email {
			t.Errorf("Email mismatch: expected %s, got %s", testMember.Email, retrieved.Email)
		}

		// LIST
		members, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(members) != 1 {
			t.Errorf("List returned wrong count: expected 1, got %d", len(members))
		}

		// UPDATE
		updateData := fixtures.MemberUpdate()
		err = repo.Update(ctx, id, updateData)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// Verify update
		retrieved, err = repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("Get after Update failed: %v", err)
		}

		if retrieved.FullName == nil || *retrieved.FullName != *updateData.FullName {
			t.Errorf("Update didn't apply: expected %v, got %v", updateData.FullName, retrieved.FullName)
		}

		// DELETE
		err = repo.Delete(ctx, id)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// Verify deletion
		db.AssertNotExists("members", id)
		db.AssertRowCount("members", 0)

		// Try to get deleted member
		_, err = repo.Get(ctx, id)
		if err == nil {
			t.Error("Get should fail for deleted member")
		}
	})

	t.Run("GetByEmail", func(t *testing.T) {
		db.Truncate("members")

		// Add member
		testMember := fixtures.MemberForCreate()
		id, err := repo.Add(ctx, testMember)
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}

		// Get by email
		retrieved, err := repo.GetByEmail(ctx, testMember.Email)
		if err != nil {
			t.Fatalf("GetByEmail failed: %v", err)
		}

		if retrieved.ID != id {
			t.Errorf("GetByEmail returned wrong member: expected %s, got %s", id, retrieved.ID)
		}

		if retrieved.Email != testMember.Email {
			t.Errorf("Email mismatch: expected %s, got %s", testMember.Email, retrieved.Email)
		}

		// Try non-existent email
		_, err = repo.GetByEmail(ctx, "nonexistent@example.com")
		if err == nil {
			t.Error("GetByEmail should fail for non-existent email")
		}
	})

	t.Run("EmailExists", func(t *testing.T) {
		db.Truncate("members")

		// Add member
		testMember := fixtures.MemberForCreate()
		_, err := repo.Add(ctx, testMember)
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}

		// Check existing email
		exists, err := repo.EmailExists(ctx, testMember.Email)
		if err != nil {
			t.Fatalf("EmailExists failed: %v", err)
		}

		if !exists {
			t.Error("EmailExists should return true for existing email")
		}

		// Check non-existent email
		exists, err = repo.EmailExists(ctx, "nonexistent@example.com")
		if err != nil {
			t.Fatalf("EmailExists failed for non-existent email: %v", err)
		}

		if exists {
			t.Error("EmailExists should return false for non-existent email")
		}
	})

	t.Run("UpdateLastLogin", func(t *testing.T) {
		db.Truncate("members")

		// Add member
		testMember := fixtures.MemberForCreate()
		id, err := repo.Add(ctx, testMember)
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}

		// Update last login
		loginTime := time.Now()
		err = repo.UpdateLastLogin(ctx, id, loginTime)
		if err != nil {
			t.Fatalf("UpdateLastLogin failed: %v", err)
		}

		// Verify update
		retrieved, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if retrieved.LastLoginAt == nil {
			t.Error("LastLoginAt should not be nil after update")
		} else {
			// Allow small time difference due to database precision
			if retrieved.LastLoginAt.Sub(loginTime).Abs() > time.Second {
				t.Errorf("LastLoginAt mismatch: expected %v, got %v", loginTime, *retrieved.LastLoginAt)
			}
		}
	})

	t.Run("Batch operations", func(t *testing.T) {
		db.Truncate("members")

		// Add multiple members
		sampleMembers := fixtures.Members()
		memberIDs := make([]string, 0, len(sampleMembers))

		for i, member := range sampleMembers {
			id, err := repo.Add(ctx, member)
			if err != nil {
				t.Fatalf("failed to add member %d: %v", i, err)
			}
			memberIDs = append(memberIDs, id)
		}

		// Verify count
		db.AssertRowCount("members", len(sampleMembers))

		// List all
		members, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(members) != len(sampleMembers) {
			t.Errorf("expected %d members, got %d", len(sampleMembers), len(members))
		}

		// Delete all
		for _, id := range memberIDs {
			if err := repo.Delete(ctx, id); err != nil {
				t.Errorf("failed to delete member %s: %v", id, err)
			}
		}

		// Verify all deleted
		db.AssertRowCount("members", 0)
	})
}
