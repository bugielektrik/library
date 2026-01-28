package postgres

import (
	"context"
	"fmt"
	"library-service/internal/repository/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, q *sqlc.Queries) error) error
}

type txManager struct {
	db *pgxpool.Pool
}

func NewTxManager(db *pgxpool.Pool) TxManager {
	return &txManager{db: db}
}

func (m *txManager) WithTx(ctx context.Context, fn func(ctx context.Context, q *sqlc.Queries) error) error {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	queries := sqlc.New(tx)

	if err := fn(ctx, queries); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("failed to rollback: %w (original: %v)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}
