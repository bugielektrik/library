package subscription

import (
	"library-service/internal/domain/member"
	"library-service/internal/service/library"
)

// Configuration is an alias for a function that will take in a pointer to a Service and modify it
type Configuration func(s *Service) error

// Service is an implementation of the Service
type Service struct {
	memberRepository member.Repository
	libraryService   *library.Service
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

// WithMemberRepository applies a given member repository to the Service
func WithMemberRepository(memberRepository member.Repository) Configuration {
	// return a function that matches the Configuration alias,
	// You need to return this so that the parent function can take in all the needed parameters
	return func(s *Service) error {
		s.memberRepository = memberRepository
		return nil
	}
}

// WithLibraryService applies a given library service to the Service
func WithLibraryService(libraryService *library.Service) Configuration {
	// Create the library service, if we needed parameters, such as connection strings they could be inputted here
	return func(s *Service) error {
		s.libraryService = libraryService
		return nil
	}
}
