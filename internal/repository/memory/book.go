package memory

import (
	"database/sql"
	"library/internal/entity"
	"sync"

	"github.com/google/uuid"
)

// BookRepository is an in-memory implementation of the BookRepository interface
type BookRepository struct {
	db map[string]entity.Book
	sync.RWMutex
}

// NewBookRepository creates a new instance of the BookRepository struct
func NewBookRepository() *BookRepository {
	return &BookRepository{
		db: make(map[string]entity.Book),
	}
}

// CreateRow creates a new row in the in-memory storage
func (r *BookRepository) CreateRow(data entity.Book) (string, error) {
	r.Lock()
	defer r.Unlock()

	id := r.generateID()
	r.db[id] = data

	return id, nil
}

// GetRowByID retrieves a row from the in-memory storage by ID
func (r *BookRepository) GetRowByID(id string) (data entity.Book, err error) {
	r.RLock()
	defer r.RUnlock()

	data, ok := r.db[id]
	if !ok {
		err = sql.ErrNoRows
		return
	}

	return
}

// SelectRows retrieves all rows from the in-memory storage
func (r *BookRepository) SelectRows() ([]entity.Book, error) {
	r.RLock()
	defer r.RUnlock()

	rows := make([]entity.Book, 0, len(r.db))
	for _, data := range r.db {
		rows = append(rows, data)
	}

	return rows, nil
}

// UpdateRow updates an existing row in the in-memory storage
func (r *BookRepository) UpdateRow(id string, data entity.Book) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	r.db[id] = data

	return nil
}

// DeleteRow deletes a row from the in-memory storage by ID
func (r *BookRepository) DeleteRow(id string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.db, id)

	return nil
}

// generateID generates a unique ID for the row
func (r *BookRepository) generateID() string {
	return uuid.New().String()
}
