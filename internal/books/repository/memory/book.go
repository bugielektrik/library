package memory

import (
	"context"
	"database/sql"
	"sync"

	"github.com/google/uuid"

	"library-service/internal/books/domain/book"
)

// BookRepository handles CRUD operations for books in an in-memory store.
type BookRepository struct {
	db map[string]book.Book
	sync.RWMutex
}

// NewBookRepository creates a new BookRepository.
func NewBookRepository() *BookRepository {
	return &BookRepository{db: make(map[string]book.Book)}
}

// List retrieves all books from the in-memory store.
func (r *BookRepository) List(ctx context.Context) ([]book.Book, error) {
	r.RLock()
	defer r.RUnlock()

	books := make([]book.Book, 0, len(r.db))
	for _, data := range r.db {
		books = append(books, data)
	}
	return books, nil
}

// Add inserts a new book into the in-memory store.
func (r *BookRepository) Add(ctx context.Context, data book.Book) (string, error) {
	r.Lock()
	defer r.Unlock()

	id := uuid.New().String()
	data.ID = id
	r.db[id] = data
	return id, nil
}

// Get retrieves a book by ID from the in-memory store.
func (r *BookRepository) Get(ctx context.Context, id string) (book.Book, error) {
	r.RLock()
	defer r.RUnlock()

	data, ok := r.db[id]
	if !ok {
		return book.Book{}, sql.ErrNoRows
	}
	return data, nil
}

// Update modifies an existing book in the in-memory store.
func (r *BookRepository) Update(ctx context.Context, id string, data book.Book) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	r.db[id] = data
	return nil
}

// Delete removes a book by ID from the in-memory store.
func (r *BookRepository) Delete(ctx context.Context, id string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.db, id)
	return nil
}
