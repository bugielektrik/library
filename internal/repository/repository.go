package repository

import (
	"library-service/internal/domain/author"
	"library-service/internal/domain/book"
	"library-service/internal/domain/member"
	"library-service/internal/domain/user"
	"library-service/internal/repository/memory"
	"library-service/internal/repository/mongo"
	"library-service/internal/repository/postgres"
	"library-service/pkg/store"
)

type Configuration func(r *Repositories) error

type Repositories struct {
	mongo      *store.Mongo
	postgres   *store.SQL
	clickhouse *store.ClickHouse

	Author   author.Repository
	Book     book.Repository
	Member   member.Repository
	User     user.Repository
	TxManger postgres.TxManager
}

func New(configs ...Configuration) (s *Repositories, err error) {
	s = &Repositories{}

	for _, cfg := range configs {
		if err = cfg(s); err != nil {
			return
		}
	}

	return
}

func (r *Repositories) Close() {
	if r.postgres != nil && r.postgres.Connection != nil {
		r.postgres.Connection.Close()
	}

	if r.clickhouse != nil && r.clickhouse.Connection != nil {
		r.clickhouse.Close()
	}

	if r.mongo != nil && r.mongo.Connection != nil {
		r.mongo.Connection.Disconnect(nil)
	}
}

func WithMemoryStore() Configuration {
	return func(s *Repositories) (err error) {
		s.Author = memory.NewAuthorRepository()
		s.Book = memory.NewBookRepository()
		s.Member = memory.NewMemberRepository()

		return
	}
}

func WithMongoStore(uri, name string) Configuration {
	return func(s *Repositories) (err error) {
		s.mongo, err = store.NewMongo(uri)
		if err != nil {
			return
		}
		database := s.mongo.Connection.Database(name)

		s.Author = mongo.NewAuthorRepository(database)
		s.Book = mongo.NewBookRepository(database)
		s.Member = mongo.NewMemberRepository(database)

		return
	}
}

func WithPostgresStore(dataSourceName string) Configuration {
	return func(s *Repositories) (err error) {
		s.postgres, err = store.NewSQL(dataSourceName)
		if err != nil {
			return
		}

		if err = store.RunMigrations(dataSourceName); err != nil {
			return
		}

		s.Author = postgres.NewAuthorRepository(s.postgres.Connection)
		s.Book = postgres.NewBookRepository(s.postgres.Connection)
		s.Member = postgres.NewMemberRepository(s.postgres.Connection)
		s.User = postgres.NewUserRepository(s.postgres.Connection)
		s.TxManger = postgres.NewTxManager(s.postgres.Connection)
		return
	}
}

func WithClickHouseStore() Configuration {
	return func(s *Repositories) (err error) {
		s.clickhouse, err = store.New()
		if err != nil {
			return
		}

		return
	}
}
