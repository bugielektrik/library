package member

import (
	"context"
)

type Repository interface {
	SelectRows(ctx context.Context) (dest []Entity, err error)
	CreateRow(ctx context.Context, data Entity) (id string, err error)
	GetRow(ctx context.Context, id string) (dest Entity, err error)
	UpdateRow(ctx context.Context, id string, data Entity) (err error)
	DeleteRow(ctx context.Context, id string) (err error)
}
