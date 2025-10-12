package app

import (
	"library-service/internal/books/domain/author"
	"library-service/internal/books/domain/book"
	bookrepo "library-service/internal/books/repository"
	bookmemory "library-service/internal/books/repository/memory"
	store "library-service/internal/infrastructure/store"
	memberdomain "library-service/internal/members/domain"
	memberrepo "library-service/internal/members/repository"
	membermemory "library-service/internal/members/repository/memory"
	paymentdomain "library-service/internal/payments/domain"
	"library-service/internal/payments/repository/postgres"
	reservationdomain "library-service/internal/reservations/domain"
	reservationrepo "library-service/internal/reservations/repository"
)

// RepositoryConfig function type for repository setup
type RepositoryConfig func(*Repositories) error

// Repositories holds all repository implementations
type Repositories struct {
	postgres *store.SQL

	Author        author.Repository
	Book          book.Repository
	Member        memberdomain.Repository
	Reservation   reservationdomain.Repository
	Payment       paymentdomain.Repository
	SavedCard     paymentdomain.SavedCardRepository
	CallbackRetry paymentdomain.CallbackRetryRepository
	Receipt       paymentdomain.ReceiptRepository
}

// NewRepositories creates a new repository container
func NewRepositories(configs ...RepositoryConfig) (*Repositories, error) {
	repos := &Repositories{}

	for _, cfg := range configs {
		if err := cfg(repos); err != nil {
			return nil, err
		}
	}

	return repos, nil
}

// Close closes all store connections
func (r *Repositories) Close() {
	if r.postgres != nil && r.postgres.Connection != nil {
		r.postgres.Connection.Close()
	}
}

// WithMemoryStore configures in-memory repositories
func WithMemoryStore() RepositoryConfig {
	return func(r *Repositories) error {
		r.Author = bookmemory.NewAuthorRepository()
		r.Book = bookmemory.NewBookRepository()
		r.Member = membermemory.NewMemberRepository()
		// Note: Reservation and Payment memory repositories not yet implemented
		r.Reservation = nil
		r.Payment = nil
		return nil
	}
}

// WithPostgresStore configures PostgreSQL repositories
func WithPostgresStore(dsn string) RepositoryConfig {
	return func(r *Repositories) error {
		db, err := store.NewSQL(dsn)
		if err != nil {
			return err
		}
		r.postgres = db

		if err := store.RunMigrations(dsn); err != nil {
			return err
		}

		r.Author = bookrepo.NewAuthorRepository(db.Connection)
		r.Book = bookrepo.NewBookRepository(db.Connection)
		r.Member = memberrepo.NewMemberRepository(db.Connection)
		r.Reservation = reservationrepo.NewReservationRepository(db.Connection)
		r.Payment = postgres.NewPaymentRepository(db.Connection)
		r.SavedCard = postgres.NewSavedCardRepository(db.Connection)
		r.CallbackRetry = postgres.NewCallbackRetryRepository(db.Connection)
		r.Receipt = postgres.NewReceiptRepository(db.Connection)

		return nil
	}
}
