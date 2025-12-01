package book

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"library-service/internal/domain/author"
	"library-service/internal/domain/book"
	"library-service/pkg/log"
	"library-service/pkg/store"
)

type BookService struct {
	bookRepository book.Repository
	bookCache      book.Cache
}

func NewBookService(r book.Repository, c book.Cache) *BookService {
	return &BookService{bookRepository: r, bookCache: c}
}

func (s *BookService) ListBooks(ctx context.Context) ([]book.Response, error) {
	logger := log.FromContext(ctx).Named("list_books")

	books, err := s.bookRepository.List(ctx)
	if err != nil {
		logger.Error("failed to list books", zap.Error(err))
		return nil, err
	}
	return book.ParseFromEntities(books), nil
}

func (s *BookService) CreateBook(ctx context.Context, req book.Request) (book.Response, error) {
	logger := log.FromContext(ctx).Named("create_book").With(zap.Any("book", req))

	newBook := book.New(req)

	id, err := s.bookRepository.Add(ctx, newBook)
	if err != nil {
		logger.Error("failed to create book", zap.Error(err))
		return book.Response{}, err
	}
	newBook.ID = id

	if err := s.bookCache.Set(ctx, id, newBook); err != nil {
		logger.Warn("failed to cache new book", zap.Error(err))
	}

	return book.ParseFromEntity(newBook), nil
}

func (s *BookService) GetBook(ctx context.Context, id string) (book.Response, error) {
	logger := log.FromContext(ctx).Named("get_book").With(zap.String("id", id))

	bookEntity, err := s.bookCache.Get(ctx, id)
	if err == nil {
		return book.ParseFromEntity(bookEntity), nil
	}

	bookEntity, err = s.bookRepository.Get(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("book not found", zap.Error(err))
			return book.Response{}, err
		}
		logger.Error("failed to get book", zap.Error(err))
		return book.Response{}, err
	}

	if err := s.bookCache.Set(ctx, id, bookEntity); err != nil {
		logger.Warn("failed to cache book", zap.Error(err))
	}

	return book.ParseFromEntity(bookEntity), nil
}

func (s *BookService) UpdateBook(ctx context.Context, id string, req book.Request) error {
	logger := log.FromContext(ctx).Named("update_book").With(zap.String("id", id), zap.Any("book", req))

	updatedBook := book.New(req)

	err := s.bookRepository.Update(ctx, id, updatedBook)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("book not found", zap.Error(err))
			return err
		}
		logger.Error("failed to update book", zap.Error(err))
		return err
	}

	if err := s.bookCache.Set(ctx, id, updatedBook); err != nil {
		logger.Warn("failed to update cache for book", zap.Error(err))
	}

	return nil
}

func (s *BookService) DeleteBook(ctx context.Context, id string) error {
	logger := log.FromContext(ctx).Named("delete_book").With(zap.String("id", id))

	err := s.bookRepository.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("book not found", zap.Error(err))
			return err
		}
		logger.Error("failed to delete book", zap.Error(err))
		return err
	}

	if err := s.bookCache.Set(ctx, id, book.Entity{}); err != nil {
		logger.Warn("failed to remove book from cache", zap.Error(err))
	}

	return nil
}

func (s *BookService) ListBookAuthors(ctx context.Context, id string) ([]author.Response, error) {
	//logger := log.FromContext(ctx).Named("list_book_authors").With(zap.String("id", id))
	//
	//bookEntity, err := s.bookRepository.Get(ctx, id)
	//if err != nil {
	//	if errors.Is(err, store.ErrorNotFound) {
	//		logger.Warn("book not found", zap.Error(err))
	//		return nil, err
	//	}
	//	logger.Error("failed to get book", zap.Error(err))
	//	return nil, err
	//}
	//
	//authors := make([]author.Response, len(bookEntity.Authors))
	//for i, authorID := range bookEntity.Authors {
	//	authorResp, err := s.GetAuthor(ctx, authorID)
	//	if err != nil {
	//		if errors.Is(err, store.ErrorNotFound) {
	//			logger.Warn("author not found", zap.Error(err))
	//			continue
	//		}
	//		logger.Error("failed to get author", zap.Error(err))
	//		return nil, err
	//	}
	//	authors[i] = authorResp
	//}
	//
	//return authors, nil
	return nil, nil
}
