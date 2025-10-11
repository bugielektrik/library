package book

import (
	"net/http"

	"go.uber.org/zap"

	"library-service/internal/adapters/http/dto"
	"library-service/internal/usecase/bookops"
	"library-service/pkg/logutil"
)

// This file contains query operations for books.

// @Summary List all books
// @Tags books
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.BookResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /books [get]
func (h *BookHandler) list(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "book_handler", "list")

	// Execute usecase
	result, err := h.useCases.Book.ListBooks.Execute(ctx, bookops.ListBooksRequest{})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTOs
	books := dto.ToBookResponses(result.Books)

	logger.Info("books listed", zap.Int("count", len(books)))
	h.RespondJSON(w, http.StatusOK, books)
}

// @Summary List authors of a book
// @Tags books
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Success 200 {array} dto.AuthorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /books/{id}/authors [get]
func (h *BookHandler) listAuthors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "book_handler", "list_authors")

	id, ok := h.GetURLParam(w, r, "id")
	if !ok {
		return
	}

	// Execute usecase
	result, err := h.useCases.Book.ListBookAuthors.Execute(ctx, bookops.ListBookAuthorsRequest{BookID: id})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTOs
	authors := dto.ToAuthorResponses(result.Authors)

	logger.Info("book authors listed", zap.String("book_id", id), zap.Int("count", len(authors)))
	h.RespondJSON(w, http.StatusOK, authors)
}
