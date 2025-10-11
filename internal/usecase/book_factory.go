package usecase

import (
	"library-service/internal/domain/author"
	"library-service/internal/domain/book"
	"library-service/internal/usecase/authorops"
	"library-service/internal/usecase/bookops"
)

// BookUseCases contains all book-related use cases
type BookUseCases struct {
	CreateBook      *bookops.CreateBookUseCase
	GetBook         *bookops.GetBookUseCase
	ListBooks       *bookops.ListBooksUseCase
	UpdateBook      *bookops.UpdateBookUseCase
	DeleteBook      *bookops.DeleteBookUseCase
	ListBookAuthors *bookops.ListBookAuthorsUseCase
}

// AuthorUseCases contains all author-related use cases
type AuthorUseCases struct {
	ListAuthors *authorops.ListAuthorsUseCase
}

// newBookUseCases creates all book-related use cases
func newBookUseCases(
	bookRepo book.Repository,
	authorRepo author.Repository,
	bookCache book.Cache,
	authorCache author.Cache,
) BookUseCases {
	// Create domain service
	bookService := book.NewService()

	return BookUseCases{
		CreateBook:      bookops.NewCreateBookUseCase(bookRepo, bookCache, bookService),
		GetBook:         bookops.NewGetBookUseCase(bookRepo, bookCache),
		ListBooks:       bookops.NewListBooksUseCase(bookRepo),
		UpdateBook:      bookops.NewUpdateBookUseCase(bookRepo, bookCache),
		DeleteBook:      bookops.NewDeleteBookUseCase(bookRepo, bookCache),
		ListBookAuthors: bookops.NewListBookAuthorsUseCase(bookRepo, authorRepo, authorCache),
	}
}

// newAuthorUseCases creates all author-related use cases
func newAuthorUseCases(authorRepo author.Repository) AuthorUseCases {
	return AuthorUseCases{
		ListAuthors: authorops.NewListAuthorsUseCase(authorRepo),
	}
}
