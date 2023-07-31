package memory

import (
	"context"
	"database/sql"
	"sync"

	"github.com/google/uuid"

	"library-service/internal/domain/member"
)

type MemberRepository struct {
	db map[string]member.Entity
	sync.RWMutex
}

func NewMemberRepository() *MemberRepository {
	return &MemberRepository{
		db: make(map[string]member.Entity),
	}
}

func (r *MemberRepository) Select(ctx context.Context) (dest []member.Entity, err error) {
	r.RLock()
	defer r.RUnlock()

	dest = make([]member.Entity, 0, len(r.db))
	for _, data := range r.db {
		dest = append(dest, data)
	}

	return
}

func (r *MemberRepository) Insert(ctx context.Context, data member.Entity) (dest string, err error) {
	r.Lock()
	defer r.Unlock()

	id := r.generateID()
	data.ID = id
	r.db[id] = data

	return id, nil
}

func (r *MemberRepository) Get(ctx context.Context, id string) (dest member.Entity, err error) {
	r.RLock()
	defer r.RUnlock()

	dest, ok := r.db[id]
	if !ok {
		err = sql.ErrNoRows
		return
	}

	return
}

func (r *MemberRepository) Update(ctx context.Context, id string, data member.Entity) (err error) {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	r.db[id] = data

	return
}

func (r *MemberRepository) Delete(ctx context.Context, id string) (err error) {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.db, id)

	return
}

func (r *MemberRepository) generateID() string {
	return uuid.New().String()
}
