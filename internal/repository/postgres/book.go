package postgres

import (
	"context"
	"errors"
	"library-service/internal/repository/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"library-service/internal/domain/book"

	"library-service/pkg/store"
)

type BookRepository struct {
	queries   *sqlc.Queries
	txManager TxManager
}

func NewBookRepository(db *pgxpool.Pool, txManager TxManager) *BookRepository {
	return &BookRepository{
		queries:   sqlc.New(db),
		txManager: txManager,
	}
}

func (r *BookRepository) List(ctx context.Context) ([]book.Entity, error) {
	dbBooks, err := r.queries.ListBooks(ctx)
	if err != nil {
		return nil, err
	}
	books := make([]book.Entity, 0, len(dbBooks))
	for _, dbBook := range dbBooks {
		books = append(books, book.Entity{
			ID:       dbBook.ID,
			Name:     &dbBook.Name,
			Genre:    &dbBook.Genre,
			ISBN:     &dbBook.Isbn,
			AuthorId: &dbBook.AuthorID,
		})
	}
	return books, nil
}

func (r *BookRepository) Add(ctx context.Context, data book.Entity) (string, error) {
	dbBook, err := r.queries.AddBook(ctx, sqlc.AddBookParams{
		Name:     *data.Name,
		Genre:    *data.Genre,
		Isbn:     *data.ISBN,
		AuthorID: *data.AuthorId,
	})
	if err != nil {
		return "", err
	}
	return dbBook, nil
}

func (r *BookRepository) Get(ctx context.Context, id string) (book.Entity, error) {
	dbBook, err := r.queries.GetBook(ctx, id)
	if err != nil {
		return book.Entity{}, err
	}
	bookEntity := book.Entity{
		ID:       dbBook.ID,
		Name:     &dbBook.Name,
		Genre:    &dbBook.Genre,
		ISBN:     &dbBook.Isbn,
		AuthorId: &dbBook.AuthorID,
	}
	return bookEntity, nil
}

func (r *BookRepository) Update(ctx context.Context, id string, data book.Entity) error {
	err := r.queries.UpdateBook(ctx, sqlc.UpdateBookParams{ID: id, Name: *data.Name, Genre: *data.Genre, Isbn: *data.ISBN, AuthorID: *data.AuthorId})
	if err != nil {
		return err
	}
	return nil
}

func (r *BookRepository) Delete(ctx context.Context, id string) error {
	err := r.queries.DeleteBook(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.ErrorNotFound
		}
		return err
	}
	return nil
}
