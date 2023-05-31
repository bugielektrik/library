package library

import (
	"library/internal/domain/author"
	"library/internal/domain/book"
)

// Configuration is an alias for a function that will take in a pointer to a Service and modify it
type Configuration func(s *Service) error

// Service is an implementation of the Service
type Service struct {
	authorRepository author.Repository
	bookRepository   book.Repository
}

// New takes a variable amount of Configuration functions and returns a new Service
// Each Configuration will be called in the order they are passed in
func New(configs ...Configuration) (s *Service, err error) {
	// Create the service
	s = &Service{}

	// Apply all Configurations passed in
	for _, cfg := range configs {
		// Pass the service into the configuration function
		if err = cfg(s); err != nil {
			return
		}
	}
	return
}

// WithAuthorRepository applies a given author repository to the Service
func WithAuthorRepository(authorRepository author.Repository) Configuration {
	// return a function that matches the Configuration alias,
	// You need to return this so that the parent function can take in all the needed parameters
	return func(s *Service) error {
		s.authorRepository = authorRepository
		return nil
	}
}

// WithBookRepository applies a given book repository to the Service
func WithBookRepository(bookRepository book.Repository) Configuration {
	// Create the book repository, if we needed parameters, such as connection strings they could be inputted here
	return func(s *Service) error {
		s.bookRepository = bookRepository
		return nil
	}
}
