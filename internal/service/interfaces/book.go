package interfaces

import (
	"context"
	"library-service/internal/domain/author"
	"library-service/internal/domain/book"
)

//go:generate mockery --name=BookService --output=../../../mocks --outpkg=mocks --filename=BookService.go
type BookService interface {
	ListBooks(ctx context.Context) ([]book.Response, error)
	CreateBook(ctx context.Context, req book.Request) (book.Response, error)
	GetBook(ctx context.Context, id string) (book.Response, error)
	UpdateBook(ctx context.Context, id string, req book.Request) error
	DeleteBook(ctx context.Context, id string) error
	ListBookAuthors(ctx context.Context, id string) ([]author.Response, error)
}
