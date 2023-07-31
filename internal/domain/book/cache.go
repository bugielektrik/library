package book

import "context"

type Cache interface {
	Get(ctx context.Context, id string) (dest Entity, err error)
}
