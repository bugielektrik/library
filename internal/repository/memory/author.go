package memory

import (
	"context"
	"database/sql"
	"library/internal/entity"
	"sync"

	"github.com/google/uuid"
)

type AuthorRepository struct {
	db map[string]entity.Author
	sync.RWMutex
}

func NewAuthorRepository() *AuthorRepository {
	return &AuthorRepository{
		db: make(map[string]entity.Author),
	}
}

func (r *AuthorRepository) CreateRow(ctx context.Context, data entity.Author) (dest string, err error) {
	r.Lock()
	defer r.Unlock()

	id := r.generateID()
	data.ID = id
	r.db[id] = data

	return id, nil
}

func (r *AuthorRepository) GetRowByID(ctx context.Context, id string) (data entity.Author, err error) {
	r.RLock()
	defer r.RUnlock()

	data, ok := r.db[id]
	if !ok {
		err = sql.ErrNoRows
		return
	}

	return
}

func (r *AuthorRepository) SelectRows(ctx context.Context) ([]entity.Author, error) {
	r.RLock()
	defer r.RUnlock()

	rows := make([]entity.Author, 0, len(r.db))
	for _, data := range r.db {
		rows = append(rows, data)
	}

	return rows, nil
}

func (r *AuthorRepository) UpdateRow(ctx context.Context, id string, data entity.Author) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	r.db[id] = data

	return nil
}

func (r *AuthorRepository) DeleteRow(ctx context.Context, id string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.db, id)

	return nil
}

func (r *AuthorRepository) generateID() string {
	return uuid.New().String()
}
