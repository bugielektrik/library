//go:build integration

package integration

import (
	"context"
	"testing"

	"library-service/internal/adapters/repository/postgres"
	"library-service/internal/domain/book"
	"library-service/test/fixtures"
)

func TestBookRepository_Integration(t *testing.T) {
	db := Setup(t)
	defer db.Cleanup()

	// Clean up test data before running tests
	db.TruncateAll()

	repo := postgres.NewBookRepository(db.DB)
	ctx := context.Background()

	t.Run("Add and Get book", func(t *testing.T) {
		// Create test book
		testBook := fixtures.BookForCreate()

		// Add to repository
		id, err := repo.Add(ctx, testBook)
		if err != nil {
			t.Fatalf("failed to add book: %v", err)
		}

		if id == "" {
			t.Fatal("expected non-empty ID from Add")
		}

		// Verify it exists
		db.AssertExists("books", id)

		// Retrieve the book
		retrieved, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("failed to get book: %v", err)
		}

		// Verify data matches
		if retrieved.ID != id {
			t.Errorf("expected ID %s, got %s", id, retrieved.ID)
		}

		if retrieved.Name == nil || *retrieved.Name != *testBook.Name {
			t.Errorf("name mismatch: expected %v, got %v", testBook.Name, retrieved.Name)
		}

		if retrieved.ISBN == nil || *retrieved.ISBN != *testBook.ISBN {
			t.Errorf("ISBN mismatch: expected %v, got %v", testBook.ISBN, retrieved.ISBN)
		}
	})

	t.Run("List books", func(t *testing.T) {
		db.Truncate("books")

		// Add multiple books
		book1 := fixtures.BookForCreate()
		book2 := fixtures.BookForCreate()
		name2 := "Second Book"
		book2.Name = &name2

		_, err := repo.Add(ctx, book1)
		if err != nil {
			t.Fatalf("failed to add book1: %v", err)
		}

		_, err = repo.Add(ctx, book2)
		if err != nil {
			t.Fatalf("failed to add book2: %v", err)
		}

		// List all books
		books, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("failed to list books: %v", err)
		}

		if len(books) != 2 {
			t.Errorf("expected 2 books, got %d", len(books))
		}
	})

	t.Run("Update book", func(t *testing.T) {
		db.Truncate("books")

		// Add book
		testBook := fixtures.BookForCreate()
		id, err := repo.Add(ctx, testBook)
		if err != nil {
			t.Fatalf("failed to add book: %v", err)
		}

		// Update book
		updateData := fixtures.BookUpdate()
		err = repo.Update(ctx, id, updateData)
		if err != nil {
			t.Fatalf("failed to update book: %v", err)
		}

		// Verify update
		retrieved, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("failed to get book after update: %v", err)
		}

		if retrieved.Name == nil || *retrieved.Name != *updateData.Name {
			t.Errorf("name not updated: expected %v, got %v", updateData.Name, retrieved.Name)
		}
	})

	t.Run("Delete book", func(t *testing.T) {
		db.Truncate("books")

		// Add book
		testBook := fixtures.BookForCreate()
		id, err := repo.Add(ctx, testBook)
		if err != nil {
			t.Fatalf("failed to add book: %v", err)
		}

		// Verify it exists
		db.AssertExists("books", id)

		// Delete book
		err = repo.Delete(ctx, id)
		if err != nil {
			t.Fatalf("failed to delete book: %v", err)
		}

		// Verify it's gone
		db.AssertNotExists("books", id)

		// Try to get deleted book (should fail)
		_, err = repo.Get(ctx, id)
		if err == nil {
			t.Error("expected error when getting deleted book")
		}
	})

	t.Run("CRUD workflow", func(t *testing.T) {
		db.Truncate("books")

		// Create
		testBook := fixtures.BookForCreate()
		id, err := repo.Add(ctx, testBook)
		if err != nil {
			t.Fatalf("CREATE failed: %v", err)
		}

		// Read
		retrieved, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("READ failed: %v", err)
		}

		if retrieved.ID != id {
			t.Errorf("READ returned wrong book: expected %s, got %s", id, retrieved.ID)
		}

		// Update
		updateData := book.Book{
			Name: stringPtr("Updated Title"),
		}
		err = repo.Update(ctx, id, updateData)
		if err != nil {
			t.Fatalf("UPDATE failed: %v", err)
		}

		// Verify update
		retrieved, err = repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("READ after UPDATE failed: %v", err)
		}

		if *retrieved.Name != "Updated Title" {
			t.Errorf("UPDATE didn't apply: expected 'Updated Title', got %s", *retrieved.Name)
		}

		// Delete
		err = repo.Delete(ctx, id)
		if err != nil {
			t.Fatalf("DELETE failed: %v", err)
		}

		// Verify deletion
		_, err = repo.Get(ctx, id)
		if err == nil {
			t.Error("DELETE didn't work: book still exists")
		}
	})
}
