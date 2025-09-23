package service

import (
	"library-service/internal/cache"
	"library-service/internal/repository"
	"library-service/internal/service/library"
	"library-service/internal/service/subscription"
)

// Dependencies holds the dependencies required for creating services
type Dependencies struct {
	Repositories *repository.Repositories
	Caches       *cache.Caches
}

// Configuration is an alias for a function that will take in a pointer to Services and modify it
type Configuration func(s *Services) error

// Services holds all business logic services
type Services struct {
	dependencies Dependencies

	Library      *library.Service
	Subscription *subscription.Service
}

// New takes a variable amount of Configuration functions and returns a new Services instance
// Each Configuration will be called in the order they are passed in
func New(dependencies Dependencies, configs ...Configuration) (s *Services, err error) {
	// Create the services container
	s = &Services{
		dependencies: dependencies,
	}

	// Apply all configurations passed in
	for _, cfg := range configs {
		// Pass the services into the configuration function
		if err = cfg(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// WithLibraryService configures the library service with repositories and caches
func WithLibraryService() Configuration {
	return func(s *Services) (err error) {
		s.Library, err = library.New(
			library.WithAuthorRepository(s.dependencies.Repositories.Author),
			library.WithBookRepository(s.dependencies.Repositories.Book),
			library.WithAuthorCache(s.dependencies.Caches.Author),
			library.WithBookCache(s.dependencies.Caches.Book),
		)
		return err
	}
}

// WithSubscriptionService configures the subscription service with dependencies
// Note: This creates a circular dependency that should be resolved through dependency injection
func WithSubscriptionService() Configuration {
	return func(s *Services) (err error) {
		s.Subscription, err = subscription.New(
			subscription.WithMemberRepository(s.dependencies.Repositories.Member),
			subscription.WithLibraryService(s.Library),
		)
		return err
	}
}
