package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"library-service/internal/domain/member"
	"library-service/pkg/store"
)

type MemberRepository struct {
	db *sqlx.DB
}

func NewMemberRepository(db *sqlx.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) List(ctx context.Context) ([]member.Entity, error) {
	query := `
		SELECT id, full_name, books
		FROM members
		ORDER BY id`

	var members []member.Entity
	if err := r.db.SelectContext(ctx, &members, query); err != nil {
		return nil, err
	}
	return members, nil
}

func (r *MemberRepository) Add(ctx context.Context, data member.Entity) (string, error) {
	query := `
		INSERT INTO members (full_name, books)
		VALUES ($1, $2)
		RETURNING id`

	args := []interface{}{data.FullName, pq.Array(data.Books)}

	var id string
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", store.ErrorNotFound
		}
		return "", err
	}
	return id, nil
}

func (r *MemberRepository) Get(ctx context.Context, id string) (member.Entity, error) {
	query := `
		SELECT id, full_name, books
		FROM members
		WHERE id=$1`

	var member member.Entity
	if err := r.db.GetContext(ctx, &member, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return member, store.ErrorNotFound
		}
		return member, err
	}
	return member, nil
}

func (r *MemberRepository) Update(ctx context.Context, id string, data member.Entity) error {
	sets, args := r.prepareUpdateArgs(data)
	if len(args) == 0 {
		return nil
	}

	args = append(args, id)
	sets = append(sets, "updated_at=CURRENT_TIMESTAMP")
	query := fmt.Sprintf("UPDATE members SET %s WHERE id=$%d RETURNING id", strings.Join(sets, ", "), len(args))

	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.ErrorNotFound
		}
		return err
	}
	return nil
}

func (r *MemberRepository) prepareUpdateArgs(data member.Entity) ([]string, []interface{}) {
	var sets []string
	var args []interface{}

	if data.FullName != nil {
		args = append(args, data.FullName)
		sets = append(sets, fmt.Sprintf("full_name=$%d", len(args)))
	}

	if len(data.Books) > 0 {
		args = append(args, pq.Array(data.Books))
		sets = append(sets, fmt.Sprintf("books=$%d", len(args)))
	}

	return sets, args
}

func (r *MemberRepository) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM members
		WHERE id=$1
		RETURNING id`

	if err := r.db.QueryRowContext(ctx, query, id).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.ErrorNotFound
		}
		return err
	}
	return nil
}
