package service

import (
	"library-service/internal/cache"
	"library-service/internal/repository"
	"library-service/internal/service/library"
	"library-service/internal/service/subscription"
)

type Dependencies struct {
	Repositories *repository.Repositories
	Caches       *cache.Caches
}

type Configuration func(s *Services) error

type Services struct {
	dependencies Dependencies

	Library      *library.Service
	Subscription *subscription.Service
}

func New(dependencies Dependencies, configs ...Configuration) (s *Services, err error) {
	s = &Services{
		dependencies: dependencies,
	}

	for _, cfg := range configs {
		if err = cfg(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func WithLibraryService() Configuration {
	return func(s *Services) (err error) {
		s.Library = library.New(
			s.dependencies.Repositories.Author,
			s.dependencies.Repositories.Book,
			s.dependencies.Caches.Author,
			s.dependencies.Caches.Book,
		)
		return err
	}
}

func WithSubscriptionService() Configuration {
	return func(s *Services) (err error) {
		s.Subscription = subscription.New(
			s.dependencies.Repositories.Member,
			s.Library,
		)
		return err
	}
}
