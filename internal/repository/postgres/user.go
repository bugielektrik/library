package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"library-service/internal/domain/user"
	"library-service/internal/repository/sqlc"
	"library-service/pkg/store"
)

type UserRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *UserRepository) Create(ctx context.Context, data user.Entity) (string, error) {
	fullName := pgtype.Text{Valid: false}
	if data.FullName != nil {
		fullName = pgtype.Text{String: *data.FullName, Valid: true}
	}

	id, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           data.ID,
		Email:        data.Email,
		PasswordHash: data.PasswordHash,
		FullName:     fullName,
		CreatedAt:    pgtype.Timestamp{Time: *data.CreatedAt, Valid: true},
		UpdatedAt:    pgtype.Timestamp{Time: *data.UpdatedAt, Valid: true},
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (user.Entity, error) {
	dbUser, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.Entity{}, store.ErrorNotFound
		}
		return user.Entity{}, err
	}

	var fullName *string
	if dbUser.FullName.Valid {
		fullName = &dbUser.FullName.String
	}

	return user.Entity{
		ID:           dbUser.ID,
		Email:        dbUser.Email,
		PasswordHash: dbUser.PasswordHash,
		FullName:     fullName,
		CreatedAt:    &dbUser.CreatedAt.Time,
		UpdatedAt:    &dbUser.UpdatedAt.Time,
	}, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (user.Entity, error) {
	dbUser, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.Entity{}, store.ErrorNotFound
		}
		return user.Entity{}, err
	}

	var fullName *string
	if dbUser.FullName.Valid {
		fullName = &dbUser.FullName.String
	}

	return user.Entity{
		ID:           dbUser.ID,
		Email:        dbUser.Email,
		PasswordHash: dbUser.PasswordHash,
		FullName:     fullName,
		CreatedAt:    &dbUser.CreatedAt.Time,
		UpdatedAt:    &dbUser.UpdatedAt.Time,
	}, nil
}
