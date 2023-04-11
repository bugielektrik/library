// Package library holds all the services that connects stores into a business flow related to library domains
package library

import (
	"github.com/jmoiron/sqlx"

	"library/internal/repository"
	"library/internal/repository/postgres"
)

type Configuration func(s *Service)

type Service struct {
	authors repository.AuthorRepository
	books   repository.BookRepository
	members repository.MemberRepository
}

func NewService(configs ...Configuration) *Service {
	service := &Service{}
	for _, config := range configs {
		config(service)
	}
	return service
}

func WithPostgresRepository(db *sqlx.DB) Configuration {
	return func(s *Service) {
		s.authors = postgres.NewAuthorRepository(db)
		s.books = postgres.NewBookRepository(db)
		s.members = postgres.NewMemberRepository(db)
		return
	}
}
