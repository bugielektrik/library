package memory

import (
	"context"
	"database/sql"
	"sync"

	"github.com/google/uuid"

	"library-service/internal/domain/member"
)

// MemberRepository provides an in-memory implementation of the member repository.
type MemberRepository struct {
	db map[string]member.Entity
	mu sync.RWMutex
}

// NewMemberRepository creates a new instance of MemberRepository.
func NewMemberRepository() *MemberRepository {
	return &MemberRepository{
		db: make(map[string]member.Entity),
	}
}

// List retrieves all members from the in-memory store.
func (r *MemberRepository) List(ctx context.Context) ([]member.Entity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	members := make([]member.Entity, 0, len(r.db))
	for _, entity := range r.db {
		members = append(members, entity)
	}

	return members, nil
}

// Add inserts a new member into the in-memory store.
func (r *MemberRepository) Add(ctx context.Context, entity member.Entity) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := uuid.New().String()
	entity.ID = id
	r.db[id] = entity

	return id, nil
}

// Get retrieves a member by ID from the in-memory store.
func (r *MemberRepository) Get(ctx context.Context, id string) (member.Entity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entity, exists := r.db[id]
	if !exists {
		return member.Entity{}, sql.ErrNoRows
	}

	return entity, nil
}

// Update modifies an existing member in the in-memory store.
func (r *MemberRepository) Update(ctx context.Context, id string, entity member.Entity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.db[id]; !exists {
		return sql.ErrNoRows
	}
	r.db[id] = entity

	return nil
}

// Delete removes a member by ID from the in-memory store.
func (r *MemberRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.db[id]; !exists {
		return sql.ErrNoRows
	}
	delete(r.db, id)

	return nil
}
