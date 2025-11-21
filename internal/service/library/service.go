package library

import (
	"context"
	"library-service/internal/domain/author"
	"library-service/internal/domain/book"
)

type Service struct {
	authorRepository author.Repository
	bookRepository   book.Repository
	authorCache      author.Cache
	bookCache        book.Cache
}

func New(authorRepository author.Repository, bookRepository book.Repository, authorCache author.Cache, bookCache book.Cache) *Service {
	return &Service{
		authorRepository: authorRepository,
		bookRepository:   bookRepository,
		authorCache:      authorCache,
		bookCache:        bookCache,
	}
}
