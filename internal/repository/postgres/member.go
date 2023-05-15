package postgres

import (
	"context"
	"fmt"
	"library/internal/entity"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// MemberRepository is a postgres implementation of the MemberRepository interface
type MemberRepository struct {
	db *sqlx.DB
}

// NewMemberRepository creates a new instance of the MemberRepository struct
func NewMemberRepository(db *sqlx.DB) *MemberRepository {
	return &MemberRepository{
		db: db,
	}
}

// CreateRow creates a new row in the postgres database
func (s *MemberRepository) CreateRow(data entity.Member) (id string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		INSERT INTO members (full_name, books)
		VALUES ($1, $2)
		RETURNING id`

	args := []any{data.FullName, data.Books}

	err = s.db.QueryRowContext(ctx, query, args...).Scan(&id)

	return
}

// GetRowByID retrieves a row from the postgres database by ID
func (s *MemberRepository) GetRowByID(id string) (dest entity.Member, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT id, full_name, books
		FROM members
		WHERE id=$1`

	args := []any{id}

	err = s.db.GetContext(ctx, &dest, query, args...)

	return
}

// SelectRows retrieves all rows from the postgres database
func (s *MemberRepository) SelectRows() (dest []entity.Member, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT id, full_name, books
		FROM members
		ORDER BY id`

	err = s.db.SelectContext(ctx, &dest, query)

	return
}

// UpdateRow updates an existing row in the postgres database
func (s *MemberRepository) UpdateRow(id string, data entity.Member) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sets, args := s.prepareArgs(data)
	if len(args) > 0 {

		args = append(args, id)
		sets = append(sets, "updated_at=CURRENT_TIMESTAMP")

		query := fmt.Sprintf("UPDATE members SET %s WHERE id=$%d", strings.Join(sets, ", "), len(args))
		_, err = s.db.ExecContext(ctx, query, args...)
	}

	return
}

func (s *MemberRepository) prepareArgs(data entity.Member) (sets []string, args []any) {
	if data.FullName != nil {
		args = append(args, data.FullName)
		sets = append(sets, fmt.Sprintf("full_name=$%d", len(args)))
	}

	if len(data.Books) > 0 {
		args = append(args, data.Books)
		sets = append(sets, fmt.Sprintf("books=$%d", len(args)))
	}

	return
}

// DeleteRow deletes a row from the postgres database by ID
func (s *MemberRepository) DeleteRow(id string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		DELETE 
		FROM members
		WHERE id=$1`

	args := []any{id}

	_, err = s.db.ExecContext(ctx, query, args...)

	return
}
