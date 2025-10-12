package book

import "context"

// Repository defines the interface for book repository service.
type Repository interface {
	// List retrieves all books.
	List(ctx context.Context) ([]Book, error)

	// Add inserts a new book and returns its ID.
	Add(ctx context.Context, data Book) (string, error)

	// Get retrieves a book by its ID.
	Get(ctx context.Context, id string) (Book, error)

	// Update modifies an existing book by its ID.
	Update(ctx context.Context, id string, data Book) error

	// Delete removes a book by its ID.
	Delete(ctx context.Context, id string) error
}
