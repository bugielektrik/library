package library

import (
	"context"

	"library/internal/domain/author"
	"library/internal/domain/book"
)

func (s *Service) ListBooks(ctx context.Context) (res []book.Response, err error) {
	data, err := s.bookRepository.Select(ctx)
	if err != nil {
		return
	}
	res = book.ParseFromEntities(data)

	return
}

func (s *Service) AddBook(ctx context.Context, req book.Request) (res book.Response, err error) {
	data := book.Entity{
		Name:    &req.Name,
		Genre:   &req.Genre,
		ISBN:    &req.ISBN,
		Authors: req.Authors,
	}

	data.ID, err = s.bookRepository.Create(ctx, data)
	if err != nil {
		return
	}
	res = book.ParseFromEntity(data)

	return
}

func (s *Service) GetBook(ctx context.Context, id string) (res book.Response, err error) {
	data, err := s.bookRepository.Get(ctx, id)
	if err != nil {
		return
	}
	res = book.ParseFromEntity(data)

	return
}

func (s *Service) UpdateBook(ctx context.Context, id string, req book.Request) (err error) {
	data := book.Entity{
		Name:    &req.Name,
		Genre:   &req.Genre,
		ISBN:    &req.ISBN,
		Authors: req.Authors,
	}
	return s.bookRepository.Update(ctx, id, data)
}

func (s *Service) DeleteBook(ctx context.Context, id string) (err error) {
	return s.bookRepository.Delete(ctx, id)
}

func (s *Service) ListBookAuthors(ctx context.Context, id string) (res []author.Response, err error) {
	data, err := s.bookRepository.Get(ctx, id)
	if err != nil {
		return
	}
	res = make([]author.Response, len(data.Authors))

	for i := 0; i < len(data.Authors); i++ {
		res[i], err = s.GetAuthor(ctx, data.Authors[i])
		if err != nil {
			return
		}
	}

	return
}
