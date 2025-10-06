package usecase

import (
	"library-service/internal/domain/author"
	"library-service/internal/domain/book"
	"library-service/internal/domain/member"
	"library-service/internal/domain/reservation"
	"library-service/internal/infrastructure/auth"
	"library-service/internal/usecase/authops"
	"library-service/internal/usecase/bookops"
	"library-service/internal/usecase/reservationops"
	"library-service/internal/usecase/subops"
)

// Container holds all application usecases
type Container struct {
	// Book usecases
	CreateBook      *bookops.CreateBookUseCase
	GetBook         *bookops.GetBookUseCase
	ListBooks       *bookops.ListBooksUseCase
	UpdateBook      *bookops.UpdateBookUseCase
	DeleteBook      *bookops.DeleteBookUseCase
	ListBookAuthors *bookops.ListBookAuthorsUseCase

	// Subscription usecases
	SubscribeMember *subops.SubscribeMemberUseCase

	// Auth usecases
	RegisterMember   *authops.RegisterUseCase
	LoginMember      *authops.LoginUseCase
	RefreshToken     *authops.RefreshTokenUseCase
	ValidateToken    *authops.ValidateTokenUseCase

	// Reservation usecases
	CreateReservation      *reservationops.CreateReservationUseCase
	CancelReservation      *reservationops.CancelReservationUseCase
	GetReservation         *reservationops.GetReservationUseCase
	ListMemberReservations *reservationops.ListMemberReservationsUseCase
}

// Repositories holds all repository interfaces
type Repositories struct {
	Book        book.Repository
	Author      author.Repository
	Member      member.Repository
	Reservation reservation.Repository
}

// Caches holds all cache interfaces
type Caches struct {
	Book   book.Cache
	Author author.Cache
}

// AuthServices holds all authentication services
type AuthServices struct {
	JWTService      *auth.JWTService
	PasswordService *auth.PasswordService
}

// NewContainer creates a new usecase container with all dependencies injected
func NewContainer(repos *Repositories, caches *Caches, authSvcs *AuthServices) *Container {
	// Create domain services
	bookService := book.NewService()
	memberService := member.NewService()
	reservationService := reservation.NewService()

	return &Container{
		// Book usecases
		CreateBook:      bookops.NewCreateBookUseCase(repos.Book, caches.Book, bookService),
		GetBook:         bookops.NewGetBookUseCase(repos.Book, caches.Book),
		ListBooks:       bookops.NewListBooksUseCase(repos.Book),
		UpdateBook:      bookops.NewUpdateBookUseCase(repos.Book, caches.Book),
		DeleteBook:      bookops.NewDeleteBookUseCase(repos.Book, caches.Book),
		ListBookAuthors: bookops.NewListBookAuthorsUseCase(repos.Book, repos.Author, caches.Author),

		// Subscription usecases
		SubscribeMember: subops.NewSubscribeMemberUseCase(repos.Member, memberService),

		// Auth usecases
		RegisterMember:   authops.NewRegisterUseCase(repos.Member, authSvcs.PasswordService, authSvcs.JWTService, memberService),
		LoginMember:      authops.NewLoginUseCase(repos.Member, authSvcs.PasswordService, authSvcs.JWTService),
		RefreshToken:     authops.NewRefreshTokenUseCase(repos.Member, authSvcs.JWTService),
		ValidateToken:    authops.NewValidateTokenUseCase(repos.Member, authSvcs.JWTService),

		// Reservation usecases
		CreateReservation:      reservationops.NewCreateReservationUseCase(repos.Reservation, repos.Member, reservationService),
		CancelReservation:      reservationops.NewCancelReservationUseCase(repos.Reservation, reservationService),
		GetReservation:         reservationops.NewGetReservationUseCase(repos.Reservation),
		ListMemberReservations: reservationops.NewListMemberReservationsUseCase(repos.Reservation),
	}
}
