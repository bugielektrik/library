/*
Package app provides the dependency injection container for all application use cases.

This is the central wiring point following Clean Architecture principles:
- Infrastructure services (JWT, Password, Gateway) are created in app.go
- Domain services (Book, Member, Reservation) are created in factories
- Use cases combine domain services with repositories

The container is organized by domain with factory functions for each bounded context:
- Book and Author use cases
- Authentication and Member use cases
- Payment, SavedCard, and Receipt use cases
- Reservation use cases

For detailed workflow, see:
- .claude/guides/common-tasks.md - Step-by-step feature guide
- .claude/adr/003-domain-services-vs-infrastructure.md - Service creation patterns
*/
package container

import (
	"library-service/internal/books/domain/author"
	"library-service/internal/books/domain/book"
	bookservice "library-service/internal/books/service"
	authorservice "library-service/internal/books/service/author"
	infraauth "library-service/internal/infrastructure/auth"
	memberdomain "library-service/internal/members/domain"
	memberauth "library-service/internal/members/service/auth"
	"library-service/internal/members/service/profile"
	"library-service/internal/members/service/subscription"
	paymentdomain "library-service/internal/payments/domain"
	paymentservice "library-service/internal/payments/service/payment"
	receiptservice "library-service/internal/payments/service/receipt"
	savedcardservice "library-service/internal/payments/service/savedcard"
	reservationdomain "library-service/internal/reservations/domain"
	reservationservice "library-service/internal/reservations/service"
)

// ================================================================================
// Container and Dependencies
// ================================================================================

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
	JWTService      *infraauth.JWTService
	PasswordService *infraauth.PasswordService
}

// GatewayServices holds all gateway services
type GatewayServices struct {
	PaymentGateway interface {
		paymentdomain.Gateway
		paymentdomain.GatewayConfig
	}
}

// Validator defines the validation interface used by use cases
type Validator interface {
	Validate(i interface{}) error
}

// ================================================================================
// Use Case Groups by Domain
// ================================================================================

// BookUseCases contains all book-related use cases
type BookUseCases struct {
	CreateBook      *bookservice.CreateBookUseCase
	GetBook         *bookservice.GetBookUseCase
	ListBooks       *bookservice.ListBooksUseCase
	UpdateBook      *bookservice.UpdateBookUseCase
	DeleteBook      *bookservice.DeleteBookUseCase
	ListBookAuthors *bookservice.ListBookAuthorsUseCase
}

// AuthorUseCases contains all author-related use cases
type AuthorUseCases struct {
	ListAuthors *authorservice.ListAuthorsUseCase
}

// AuthUseCases contains all authentication-related use cases
type AuthUseCases struct {
	RegisterMember *memberauth.RegisterUseCase
	LoginMember    *memberauth.LoginUseCase
	RefreshToken   *memberauth.RefreshTokenUseCase
	ValidateToken  *memberauth.ValidateTokenUseCase
}

// MemberUseCases contains all member-related use cases
type MemberUseCases struct {
	ListMembers      *profile.ListMembersUseCase
	GetMemberProfile *profile.GetMemberProfileUseCase
}

// SubscriptionUseCases contains subscription-related use cases
type SubscriptionUseCases struct {
	SubscribeMember *subscription.SubscribeMemberUseCase
}

// ReservationUseCases contains all reservation-related use cases
type ReservationUseCases struct {
	CreateReservation      *reservationservice.CreateReservationUseCase
	CancelReservation      *reservationservice.CancelReservationUseCase
	GetReservation         *reservationservice.GetReservationUseCase
	ListMemberReservations *reservationservice.ListMemberReservationsUseCase
}

// PaymentUseCases contains all payment-related use cases
type PaymentUseCases struct {
	InitiatePayment        *paymentservice.InitiatePaymentUseCase
	VerifyPayment          *paymentservice.VerifyPaymentUseCase
	HandleCallback         *paymentservice.HandleCallbackUseCase
	ListMemberPayments     *paymentservice.ListMemberPaymentsUseCase
	CancelPayment          *paymentservice.CancelPaymentUseCase
	RefundPayment          *paymentservice.RefundPaymentUseCase
	PayWithSavedCard       *savedcardservice.PayWithSavedCardUseCase
	ExpirePayments         *paymentservice.ExpirePaymentsUseCase
	ProcessCallbackRetries *paymentservice.ProcessCallbackRetriesUseCase
}

// SavedCardUseCases contains all saved card-related use cases
type SavedCardUseCases struct {
	SaveCard        *paymentservice.SaveCardUseCase
	ListSavedCards  *savedcardservice.ListSavedCardsUseCase
	DeleteSavedCard *savedcardservice.DeleteSavedCardUseCase
	SetDefaultCard  *paymentservice.SetDefaultCardUseCase
}

// ReceiptUseCases contains all receipt-related use cases
type ReceiptUseCases struct {
	GenerateReceipt *receiptservice.GenerateReceiptUseCase
	GetReceipt      *receiptservice.GetReceiptUseCase
	ListReceipts    *receiptservice.ListReceiptsUseCase
}

// PaymentRepositories contains payment-related repositories
type PaymentRepositories struct {
	Payment       paymentdomain.Repository
	SavedCard     paymentdomain.SavedCardRepository
	CallbackRetry paymentdomain.CallbackRetryRepository
	Receipt       paymentdomain.ReceiptRepository
}

// ================================================================================
// Main Container Constructor
// ================================================================================

// NewContainer creates a new use case container using domain-specific factories
func NewContainer(
	repos *Repositories,
	caches *Caches,
	authSvcs *AuthServices,
	gatewaySvcs *GatewayServices,
	validator Validator,
) *Container {
	// Create book-related use cases
	bookUseCases := newBookUseCases(
		repos.Book,
		repos.Author,
		caches.Book,
		caches.Author,
		validator,
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
		validator,
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
