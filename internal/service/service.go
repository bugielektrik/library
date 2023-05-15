package service

import "library/internal/repository"

type Dependencies struct {
	AuthorRepository repository.AuthorRepository
	BookRepository   repository.BookRepository
	MemberRepository repository.MemberRepository
}

type Service struct {
	Author AuthorService
	Book   BookService
	Member MemberService
}

func New(d Dependencies) Service {
	return Service{
		Author: NewAuthorService(d.AuthorRepository),
		Book:   NewBookService(d.BookRepository),
		Member: NewMemberService(d.MemberRepository),
	}
}
