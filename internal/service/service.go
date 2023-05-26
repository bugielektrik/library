package service

import (
	"library/internal/repository"
)

type Dependencies struct {
	AuthorRepository repository.Author
	BookRepository   repository.Book
	MemberRepository repository.Member
}

type Service struct {
	Author Author
	Book   Book
	Member Member
}

func New(d Dependencies) Service {
	authorService := NewAuthorService(d.AuthorRepository)
	bookService := NewBookService(d.BookRepository, authorService)
	memberService := NewMemberService(d.MemberRepository, bookService)

	return Service{
		Author: authorService,
		Book:   bookService,
		Member: memberService,
	}
}
