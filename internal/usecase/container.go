package usecase

import (
	bookuc "library-service/internal/usecase/book"
	subscriptionuc "library-service/internal/usecase/subscription"

	"library-service/internal/domain/author"
	"library-service/internal/domain/book"
	"library-service/internal/domain/member"
)

// Container holds all application usecases
type Container struct {
	// Book usecases
	CreateBook      *bookuc.CreateBookUseCase
	GetBook         *bookuc.GetBookUseCase
	ListBooks       *bookuc.ListBooksUseCase
	UpdateBook      *bookuc.UpdateBookUseCase
	DeleteBook      *bookuc.DeleteBookUseCase
	ListBookAuthors *bookuc.ListBookAuthorsUseCase

	// Subscription usecases
	SubscribeMember *subscriptionuc.SubscribeMemberUseCase
}

// Repositories holds all repository interfaces
type Repositories struct {
	Book   book.Repository
	Author author.Repository
	Member member.Repository
}

// Caches holds all cache interfaces
type Caches struct {
	Book   book.Cache
	Author author.Cache
}

// NewContainer creates a new usecase container with all dependencies injected
func NewContainer(repos *Repositories, caches *Caches) *Container {
	// Create domain services
	bookService := book.NewService()
	memberService := member.NewService()

	return &Container{
		// Book usecases
		CreateBook:      bookuc.NewCreateBookUseCase(repos.Book, caches.Book, bookService),
		GetBook:         bookuc.NewGetBookUseCase(repos.Book, caches.Book),
		ListBooks:       bookuc.NewListBooksUseCase(repos.Book),
		UpdateBook:      bookuc.NewUpdateBookUseCase(repos.Book, caches.Book),
		DeleteBook:      bookuc.NewDeleteBookUseCase(repos.Book, caches.Book),
		ListBookAuthors: bookuc.NewListBookAuthorsUseCase(repos.Book, repos.Author, caches.Author),

		// Subscription usecases
		SubscribeMember: subscriptionuc.NewSubscribeMemberUseCase(repos.Member, memberService),
	}
}
