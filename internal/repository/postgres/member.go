package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"library-service/internal/domain/member"
	"library-service/internal/repository/sqlc"
	"library-service/pkg/store"
)

type MemberRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewMemberRepository(db *pgxpool.Pool) *MemberRepository {
	return &MemberRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *MemberRepository) List(ctx context.Context) ([]member.Entity, error) {
	dbMembers, err := r.queries.ListMembers(ctx)
	if err != nil {
		return nil, err
	}

	members := make([]member.Entity, 0, len(dbMembers))
	for _, dbMember := range dbMembers {
		books, err := r.queries.GetMemberBooks(ctx, dbMember.ID)
		if err != nil {
			return nil, err
		}

		members = append(members, member.Entity{
			ID:       dbMember.ID,
			FullName: &dbMember.FullName,
			Books:    books,
		})
	}

	return members, nil
}

func (r *MemberRepository) Add(ctx context.Context, data member.Entity) (string, error) {
	id, err := r.queries.AddMember(ctx, sqlc.AddMemberParams{
		ID:       data.ID,
		FullName: *data.FullName,
	})
	if err != nil {
		return "", err
	}

	for _, bookID := range data.Books {
		err := r.queries.AddMemberBook(ctx, sqlc.AddMemberBookParams{
			BookID:   bookID,
			MemberID: id,
		})
		if err != nil {
			return "", err
		}
	}

	return id, nil
}

func (r *MemberRepository) Get(ctx context.Context, id string) (member.Entity, error) {
	dbMember, err := r.queries.GetMember(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return member.Entity{}, store.ErrorNotFound
		}
		return member.Entity{}, err
	}

	books, err := r.queries.GetMemberBooks(ctx, id)
	if err != nil {
		return member.Entity{}, err
	}

	return member.Entity{
		ID:       dbMember.ID,
		FullName: &dbMember.FullName,
		Books:    books,
	}, nil
}

func (r *MemberRepository) Update(ctx context.Context, id string, data member.Entity) error {
	if data.FullName != nil {
		_, err := r.queries.UpdateMember(ctx, sqlc.UpdateMemberParams{
			ID:       id,
			FullName: *data.FullName,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return store.ErrorNotFound
			}
			return err
		}
	}

	if len(data.Books) > 0 {
		err := r.queries.DeleteAllMemberBooks(ctx, id)
		if err != nil {
			return err
		}

		for _, bookID := range data.Books {
			err := r.queries.AddMemberBook(ctx, sqlc.AddMemberBookParams{
				BookID:   bookID,
				MemberID: id,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *MemberRepository) Delete(ctx context.Context, id string) error {
	err := r.queries.DeleteAllMemberBooks(ctx, id)
	if err != nil {
		return err
	}

	err = r.queries.DeleteMember(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.ErrorNotFound
		}
		return err
	}

	return nil
}
