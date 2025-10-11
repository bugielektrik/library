package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"library-service/internal/domain/member"
)

type MemberRepository struct {
	BaseRepository[member.Member]
}

// NewMemberRepository creates a new instance of MemberRepository.
func NewMemberRepository(db *sqlx.DB) *MemberRepository {
	return &MemberRepository{
		BaseRepository: NewBaseRepository[member.Member](db, "members"),
	}
}

// List is inherited from BaseRepository

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
	if err := r.GetDB().QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
		return "", fmt.Errorf("inserting member: %w", HandleSQLError(err))
	}
	return id, nil
}

// Get is inherited from BaseRepository

// Update modifies an existing member in the store.
func (r *MemberRepository) Update(ctx context.Context, id string, data member.Member) error {
	sets, args := r.prepareUpdateArgs(data)
	if len(args) == 0 {
		return nil
	}

	args = append(args, id)
	sets = append(sets, "updated_at=CURRENT_TIMESTAMP")
	query := fmt.Sprintf("UPDATE members SET %s WHERE id=$%d RETURNING id", strings.Join(sets, ", "), len(args))

	err := r.GetDB().QueryRowContext(ctx, query, args...).Scan(&id)
	return HandleSQLError(err)
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

// Delete is inherited from BaseRepository

// GetByEmail retrieves a member by email
func (r *MemberRepository) GetByEmail(ctx context.Context, email string) (member.Member, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, books, created_at, updated_at, last_login_at
		FROM members
		WHERE email=$1`

	var m member.Member
	err := r.GetDB().GetContext(ctx, &m, query, email)
	return m, HandleSQLError(err)
}

// UpdateLastLogin updates last login time
func (r *MemberRepository) UpdateLastLogin(ctx context.Context, id string, loginTime time.Time) error {
	query := `
		UPDATE members
		SET last_login_at=$1, updated_at=CURRENT_TIMESTAMP
		WHERE id=$2
		RETURNING id`

	err := r.GetDB().QueryRowContext(ctx, query, loginTime, id).Scan(&id)
	return HandleSQLError(err)
}

// EmailExists checks if email exists
func (r *MemberRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM members WHERE email=$1)`

	var exists bool
	if err := r.GetDB().GetContext(ctx, &exists, query, email); err != nil {
		return false, fmt.Errorf("checking email existence: %w", err)
	}
	return exists, nil
}
