package repository

import (
	"library-service/internal/adapters/repository/memory"
	"library-service/internal/adapters/repository/mongo"
	"library-service/internal/adapters/repository/postgres"
	"library-service/internal/domain/author"
	"library-service/internal/domain/book"
	"library-service/internal/domain/member"
	"library-service/internal/domain/payment"
	"library-service/internal/domain/reservation"
	store "library-service/internal/infrastructure/store"
)

// Configuration function type for repository setup
type Configuration func(*Repositories) error

// Repositories holds all repository implementations
type Repositories struct {
	mongo    *store.Mongo
	postgres *store.SQL

	Author      author.Repository
	Book        book.Repository
	Member      member.Repository
	Reservation reservation.Repository
	Payment     payment.Repository
	SavedCard   payment.SavedCardRepository
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

		r.Author = postgres.NewAuthorRepository(db.Connection)
		r.Book = postgres.NewBookRepository(db.Connection)
		r.Member = postgres.NewMemberRepository(db.Connection)
		r.Reservation = postgres.NewReservationRepository(db.Connection)
		r.Payment = postgres.NewPaymentRepository(db.Connection)
		r.SavedCard = postgres.NewSavedCardRepository(db.Connection)

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
