package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"library-service/internal/domain/member"
	"library-service/internal/infrastructure/store"
)

type MemberRepository struct {
	db *sqlx.DB
}

// NewMemberRepository creates a new instance of MemberRepository.
func NewMemberRepository(db *sqlx.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

// List retrieves all members from the store.
func (r *MemberRepository) List(ctx context.Context) ([]member.Member, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, books, created_at, updated_at, last_login_at
		FROM members
		ORDER BY id`

	var members []member.Member
	if err := r.db.SelectContext(ctx, &members, query); err != nil {
		return nil, err
	}
	return members, nil
}

// Add inserts a new member into the store.
func (r *MemberRepository) Add(ctx context.Context, data member.Member) (string, error) {
	query := `
		INSERT INTO members (email, password_hash, full_name, role, books, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	args := []interface{}{
		data.Email,
		data.PasswordHash,
		data.FullName,
		data.Role,
		pq.Array(data.Books),
		data.CreatedAt,
		data.UpdatedAt,
	}

	var id string
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

// Get retrieves a member by ID from the store.
func (r *MemberRepository) Get(ctx context.Context, id string) (member.Member, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, books, created_at, updated_at, last_login_at
		FROM members
		WHERE id=$1`

	var member member.Member
	if err := r.db.GetContext(ctx, &member, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return member, store.ErrorNotFound
		}
		return member, err
	}
	return member, nil
}

// Update modifies an existing member in the store.
func (r *MemberRepository) Update(ctx context.Context, id string, data member.Member) error {
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

// prepareUpdateArgs prepares the arguments for the update query.
func (r *MemberRepository) prepareUpdateArgs(data member.Member) ([]string, []interface{}) {
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

// Delete removes a member by ID from the store.
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

// GetByEmail retrieves a member by email
func (r *MemberRepository) GetByEmail(ctx context.Context, email string) (member.Member, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, books, created_at, updated_at, last_login_at
		FROM members
		WHERE email=$1`

	var m member.Member
	if err := r.db.GetContext(ctx, &m, query, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return m, store.ErrorNotFound
		}
		return m, err
	}
	return m, nil
}

// UpdateLastLogin updates last login time
func (r *MemberRepository) UpdateLastLogin(ctx context.Context, id string, loginTime time.Time) error {
	query := `
		UPDATE members
		SET last_login_at=$1, updated_at=CURRENT_TIMESTAMP
		WHERE id=$2
		RETURNING id`

	if err := r.db.QueryRowContext(ctx, query, loginTime, id).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.ErrorNotFound
		}
		return err
	}
	return nil
}

// EmailExists checks if email exists
func (r *MemberRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM members WHERE email=$1)`

	var exists bool
	if err := r.db.GetContext(ctx, &exists, query, email); err != nil {
		return false, err
	}
	return exists, nil
}
