package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

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
func (s *MemberRepository) CreateRow(data entity.Member) (dest string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		INSERT INTO readers (fullname,booklist)
		VALUES ($1, $2)
		RETURNING id`

	args := []any{data.FullName, data.Books}

	err = s.db.QueryRowContext(ctx, query, args...).Scan(&dest)

	return
}

func (s *MemberRepository) GetRowByID(id string) (dest entity.Member, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT id,fullname,booklist 
		FROM readers
		WHERE id=$1`

	args := []any{id}

	err = s.db.GetContext(ctx, &dest, query, args...)

	return
}

func (s *MemberRepository) SelectRows() (dest []entity.Member, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT id, fullname,booklist
		FROM readers
		ORDER BY id`

	err = s.db.SelectContext(ctx, &dest, query)

	return
}

func (s *MemberRepository) UpdateRow(data entity.Member) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sets, args := s.prepareArgs(data)
	if len(args) > 0 {

		args = append(args, data.ID)
		sets = append(sets, "updated_at=CURRENT_TIMESTAMP")

		query := fmt.Sprintf("UPDATE readers SET %s WHERE id=$%d", strings.Join(sets, ", "), len(args))
		_, err = s.db.ExecContext(ctx, query, args...)
	}
	return
}

func (s *MemberRepository) prepareArgs(data entity.Member) (sets []string, args []any) {
	if data.Books != nil {
		args = append(args, data.Books)
		sets = append(sets, fmt.Sprintf("bookList=$%d", len(args)))
	}
	if data.FullName != nil {
		args = append(args, data.FullName)
		sets = append(sets, fmt.Sprintf("fullName=$%d", len(args)))
	}
	return
}

func (s *MemberRepository) DeleteRow(id string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		DELETE 
		FROM readers
		WHERE id=$1`

	args := []any{id}

	_, err = s.db.ExecContext(ctx, query, args...)

	return
}
