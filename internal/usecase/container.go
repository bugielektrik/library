/*
Package usecase provides the dependency injection container for all application use cases.

This is the central wiring point following Clean Architecture principles:
- Infrastructure services (JWT, Password, Gateway) are created in app.go
- Domain services (Book, Member, Reservation) are created in factories
- Use cases combine domain services with repositories

The container is now split into domain-specific factories for better organization:
- book_factory.go - Book and author use cases
- auth_factory.go - Authentication and member use cases
- payment_factory.go - Payment, saved card, and receipt use cases
- reservation_factory.go - Reservation use cases

For detailed workflow, see:
- .claude/development-workflows.md - Step-by-step feature guide
- .claude/adr/003-domain-services-vs-infrastructure.md - Service creation patterns
*/
package usecase

import (
	"library-service/internal/books/domain/author"
	"library-service/internal/books/domain/book"
	"library-service/internal/infrastructure/auth"
	memberdomain "library-service/internal/members/domain"
	paymentdomain "library-service/internal/payments/domain"
	reservationdomain "library-service/internal/reservations/domain"
)

// Container holds all application use cases organized by domain
type Container struct {
	// Book domain
	Book   BookUseCases
	Author AuthorUseCases

	// Member domain
	Auth         AuthUseCases
	Member       MemberUseCases
	Subscription SubscriptionUseCases

	// Reservation domain
	Reservation ReservationUseCases

	// Payment domain
	Payment   PaymentUseCases
	SavedCard SavedCardUseCases
	Receipt   ReceiptUseCases
}

// Repositories holds all repository interfaces
type Repositories struct {
	Book          book.Repository
	Author        author.Repository
	Member        memberdomain.Repository
	Reservation   reservationdomain.Repository
	Payment       paymentdomain.Repository
	SavedCard     paymentdomain.SavedCardRepository
	CallbackRetry paymentdomain.CallbackRetryRepository
	Receipt       paymentdomain.ReceiptRepository
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

// GatewayServices holds all gateway services
type GatewayServices struct {
	PaymentGateway interface {
		paymentdomain.Gateway
		paymentdomain.GatewayConfig
	}
}

// NewContainer creates a new use case container using domain-specific factories
func NewContainer(
	repos *Repositories,
	caches *Caches,
	authSvcs *AuthServices,
	gatewaySvcs *GatewayServices,
) *Container {
	// Create book-related use cases
	bookUseCases := newBookUseCases(
		repos.Book,
		repos.Author,
		caches.Book,
		caches.Author,
	)

	authorUseCases := newAuthorUseCases(repos.Author)

	// Create authentication and member use cases
	authUseCases := newAuthUseCases(
		repos.Member,
		authSvcs.JWTService,
		authSvcs.PasswordService,
	)

	memberUseCases := newMemberUseCases(repos.Member)
	subscriptionUseCases := newSubscriptionUseCases(repos.Member)

	// Create reservation use cases
	reservationUseCases := newReservationUseCases(
		repos.Reservation,
		repos.Member,
	)

	// Create payment-related use cases
	paymentRepos := PaymentRepositories{
		Payment:       repos.Payment,
		SavedCard:     repos.SavedCard,
		CallbackRetry: repos.CallbackRetry,
		Receipt:       repos.Receipt,
	}

	paymentUseCases, savedCardUseCases, receiptUseCases := newPaymentUseCases(
		paymentRepos,
		repos.Member,
		gatewaySvcs.PaymentGateway,
	)

	return &Container{
		Book:         bookUseCases,
		Author:       authorUseCases,
		Auth:         authUseCases,
		Member:       memberUseCases,
		Subscription: subscriptionUseCases,
		Reservation:  reservationUseCases,
		Payment:      paymentUseCases,
		SavedCard:    savedCardUseCases,
		Receipt:      receiptUseCases,
	}
}
