package library

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"library-service/internal/domain/author"
	"library-service/internal/domain/book"
	"library-service/pkg/log"
	"library-service/pkg/store"
)

// ListBooks retrieves all books from the repository.
func (s *Service) ListBooks(ctx context.Context) ([]book.Response, error) {
	logger := log.LoggerFromContext(ctx).Named("list_books")

	books, err := s.bookRepository.List(ctx)
	if err != nil {
		logger.Error("failed to list books", zap.Error(err))
		return nil, err
	}

	return book.ParseFromEntities(books), nil
}

// CreateBook adds a new book to the repository.
func (s *Service) CreateBook(ctx context.Context, req book.Request) (book.Response, error) {
	logger := log.LoggerFromContext(ctx).Named("create_book")

	newBook := book.Entity{
		Name:    &req.Name,
		Genre:   &req.Genre,
		ISBN:    &req.ISBN,
		Authors: req.Authors,
	}

	id, err := s.bookRepository.Add(ctx, newBook)
	if err != nil {
		logger.Error("failed to create book", zap.Error(err))
		return book.Response{}, err
	}
	newBook.ID = id

	// Cache the newly created book
	if err := s.bookCache.Set(ctx, id, newBook); err != nil {
		logger.Warn("failed to cache new book", zap.Error(err))
	}

	return book.ParseFromEntity(newBook), nil
}

// GetBook retrieves a book by ID from the repository.
func (s *Service) GetBook(ctx context.Context, id string) (book.Response, error) {
	logger := log.LoggerFromContext(ctx).Named("get_book").With(zap.String("id", id))

	// Try to get the book from the cache
	bookEntity, err := s.bookCache.Get(ctx, id)
	if err == nil {
		return book.ParseFromEntity(bookEntity), nil
	}

	// If not found in cache, get it from the repository
	bookEntity, err = s.bookRepository.Get(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("book not found", zap.Error(err))
			return book.Response{}, err
		}
		logger.Error("failed to get book", zap.Error(err))
		return book.Response{}, err
	}

	// Cache the retrieved book
	if err := s.bookCache.Set(ctx, id, bookEntity); err != nil {
		logger.Warn("failed to cache book", zap.Error(err))
	}

	return book.ParseFromEntity(bookEntity), nil
}

// UpdateBook updates an existing book in the repository.
func (s *Service) UpdateBook(ctx context.Context, id string, req book.Request) error {
	logger := log.LoggerFromContext(ctx).Named("update_book").With(zap.String("id", id))

	updatedBook := book.Entity{
		Name:    &req.Name,
		Genre:   &req.Genre,
		ISBN:    &req.ISBN,
		Authors: req.Authors,
	}

	err := s.bookRepository.Update(ctx, id, updatedBook)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("book not found", zap.Error(err))
			return err
		}
		logger.Error("failed to update book", zap.Error(err))
		return err
	}

	// Update the cache with the new book data
	if err := s.bookCache.Set(ctx, id, updatedBook); err != nil {
		logger.Warn("failed to update cache for book", zap.Error(err))
	}

	return nil
}

// DeleteBook deletes a book by ID from the repository.
func (s *Service) DeleteBook(ctx context.Context, id string) error {
	logger := log.LoggerFromContext(ctx).Named("delete_book").With(zap.String("id", id))

	err := s.bookRepository.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("book not found", zap.Error(err))
			return err
		}
		logger.Error("failed to delete book", zap.Error(err))
		return err
	}

	// Remove the book from the cache
	if err := s.bookCache.Set(ctx, id, book.Entity{}); err != nil {
		logger.Warn("failed to remove book from cache", zap.Error(err))
	}

	return nil
}

// ListBookAuthors retrieves all authors of a book by book ID.
func (s *Service) ListBookAuthors(ctx context.Context, id string) ([]author.Response, error) {
	logger := log.LoggerFromContext(ctx).Named("list_book_authors").With(zap.String("id", id))

	bookEntity, err := s.bookRepository.Get(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("book not found", zap.Error(err))
			return nil, err
		}
		logger.Error("failed to get book", zap.Error(err))
		return nil, err
	}

	authors := make([]author.Response, len(bookEntity.Authors))
	for i, authorID := range bookEntity.Authors {
		authorResp, err := s.GetAuthor(ctx, authorID)
		if err != nil {
			if errors.Is(err, store.ErrorNotFound) {
				logger.Warn("author not found", zap.Error(err))
				continue
			}
			logger.Error("failed to get author", zap.Error(err))
			return nil, err
		}
		authors[i] = authorResp
	}

	return authors, nil
}
