package usecase

import (
	"library-service/internal/books/domain/author"
	"library-service/internal/books/domain/book"
	"library-service/internal/books/operations"
	authorops "library-service/internal/books/operations/author"
)

// BookUseCases contains all book-related use cases
type BookUseCases struct {
	CreateBook      *operations.CreateBookUseCase
	GetBook         *operations.GetBookUseCase
	ListBooks       *operations.ListBooksUseCase
	UpdateBook      *operations.UpdateBookUseCase
	DeleteBook      *operations.DeleteBookUseCase
	ListBookAuthors *operations.ListBookAuthorsUseCase
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
		CreateBook:      operations.NewCreateBookUseCase(bookRepo, bookCache, bookService),
		GetBook:         operations.NewGetBookUseCase(bookRepo, bookCache),
		ListBooks:       operations.NewListBooksUseCase(bookRepo),
		UpdateBook:      operations.NewUpdateBookUseCase(bookRepo, bookCache),
		DeleteBook:      operations.NewDeleteBookUseCase(bookRepo, bookCache),
		ListBookAuthors: operations.NewListBookAuthorsUseCase(bookRepo, authorRepo, authorCache),
	}
}

// newAuthorUseCases creates all author-related use cases
func newAuthorUseCases(authorRepo author.Repository) AuthorUseCases {
	return AuthorUseCases{
		ListAuthors: authorops.NewListAuthorsUseCase(authorRepo),
	}
}
