package service

import "library/internal/repository"

type Dependencies struct {
	AuthorRepository repository.AuthorRepository
	BookRepository   repository.BookRepository
	MemberRepository repository.MemberRepository
}

// Service is an implementation of the Service
type Service struct {
	Author AuthorService
	Book   BookService
	Member MemberService
}

// New creates a new instance of the Service struct
func New(d Dependencies) Service {
	return Service{
		Author: NewAuthorService(d.AuthorRepository),
		Book:   NewBookService(d.BookRepository),
		Member: NewMemberService(d.MemberRepository),
	}
}
