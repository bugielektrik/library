//go:build integration

package integration

import (
	"context"
	"testing"

	"library-service/internal/infrastructure/pkg/repository/postgres"
	"library-service/test/fixtures"
)

func TestAuthorRepository_Integration(t *testing.T) {
	db := Setup(t)
	defer db.Cleanup()

	// Clean up test data
	db.TruncateAll()

	repo := postgres.NewAuthorRepository(db.DB)
	ctx := context.Background()

	t.Run("Complete CRUD workflow", func(t *testing.T) {
		db.Truncate("authors")

		// CREATE
		testAuthor := fixtures.AuthorForCreate()
		id, err := repo.Add(ctx, testAuthor)
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}

		if id == "" {
			t.Fatal("Add returned empty ID")
		}

		// Verify created
		db.AssertExists("authors", id)
		db.AssertRowCount("authors", 1)

		// READ
		retrieved, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if retrieved.ID != id {
			t.Errorf("ID mismatch: expected %s, got %s", id, retrieved.ID)
		}

		if retrieved.FullName == nil || *retrieved.FullName != *testAuthor.FullName {
			t.Errorf("FullName mismatch: expected %v, got %v", testAuthor.FullName, retrieved.FullName)
		}

		// LIST
		authors, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(authors) != 1 {
			t.Errorf("List returned wrong count: expected 1, got %d", len(authors))
		}

		// UPDATE
		updateData := fixtures.AuthorUpdate()
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
		db.AssertNotExists("authors", id)
		db.AssertRowCount("authors", 0)

		// Try to get deleted author
		_, err = repo.Get(ctx, id)
		if err == nil {
			t.Error("Get should fail for deleted author")
		}
	})

	t.Run("Batch operations", func(t *testing.T) {
		db.Truncate("authors")

		// Add multiple authors
		sampleAuthors := fixtures.Authors()
		authorIDs := make([]string, 0, 5)

		for i := 0; i < 5; i++ {
			author := sampleAuthors[i]
			id, err := repo.Add(ctx, author)
			if err != nil {
				t.Fatalf("failed to add author %d: %v", i, err)
			}
			authorIDs = append(authorIDs, id)
		}

		// Verify count
		db.AssertRowCount("authors", 5)

		// List all
		authors, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(authors) != 5 {
			t.Errorf("expected 5 authors, got %d", len(authors))
		}

		// Delete all
		for _, id := range authorIDs {
			if err := repo.Delete(ctx, id); err != nil {
				t.Errorf("failed to delete author %s: %v", id, err)
			}
		}

		// Verify all deleted
		db.AssertRowCount("authors", 0)
	})

	t.Run("Update with partial data", func(t *testing.T) {
		db.Truncate("authors")

		// Create author
		testAuthor := fixtures.AuthorForCreate()
		id, err := repo.Add(ctx, testAuthor)
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}

		// Partial update (only FullName)
		newName := "Partially Updated Name"
		partialUpdate := fixtures.Author()
		partialUpdate.FullName = &newName
		partialUpdate.Specialty = nil // Don't update specialty

		err = repo.Update(ctx, id, partialUpdate)
		if err != nil {
			t.Fatalf("partial Update failed: %v", err)
		}

		// Verify only FullName changed
		retrieved, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if *retrieved.FullName != newName {
			t.Errorf("FullName not updated: expected %s, got %s", newName, *retrieved.FullName)
		}

		// Specialty should remain unchanged
		if retrieved.Specialty == nil || *retrieved.Specialty != *testAuthor.Specialty {
			t.Errorf("Specialty should not change: expected %v, got %v", testAuthor.Specialty, retrieved.Specialty)
		}
	})
}
