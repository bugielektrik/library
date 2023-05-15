package memory

import (
	"database/sql"
	"library/internal/entity"
	"sync"

	"github.com/google/uuid"
)

// MemberRepository is an in-memory implementation of the MemberRepository interface
type MemberRepository struct {
	db map[string]entity.Member
	sync.RWMutex
}

// NewMemberRepository creates a new instance of the MemberRepository struct
func NewMemberRepository() *MemberRepository {
	return &MemberRepository{
		db: make(map[string]entity.Member),
	}
}

// CreateRow creates a new row in the in-memory storage
func (r *MemberRepository) CreateRow(data entity.Member) (string, error) {
	r.Lock()
	defer r.Unlock()

	id := generateID()
	r.db[id] = data

	return id, nil
}

// GetRowByID retrieves a row from the in-memory storage by ID
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

// SelectRows retrieves all rows from the in-memory storage
func (r *MemberRepository) SelectRows() ([]entity.Member, error) {
	r.RLock()
	defer r.RUnlock()

	rows := make([]entity.Member, 0, len(r.db))
	for _, data := range r.db {
		rows = append(rows, data)
	}

	return rows, nil
}

// UpdateRow updates an existing row in the in-memory storage
func (r *MemberRepository) UpdateRow(id string, data entity.Member) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	r.db[id] = data

	return nil
}

// DeleteRow deletes a row from the in-memory storage by ID
func (r *MemberRepository) DeleteRow(id string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.db, id)

	return nil
}

// generateID generates a unique ID for the row
func generateID() string {
	return uuid.New().String()
}
