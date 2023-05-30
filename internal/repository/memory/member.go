package memory

import (
	"context"
	"database/sql"
	"sync"

	"github.com/google/uuid"

	"library/internal/domain/member"
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

func (r *MemberRepository) SelectRows(ctx context.Context) ([]member.Entity, error) {
	r.RLock()
	defer r.RUnlock()

	rows := make([]member.Entity, 0, len(r.db))
	for _, data := range r.db {
		rows = append(rows, data)
	}

	return rows, nil
}

func (r *MemberRepository) CreateRow(ctx context.Context, data member.Entity) (string, error) {
	r.Lock()
	defer r.Unlock()

	id := r.generateID()
	data.ID = id
	r.db[id] = data

	return id, nil
}

func (r *MemberRepository) GetRow(ctx context.Context, id string) (data member.Entity, err error) {
	r.RLock()
	defer r.RUnlock()

	data, ok := r.db[id]
	if !ok {
		err = sql.ErrNoRows
		return
	}

	return
}

func (r *MemberRepository) UpdateRow(ctx context.Context, id string, data member.Entity) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	r.db[id] = data

	return nil
}

func (r *MemberRepository) DeleteRow(ctx context.Context, id string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.db, id)

	return nil
}

func (r *MemberRepository) generateID() string {
	return uuid.New().String()
}
