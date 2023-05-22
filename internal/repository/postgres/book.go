package postgres

import (
	"context"
	"fmt"
	"strings"

	"library/internal/entity"
	"library/pkg/database"
)

type BookRepository struct {
	dataSourceName string
}

func NewBookRepository(dataSourceName string) *BookRepository {
	return &BookRepository{
		dataSourceName: dataSourceName,
	}
}

func (s *BookRepository) SelectRows(ctx context.Context) (dest []entity.Book, err error) {
	db, err := database.New(s.dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()

	query := `
		SELECT id, name, genre, isbn, authors
		FROM books
		ORDER BY id`

	err = db.SelectContext(ctx, &dest, query)

	return
}

func (s *BookRepository) CreateRow(ctx context.Context, data entity.Book) (id string, err error) {
	db, err := database.New(s.dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()

	query := `
		INSERT INTO books (name, genre, isbn, authors)
		VALUES ($1, $2, $3)
		RETURNING id`

	args := []any{data.Name, data.Genre, data.ISBN, data.Authors}

	err = db.QueryRowContext(ctx, query, args...).Scan(&id)

	return
}

func (s *BookRepository) GetRow(ctx context.Context, id string) (dest entity.Book, err error) {
	db, err := database.New(s.dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()

	query := `
		SELECT id, name, genre, isbn, authors
		FROM books
		WHERE id=$1`

	args := []any{id}

	err = db.GetContext(ctx, &dest, query, args...)

	return
}

func (s *BookRepository) UpdateRow(ctx context.Context, id string, data entity.Book) (err error) {
	db, err := database.New(s.dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()

	sets, args := s.prepareArgs(data)
	if len(args) > 0 {

		args = append(args, id)
		sets = append(sets, "updated_at=CURRENT_TIMESTAMP")

		query := fmt.Sprintf("UPDATE books SET %s WHERE id=$%d", strings.Join(sets, ", "), len(args))
		_, err = db.ExecContext(ctx, query, args...)
	}

	return
}

func (s *BookRepository) prepareArgs(data entity.Book) (sets []string, args []any) {
	if data.Name != nil {
		args = append(args, data.Name)
		sets = append(sets, fmt.Sprintf("name=$%d", len(args)))
	}

	if data.Genre != nil {
		args = append(args, data.Genre)
		sets = append(sets, fmt.Sprintf("genre=$%d", len(args)))
	}

	if data.ISBN != nil {
		args = append(args, data.ISBN)
		sets = append(sets, fmt.Sprintf("isbn=$%d", len(args)))
	}

	if len(data.Authors) > 0 {
		args = append(args, data.Authors)
		sets = append(sets, fmt.Sprintf("authors=$%d", len(args)))
	}

	return
}

func (s *BookRepository) DeleteRow(ctx context.Context, id string) (err error) {
	db, err := database.New(s.dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()

	query := `
		DELETE 
		FROM books
		WHERE id=$1`

	args := []any{id}

	_, err = db.ExecContext(ctx, query, args...)

	return
}
