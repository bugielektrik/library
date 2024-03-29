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
	return &AuthorRepository{
		db: make(map[string]author.Entity),
	}
}

func (r *AuthorRepository) List(ctx context.Context) (dest []author.Entity, err error) {
	r.RLock()
	defer r.RUnlock()

	dest = make([]author.Entity, 0, len(r.db))
	for _, data := range r.db {
		dest = append(dest, data)
	}

	return
}

func (r *AuthorRepository) Add(ctx context.Context, data author.Entity) (dest string, err error) {
	r.Lock()
	defer r.Unlock()

	id := r.generateID()
	data.ID = id
	r.db[id] = data

	return id, nil
}

func (r *AuthorRepository) Get(ctx context.Context, id string) (dest author.Entity, err error) {
	r.RLock()
	defer r.RUnlock()

	dest, ok := r.db[id]
	if !ok {
		err = sql.ErrNoRows
		return
	}

	return
}

func (r *AuthorRepository) Update(ctx context.Context, id string, data author.Entity) (err error) {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	r.db[id] = data

	return
}

func (r *AuthorRepository) Delete(ctx context.Context, id string) (err error) {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.db, id)

	return
}

func (r *AuthorRepository) generateID() string {
	return uuid.New().String()
}
