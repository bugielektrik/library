package author

import (
	"context"
)

type Repository interface {
	List(ctx context.Context) ([]Entity, error)
	Add(ctx context.Context, data Entity) (string, error)
	Get(ctx context.Context, id string) (Entity, error)
	Update(ctx context.Context, id string, data Entity) error
	Delete(ctx context.Context, id string) error
}
