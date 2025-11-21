package book

import "context"

type Cache interface {
	Get(ctx context.Context, id string) (Entity, error)
	Set(ctx context.Context, id string, entity Entity) error
}
