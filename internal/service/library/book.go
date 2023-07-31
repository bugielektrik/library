package library

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/domain/author"
	"library-service/internal/domain/book"
	"library-service/pkg/log"
)

func (s *Service) ListBooks(ctx context.Context) (res []book.Response, err error) {
	logger := log.LoggerFromContext(ctx).Named("ListBooks")

	data, err := s.bookRepository.Select(ctx)
	if err != nil {
		logger.Error("failed to select", zap.Error(err))
		return
	}
	res = book.ParseFromEntities(data)

	return
}

func (s *Service) AddBook(ctx context.Context, req book.Request) (res book.Response, err error) {
	logger := log.LoggerFromContext(ctx).Named("AddBook")

	data := book.Entity{
		Name:    &req.Name,
		Genre:   &req.Genre,
		ISBN:    &req.ISBN,
		Authors: req.Authors,
	}

	data.ID, err = s.bookRepository.Insert(ctx, data)
	if err != nil {
		logger.Error("failed to create", zap.Error(err))
		return
	}
	res = book.ParseFromEntity(data)

	return
}

func (s *Service) GetBookByID(ctx context.Context, id string) (res book.Response, err error) {
	logger := log.LoggerFromContext(ctx).Named("GetBookByID").With(zap.String("id", id))

	data, err := s.bookRepository.Get(ctx, id)
	if err != nil {
		logger.Error("failed to get by id", zap.Error(err))
		return
	}
	res = book.ParseFromEntity(data)

	return
}

func (s *Service) UpdateBook(ctx context.Context, id string, req book.Request) (err error) {
	logger := log.LoggerFromContext(ctx).Named("UpdateBook").With(zap.String("id", id))

	data := book.Entity{
		Name:    &req.Name,
		Genre:   &req.Genre,
		ISBN:    &req.ISBN,
		Authors: req.Authors,
	}

	err = s.bookRepository.Update(ctx, id, data)
	if err != nil {
		logger.Error("failed to update by id", zap.Error(err))
		return
	}

	return
}

func (s *Service) DeleteBook(ctx context.Context, id string) (err error) {
	logger := log.LoggerFromContext(ctx).Named("DeleteBook").With(zap.String("id", id))

	err = s.bookRepository.Delete(ctx, id)
	if err != nil {
		logger.Error("failed to delete by id", zap.Error(err))
		return
	}

	return
}

func (s *Service) ListBookAuthors(ctx context.Context, id string) (res []author.Response, err error) {
	logger := log.LoggerFromContext(ctx).Named("ListBookAuthors").With(zap.String("id", id))

	data, err := s.bookRepository.Get(ctx, id)
	if err != nil {
		logger.Error("failed to get by id", zap.Error(err))
		return
	}
	res = make([]author.Response, len(data.Authors))

	for i := 0; i < len(data.Authors); i++ {
		res[i], err = s.GetAuthorByID(ctx, data.Authors[i])
		if err != nil {
			logger.Error("failed to get author by id", zap.Error(err))
			return
		}
	}

	return
}
