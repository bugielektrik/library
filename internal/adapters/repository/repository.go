package repository

import (
	"library-service/internal/adapters/repository/memory"
	"library-service/internal/adapters/repository/mongo"
	"library-service/internal/books/domain/author"
	"library-service/internal/books/domain/book"
	bookrepo "library-service/internal/books/repository"
	store "library-service/internal/infrastructure/store"
	memberdomain "library-service/internal/members/domain"
	memberrepo "library-service/internal/members/repository"
	paymentdomain "library-service/internal/payments/domain"
	paymentrepo "library-service/internal/payments/repository"
	reservationdomain "library-service/internal/reservations/domain"
	reservationrepo "library-service/internal/reservations/repository"
)

// Configuration function type for repository setup
type Configuration func(*Repositories) error

// Repositories holds all repository implementations
type Repositories struct {
	mongo    *store.Mongo
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
func NewRepositories(configs ...Configuration) (*Repositories, error) {
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
	if r.mongo != nil && r.mongo.Connection != nil {
		r.mongo.Connection.Disconnect(nil)
	}
}

// WithMemoryStore configures in-memory repositories
func WithMemoryStore() Configuration {
	return func(r *Repositories) error {
		r.Author = memory.NewAuthorRepository()
		r.Book = memory.NewBookRepository()
		r.Member = memory.NewMemberRepository()
		// Note: Reservation and Payment memory repositories not yet implemented
		r.Reservation = nil
		r.Payment = nil
		return nil
	}
}

// WithPostgresStore configures PostgreSQL repositories
func WithPostgresStore(dsn string) Configuration {
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
		r.Payment = paymentrepo.NewPaymentRepository(db.Connection)
		r.SavedCard = paymentrepo.NewSavedCardRepository(db.Connection)
		r.CallbackRetry = paymentrepo.NewCallbackRetryRepository(db.Connection)
		r.Receipt = paymentrepo.NewReceiptRepository(db.Connection)

		return nil
	}
}

// WithMongoStore configures MongoDB repositories
func WithMongoStore(uri, dbName string) Configuration {
	return func(r *Repositories) error {
		db, err := store.NewMongo(uri)
		if err != nil {
			return err
		}
		r.mongo = db
		database := db.Connection.Database(dbName)

		r.Author = mongo.NewAuthorRepository(database)
		r.Book = mongo.NewBookRepository(database)
		r.Member = mongo.NewMemberRepository(database)

		return nil
	}
}
