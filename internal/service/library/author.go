package library

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/domain/author"
	"library-service/pkg/log"
)

func (s *Service) ListAuthors(ctx context.Context) (res []author.Response, err error) {
	logger := log.LoggerFromContext(ctx).Named("ListAuthors")

	data, err := s.authorRepository.List(ctx)
	if err != nil {
		logger.Error("failed to select", zap.Error(err))
		return
	}
	res = author.ParseFromEntities(data)

	return
}

func (s *Service) CreateAuthor(ctx context.Context, req author.Request) (res author.Response, err error) {
	logger := log.LoggerFromContext(ctx).Named("CreateAuthor")

	data := author.Entity{
		FullName:  &req.FullName,
		Pseudonym: &req.Pseudonym,
		Specialty: &req.Specialty,
	}

	data.ID, err = s.authorRepository.Create(ctx, data)
	if err != nil {
		logger.Error("failed to create", zap.Error(err))
		return
	}
	res = author.ParseFromEntity(data)

	return
}

func (s *Service) GetAuthor(ctx context.Context, id string) (res author.Response, err error) {
	logger := log.LoggerFromContext(ctx).Named("GetAuthor").With(zap.String("id", id))

	data, err := s.authorRepository.Get(ctx, id)
	if err != nil {
		logger.Error("failed to get by id", zap.Error(err))
		return
	}
	res = author.ParseFromEntity(data)

	return
}

func (s *Service) UpdateAuthor(ctx context.Context, id string, req author.Request) (err error) {
	logger := log.LoggerFromContext(ctx).Named("UpdateAuthor").With(zap.String("id", id))

	data := author.Entity{
		FullName:  &req.FullName,
		Pseudonym: &req.Pseudonym,
		Specialty: &req.Specialty,
	}

	err = s.authorRepository.Update(ctx, id, data)
	if err != nil {
		logger.Error("failed to update by id", zap.Error(err))
		return
	}

	return
}

func (s *Service) DeleteAuthor(ctx context.Context, id string) (err error) {
	logger := log.LoggerFromContext(ctx).Named("DeleteAuthor").With(zap.String("id", id))

	err = s.authorRepository.Delete(ctx, id)
	if err != nil {
		logger.Error("failed to delete by id", zap.Error(err))
		return
	}

	return
}
