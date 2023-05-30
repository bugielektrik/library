package service

import (
	"library/internal/domain/author"
	"library/internal/domain/book"
	"library/internal/domain/member"
	"library/internal/service/library"
	"library/internal/service/subscription"
)

type Dependencies struct {
	AuthorRepository author.Repository
	BookRepository   book.Repository
	MemberRepository member.Repository
}

// Configuration is an alias for a function that will take in a pointer to a Service and modify it
type Configuration func(s *Service) error

// Service is an implementation of the Service
type Service struct {
	dependencies Dependencies

	Library      library.Service
	Subscription subscription.Service
}

// New takes a variable amount of Configuration functions and returns a new Service
// Each Configuration will be called in the order they are passed in
func New(d Dependencies, configs ...Configuration) (s *Service, err error) {
	// Create the service
	s = &Service{
		dependencies: d,
	}

	// Apply all Configurations passed in
	for _, cfg := range configs {
		// Pass the service into the configuration function
		if err = cfg(s); err != nil {
			return
		}
	}
	return
}

// WithLibraryService applies a library service to the Service
func WithLibraryService() Configuration {
	return func(s *Service) (err error) {
		// Create the library service, if we needed parameters, such as connection strings they could be inputted here
		s.Library = library.New(
			s.dependencies.AuthorRepository,
			s.dependencies.BookRepository)
		return
	}
}

// WithSubscriptionService applies a subscription service to the Service
func WithSubscriptionService() Configuration {
	return func(s *Service) (err error) {
		// Create the subscription service, if we needed parameters, such as connection strings they could be inputted here
		s.Subscription = subscription.New(
			s.Library,
			s.dependencies.MemberRepository)
		return
	}
}
