package memory

import (
	"context"
	"database/sql"
	"sync"

	"github.com/google/uuid"

	"library-service/internal/domain/author"
)

type AuthorRepository struct {
	db map[string]author.Entity
	sync.RWMutex
}

func NewAuthorRepository() *AuthorRepository {
	return &AuthorRepository{db: make(map[string]author.Entity)}
}

func (r *AuthorRepository) List(ctx context.Context) ([]author.Entity, error) {
	r.RLock()
	defer r.RUnlock()

	authors := make([]author.Entity, 0, len(r.db))
	for _, data := range r.db {
		authors = append(authors, data)
	}
	return authors, nil
}

func (r *AuthorRepository) Add(ctx context.Context, data author.Entity) (string, error) {
	r.Lock()
	defer r.Unlock()

	id := uuid.New().String()
	data.ID = id
	r.db[id] = data
	return id, nil
}

func (r *AuthorRepository) Get(ctx context.Context, id string) (author.Entity, error) {
	r.RLock()
	defer r.RUnlock()

	data, ok := r.db[id]
	if !ok {
		return author.Entity{}, sql.ErrNoRows
	}
	return data, nil
}

func (r *AuthorRepository) Update(ctx context.Context, id string, data author.Entity) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	r.db[id] = data
	return nil
}

func (r *AuthorRepository) Delete(ctx context.Context, id string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.db, id)
	return nil
}
