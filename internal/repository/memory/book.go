package memory

import (
	"database/sql"
	"library/internal/entity"
	"sync"

	"github.com/google/uuid"
)

type BookRepository struct {
	db map[string]entity.Book
	sync.RWMutex
}

func NewBookRepository() *BookRepository {
	return &BookRepository{
		db: make(map[string]entity.Book),
	}
}

func (r *BookRepository) CreateRow(data entity.Book) (string, error) {
	r.Lock()
	defer r.Unlock()

	id := r.generateID()
	data.ID = id
	r.db[id] = data

	return id, nil
}

func (r *BookRepository) GetRowByID(id string) (data entity.Book, err error) {
	r.RLock()
	defer r.RUnlock()

	data, ok := r.db[id]
	if !ok {
		err = sql.ErrNoRows
		return
	}

	return
}

func (r *BookRepository) SelectRows() ([]entity.Book, error) {
	r.RLock()
	defer r.RUnlock()

	rows := make([]entity.Book, 0, len(r.db))
	for _, data := range r.db {
		rows = append(rows, data)
	}

	return rows, nil
}

func (r *BookRepository) UpdateRow(id string, data entity.Book) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	r.db[id] = data

	return nil
}

func (r *BookRepository) DeleteRow(id string) error {
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
