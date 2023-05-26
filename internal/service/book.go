package service

import (
	"context"

	"library/internal/dto"
	"library/internal/entity"
	"library/internal/repository"
)

type Book struct {
	bookRepository repository.Book
	authorService  Author
}

func NewBookService(b repository.Book, a Author) Book {
	return Book{
		bookRepository: b,
		authorService:  a,
	}
}

func (s *Book) List(ctx context.Context) (res []dto.BookResponse, err error) {
	data, err := s.bookRepository.SelectRows(ctx)
	if err != nil {
		return
	}
	res = dto.ParseFromBooks(data)

	return
}

func (s *Book) Add(ctx context.Context, req dto.BookRequest) (res dto.BookResponse, err error) {
	data := entity.Book{
		Name:    &req.Name,
		Genre:   &req.Genre,
		ISBN:    &req.ISBN,
		Authors: req.Authors,
	}

	data.ID, err = s.bookRepository.CreateRow(ctx, data)
	if err != nil {
		return
	}
	res = dto.ParseFromBook(data)

	return
}

func (s *Book) Get(ctx context.Context, id string) (res dto.BookResponse, err error) {
	data, err := s.bookRepository.GetRow(ctx, id)
	if err != nil {
		return
	}
	res = dto.ParseFromBook(data)

	return
}

func (s *Book) Update(ctx context.Context, id string, req dto.BookRequest) (err error) {
	data := entity.Book{
		Name:    &req.Name,
		Genre:   &req.Genre,
		ISBN:    &req.ISBN,
		Authors: req.Authors,
	}
	return s.bookRepository.UpdateRow(ctx, id, data)
}

func (s *Book) Delete(ctx context.Context, id string) (err error) {
	return s.bookRepository.DeleteRow(ctx, id)
}

func (s *Book) ListAuthor(ctx context.Context, id string) (res []dto.AuthorResponse, err error) {
	data, err := s.bookRepository.GetRow(ctx, id)
	if err != nil {
		return
	}
	res = make([]dto.AuthorResponse, len(data.Authors))

	for i := 0; i < len(data.Authors); i++ {
		res[i], err = s.authorService.Get(ctx, data.Authors[i])
		if err != nil {
			return
		}
	}

	return
}
