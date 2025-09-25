package library

import (
	"library-service/internal/domain/author"
	"library-service/internal/domain/book"
)

// Service aggregates the repositories and caches.
type Service struct {
	authorRepository author.Repository
	bookRepository   book.Repository
	authorCache      author.Cache
	bookCache        book.Cache
}

// New creates a new instance of the Service with the provided repositories and caches.
func New(authorRepository author.Repository, bookRepository book.Repository, authorCache author.Cache, bookCache book.Cache) *Service {
	return &Service{
		authorRepository: authorRepository,
		bookRepository:   bookRepository,
		authorCache:      authorCache,
		bookCache:        bookCache,
	}
}
