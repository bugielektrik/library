package user

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, data Entity) (string, error)
	GetByEmail(ctx context.Context, email string) (Entity, error)
	GetByID(ctx context.Context, id string) (Entity, error)
}
