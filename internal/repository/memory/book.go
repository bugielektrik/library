package memory

import (
	"context"
	"database/sql"
	"sync"

	"github.com/google/uuid"

	"library-service/internal/domain/book"
)

// BookRepository handles CRUD operations for books in an in-memory database.
type BookRepository struct {
	db map[string]book.Entity
	sync.RWMutex
}

// NewBookRepository creates a new BookRepository.
func NewBookRepository() *BookRepository {
	return &BookRepository{db: make(map[string]book.Entity)}
}

// List retrieves all books from the in-memory database.
func (r *BookRepository) List(ctx context.Context) ([]book.Entity, error) {
	r.RLock()
	defer r.RUnlock()

	books := make([]book.Entity, 0, len(r.db))
	for _, data := range r.db {
		books = append(books, data)
	}
	return books, nil
}

// Add inserts a new book into the in-memory database.
func (r *BookRepository) Add(ctx context.Context, data book.Entity) (string, error) {
	r.Lock()
	defer r.Unlock()

	id := uuid.New().String()
	data.ID = id
	r.db[id] = data
	return id, nil
}

// Get retrieves a book by ID from the in-memory database.
func (r *BookRepository) Get(ctx context.Context, id string) (book.Entity, error) {
	r.RLock()
	defer r.RUnlock()

	data, ok := r.db[id]
	if !ok {
		return book.Entity{}, sql.ErrNoRows
	}
	return data, nil
}

// Update modifies an existing book in the in-memory database.
func (r *BookRepository) Update(ctx context.Context, id string, data book.Entity) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	r.db[id] = data
	return nil
}

// Delete removes a book by ID from the in-memory database.
func (r *BookRepository) Delete(ctx context.Context, id string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.db, id)
	return nil
}
