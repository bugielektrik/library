package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"library/internal/domain/author"
	"library/pkg/store"
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

func (r *AuthorRepository) Select(ctx context.Context) (dest []author.Entity, err error) {
	r.RLock()
	defer r.RUnlock()

	dest = make([]author.Entity, 0, len(r.db))
	for _, data := range r.db {
		dest = append(dest, data)
	}

	return
}

func (r *AuthorRepository) Create(ctx context.Context, data author.Entity) (dest string, err error) {
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
		err = store.ErrorNotFound
		return
	}

	return
}

func (r *AuthorRepository) Update(ctx context.Context, id string, data author.Entity) (err error) {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return store.ErrorNotFound
	}
	r.db[id] = data

	return
}

func (r *AuthorRepository) Delete(ctx context.Context, id string) (err error) {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return store.ErrorNotFound
	}
	delete(r.db, id)

	return
}

func (r *AuthorRepository) generateID() string {
	return uuid.New().String()
}
