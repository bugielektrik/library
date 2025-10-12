package container

import (
	"library-service/internal/books/domain/author"
	"library-service/internal/books/domain/book"
	bookservice "library-service/internal/books/service"
	authorservice "library-service/internal/books/service/author"
)

// ================================================================================
// Factory Functions - Book Domain
// ================================================================================

// newBookUseCases creates all book-related use cases
func newBookUseCases(
	bookRepo book.Repository,
	authorRepo author.Repository,
	bookCache book.Cache,
	authorCache author.Cache,
	validator bookservice.Validator,
) BookUseCases {
	// Create domain service
	bookService := book.NewService()

	return BookUseCases{
		CreateBook:      bookservice.NewCreateBookUseCase(bookRepo, bookCache, bookService, validator),
		GetBook:         bookservice.NewGetBookUseCase(bookRepo, bookCache),
		ListBooks:       bookservice.NewListBooksUseCase(bookRepo),
		UpdateBook:      bookservice.NewUpdateBookUseCase(bookRepo, bookCache),
		DeleteBook:      bookservice.NewDeleteBookUseCase(bookRepo, bookCache),
		ListBookAuthors: bookservice.NewListBookAuthorsUseCase(bookRepo, authorRepo, authorCache),
	}
}

// newAuthorUseCases creates all author-related use cases
func newAuthorUseCases(authorRepo author.Repository) AuthorUseCases {
	return AuthorUseCases{
		ListAuthors: authorservice.NewListAuthorsUseCase(authorRepo),
	}
}
