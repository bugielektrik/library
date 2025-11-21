package interfaces

import (
	"context"
	"library-service/internal/domain/author"
)

type AuthorService interface {
	ListAuthors(ctx context.Context) ([]author.Response, error)
	AddAuthor(ctx context.Context, req author.Request) (author.Response, error)
	GetAuthor(ctx context.Context, id string) (author.Response, error)
	UpdateAuthor(ctx context.Context, id string, req author.Request) error
	DeleteAuthor(ctx context.Context, id string) error
}
