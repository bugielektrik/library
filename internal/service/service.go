package service

import (
	"errors"
	"library-service/internal/cache"
	"library-service/internal/repository"
	"library-service/internal/service/interfaces"
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
	Author       interfaces.AuthorService
	Book         interfaces.BookService
	Member       interfaces.MemberService
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
		s.Author = library.NewAuthorService(
			s.dependencies.Repositories.Author,
			s.dependencies.Caches.Author,
		)
		s.Book = library.NewBookService(
			s.dependencies.Repositories.Book,
			s.dependencies.Caches.Book,
		)
		return nil
	}
}

func WithSubscriptionService() Configuration {
	return func(s *Services) (err error) {
		if s.Book == nil {
			return errors.New("book service is required for subscription service")
		}

		s.Member = subscription.NewMemberService(
			s.dependencies.Repositories.Member,
			s.Book,
		)
		return nil
	}
}
