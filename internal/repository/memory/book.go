package memory

import (
	"context"
	"database/sql"
	"sync"

	"github.com/google/uuid"

	"library/internal/domain/book"
)

type BookRepository struct {
	db map[string]book.Entity
	sync.RWMutex
}

func NewBookRepository() *BookRepository {
	return &BookRepository{
		db: make(map[string]book.Entity),
	}
}

func (r *BookRepository) SelectRows(ctx context.Context) ([]book.Entity, error) {
	r.RLock()
	defer r.RUnlock()

	rows := make([]book.Entity, 0, len(r.db))
	for _, data := range r.db {
		rows = append(rows, data)
	}

	return rows, nil
}

func (r *BookRepository) CreateRow(ctx context.Context, data book.Entity) (string, error) {
	r.Lock()
	defer r.Unlock()

	id := r.generateID()
	data.ID = id
	r.db[id] = data

	return id, nil
}

func (r *BookRepository) GetRow(ctx context.Context, id string) (data book.Entity, err error) {
	r.RLock()
	defer r.RUnlock()

	data, ok := r.db[id]
	if !ok {
		err = sql.ErrNoRows
		return
	}

	return
}

func (r *BookRepository) UpdateRow(ctx context.Context, id string, data book.Entity) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	r.db[id] = data

	return nil
}

func (r *BookRepository) DeleteRow(ctx context.Context, id string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.db, id)

	return nil
}

func (r *BookRepository) generateID() string {
	return uuid.New().String()
}
