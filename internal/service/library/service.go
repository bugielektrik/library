package library

import (
	"library/internal/domain/author"
	"library/internal/domain/book"
)

type Service struct {
	authorRepository author.Repository
	bookRepository   book.Repository
}

func New(a author.Repository, b book.Repository) Service {
	return Service{
		authorRepository: a,
		bookRepository:   b,
	}
}
