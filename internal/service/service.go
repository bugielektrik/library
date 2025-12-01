package service

import (
	"errors"
	"library-service/config"
	"library-service/internal/cache"
	"library-service/internal/repository"
	"library-service/internal/service/auth"
	"library-service/internal/service/author"
	"library-service/internal/service/book"
	"library-service/internal/service/interfaces"
	"library-service/internal/service/member"
)

type Dependencies struct {
	Repositories *repository.Repositories
	Caches       *cache.Caches
	Configs      *config.Configs
}

type Configuration func(s *Services) error

type Services struct {
	dependencies Dependencies
	Author       interfaces.AuthorService
	Book         interfaces.BookService
	Member       interfaces.MemberService
	Auth         interfaces.AuthService
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
		s.Author = author.NewAuthorService(
			s.dependencies.Repositories.Author,
			s.dependencies.Caches.Author,
		)
		s.Book = book.NewBookService(
			s.dependencies.Repositories.Book,
			s.dependencies.Caches.Book,
		)
		return nil
	}
}

func WithSubscriptionService() Configuration {
	return func(s *Services) (err error) {
		if s.Book == nil {
			return errors.New("book service is required for member service")
		}

		s.Member = member.NewMemberService(
			s.dependencies.Repositories.Member,
			s.Book,
		)
		return nil
	}
}

func WithAuthService() Configuration {
	return func(s *Services) (err error) {
		s.Auth = auth.NewAuthService(
			s.dependencies.Repositories.User,
			s.dependencies.Configs.JWT,
		)
		return nil
	}
}
