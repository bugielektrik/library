package service

import (
	"context"

	"library/internal/dto"
	"library/internal/entity"
	"library/internal/repository"
)

type BookService interface {
	List(ctx context.Context) (res []dto.BookResponse, err error)
	Add(ctx context.Context, req dto.BookRequest) (res dto.BookResponse, err error)
	Get(ctx context.Context, id string) (res dto.BookResponse, err error)
	Update(ctx context.Context, id string, req dto.BookRequest) (err error)
	Delete(ctx context.Context, id string) (err error)
}

type bookService struct {
	bookRepository repository.BookRepository
}

func NewBookService(b repository.BookRepository) BookService {
	return &bookService{
		bookRepository: b,
	}
}

func (s *bookService) List(ctx context.Context) (res []dto.BookResponse, err error) {
	data, err := s.bookRepository.SelectRows(ctx)
	if err != nil {
		return
	}
	res = dto.ParseFromBooks(data)

	return
}

func (s *bookService) Add(ctx context.Context, req dto.BookRequest) (res dto.BookResponse, err error) {
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

func (s *bookService) Get(ctx context.Context, id string) (res dto.BookResponse, err error) {
	data, err := s.bookRepository.GetRow(ctx, id)
	if err != nil {
		return
	}
	res = dto.ParseFromBook(data)

	return
}

func (s *bookService) Update(ctx context.Context, id string, req dto.BookRequest) (err error) {
	data := entity.Book{
		Name:    &req.Name,
		Genre:   &req.Genre,
		ISBN:    &req.ISBN,
		Authors: req.Authors,
	}
	return s.bookRepository.UpdateRow(ctx, id, data)
}

func (s *bookService) Delete(ctx context.Context, id string) (err error) {
	return s.bookRepository.DeleteRow(ctx, id)
}
