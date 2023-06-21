package book

import "context"

type Cache interface {
	GetByID(ctx context.Context, id string) (dest Entity, err error)
}
