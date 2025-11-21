package library

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"library-service/internal/domain/author"
	"library-service/pkg/log"
	"library-service/pkg/store"
)

func (s *Service) ListAuthors(ctx context.Context) ([]author.Response, error) {
	logger := log.FromContext(ctx).Named("list_authors")

	authors, err := s.authorRepository.List(ctx)
	if err != nil {
		logger.Error("failed to list authors", zap.Error(err))
		return nil, err
	}
	return author.ParseFromEntities(authors), nil
}

func (s *Service) AddAuthor(ctx context.Context, req author.Request) (author.Response, error) {
	logger := log.FromContext(ctx).Named("add_author").With(zap.Any("author", req))

	newAuthor := author.New(req)

	id, err := s.authorRepository.Add(ctx, newAuthor)
	if err != nil {
		logger.Error("failed to add author", zap.Error(err))
		return author.Response{}, err
	}
	newAuthor.ID = id

	if err := s.authorCache.Set(ctx, id, newAuthor); err != nil {
		logger.Warn("failed to cache new author", zap.Error(err))
	}

	return author.ParseFromEntity(newAuthor), nil
}

func (s *Service) GetAuthor(ctx context.Context, id string) (author.Response, error) {
	logger := log.FromContext(ctx).Named("get_author").With(zap.String("id", id))

	cachedAuthor, err := s.authorCache.Get(ctx, id)
	if err == nil {
		return author.ParseFromEntity(cachedAuthor), nil
	}

	repoAuthor, err := s.authorRepository.Get(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("author not found", zap.Error(err))
			return author.Response{}, err
		}
		logger.Error("failed to get author", zap.Error(err))
		return author.Response{}, err
	}

	if cacheErr := s.authorCache.Set(ctx, id, repoAuthor); cacheErr != nil {
		logger.Warn("failed to cache author", zap.Error(cacheErr))
	}

	return author.ParseFromEntity(repoAuthor), nil
}

func (s *Service) UpdateAuthor(ctx context.Context, id string, req author.Request) error {
	logger := log.FromContext(ctx).Named("update_author").With(zap.String("id", id), zap.Any("author", req))

	updatedAuthor := author.New(req)

	err := s.authorRepository.Update(ctx, id, updatedAuthor)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("author not found", zap.Error(err))
			return err
		}
		logger.Error("failed to update author", zap.Error(err))
		return err
	}

	if err := s.authorCache.Set(ctx, id, updatedAuthor); err != nil {
		logger.Warn("failed to update cache for author", zap.Error(err))
	}

	return nil
}

func (s *Service) DeleteAuthor(ctx context.Context, id string) error {
	logger := log.FromContext(ctx).Named("delete_author").With(zap.String("id", id))

	err := s.authorRepository.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("author not found", zap.Error(err))
			return err
		}
		logger.Error("failed to delete author", zap.Error(err))
		return err
	}

	if err := s.authorCache.Set(ctx, id, author.Entity{}); err != nil {
		logger.Warn("failed to remove author from cache", zap.Error(err))
	}

	return nil
}
