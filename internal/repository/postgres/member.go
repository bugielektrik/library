package postgres

import (
	"context"
	"fmt"
	"strings"

	"library/internal/entity"
	"library/pkg/database"
)

type MemberRepository struct {
	dataSourceName string
}

func NewMemberRepository(dataSourceName string) *MemberRepository {
	return &MemberRepository{
		dataSourceName: dataSourceName,
	}
}

func (s *MemberRepository) SelectRows(ctx context.Context) (dest []entity.Member, err error) {
	db, err := database.New(s.dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()

	query := `
		SELECT id, full_name, books
		FROM members
		ORDER BY id`

	err = db.SelectContext(ctx, &dest, query)

	return
}

func (s *MemberRepository) CreateRow(ctx context.Context, data entity.Member) (id string, err error) {
	db, err := database.New(s.dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()

	query := `
		INSERT INTO members (full_name, books)
		VALUES ($1, $2)
		RETURNING id`

	args := []any{data.FullName, data.Books}

	err = db.QueryRowContext(ctx, query, args...).Scan(&id)

	return
}

func (s *MemberRepository) GetRow(ctx context.Context, id string) (dest entity.Member, err error) {
	db, err := database.New(s.dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()

	query := `
		SELECT id, full_name, books
		FROM members
		WHERE id=$1`

	args := []any{id}

	err = db.GetContext(ctx, &dest, query, args...)

	return
}

func (s *MemberRepository) UpdateRow(ctx context.Context, id string, data entity.Member) (err error) {
	db, err := database.New(s.dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()

	sets, args := s.prepareArgs(data)
	if len(args) > 0 {

		args = append(args, id)
		sets = append(sets, "updated_at=CURRENT_TIMESTAMP")

		query := fmt.Sprintf("UPDATE members SET %s WHERE id=$%d", strings.Join(sets, ", "), len(args))
		_, err = db.ExecContext(ctx, query, args...)
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
	db, err := database.New(s.dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()

	query := `
		DELETE 
		FROM members
		WHERE id=$1`

	args := []any{id}

	_, err = db.ExecContext(ctx, query, args...)

	return
}
