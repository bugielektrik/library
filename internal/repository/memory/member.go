package memory

import (
	"database/sql"
	"library/internal/entity"
	"sync"

	"github.com/google/uuid"
)

type MemberRepository struct {
	db map[string]entity.Member
	sync.RWMutex
}

func NewMemberRepository() *MemberRepository {
	return &MemberRepository{
		db: make(map[string]entity.Member),
	}
}

func (r *MemberRepository) CreateRow(data entity.Member) (string, error) {
	r.Lock()
	defer r.Unlock()

	id := r.generateID()
	data.ID = id
	r.db[id] = data

	return id, nil
}

func (r *MemberRepository) GetRowByID(id string) (data entity.Member, err error) {
	r.RLock()
	defer r.RUnlock()

	data, ok := r.db[id]
	if !ok {
		err = sql.ErrNoRows
		return
	}

	return
}

func (r *MemberRepository) SelectRows() ([]entity.Member, error) {
	r.RLock()
	defer r.RUnlock()

	rows := make([]entity.Member, 0, len(r.db))
	for _, data := range r.db {
		rows = append(rows, data)
	}

	return rows, nil
}

func (r *MemberRepository) UpdateRow(id string, data entity.Member) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	r.db[id] = data

	return nil
}

func (r *MemberRepository) DeleteRow(id string) error {
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
