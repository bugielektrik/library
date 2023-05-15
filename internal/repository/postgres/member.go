package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"library/internal/entity"
)

type MemberRepository struct {
	db *sqlx.DB
}

func NewMemberRepository(db *sqlx.DB) *MemberRepository {
	return &MemberRepository{
		db: db,
	}
}

func (s *MemberRepository) CreateRow(ctx context.Context, data entity.Member) (id string, err error) {
	query := `
		INSERT INTO members (full_name, books)
		VALUES ($1, $2)
		RETURNING id`

	args := []any{data.FullName, data.Books}

	err = s.db.QueryRowContext(ctx, query, args...).Scan(&id)

	return
}

func (s *MemberRepository) GetRowByID(ctx context.Context, id string) (dest entity.Member, err error) {
	query := `
		SELECT id, full_name, books
		FROM members
		WHERE id=$1`

	args := []any{id}

	err = s.db.GetContext(ctx, &dest, query, args...)

	return
}

func (s *MemberRepository) SelectRows(ctx context.Context) (dest []entity.Member, err error) {
	query := `
		SELECT id, full_name, books
		FROM members
		ORDER BY id`

	err = s.db.SelectContext(ctx, &dest, query)

	return
}

func (s *MemberRepository) UpdateRow(ctx context.Context, id string, data entity.Member) (err error) {
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

func (s *MemberRepository) DeleteRow(ctx context.Context, id string) (err error) {
	query := `
		DELETE 
		FROM members
		WHERE id=$1`

	args := []any{id}

	_, err = s.db.ExecContext(ctx, query, args...)

	return
}
