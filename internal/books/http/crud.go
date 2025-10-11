package book

import (
	"net/http"

	"go.uber.org/zap"

	"library-service/internal/adapters/http/dto"
	"library-service/internal/books/operations"
	"library-service/pkg/httputil"
	"library-service/pkg/logutil"
)

// This file contains CRUD operations for books.

// @Summary Create a new book
// @Tags books
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateBookRequest true "Book data"
// @Success 201 {object} dto.BookResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /books [post]
func (h *BookHandler) create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "book_handler", "create")

	// Decode request
	var req dto.CreateBookRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Validate request
	if !h.validator.ValidateStruct(w, req) {
		return
	}

	// Execute usecase
	result, err := h.useCases.Book.CreateBook.Execute(ctx, operations.CreateBookRequest{
		Name:    req.Name,
		Genre:   req.Genre,
		ISBN:    req.ISBN,
		Authors: req.Authors,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTO
	response := dto.ToBookResponseFromCreate(result)

	logger.Info("book created", zap.String("id", response.ID))
	h.RespondJSON(w, http.StatusCreated, response)
}

// @Summary Get a book by ID
// @Tags books
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Success 200 {object} dto.BookResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /books/{id} [get]
func (h *BookHandler) get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "book_handler", "get")

	id, ok := h.GetURLParam(w, r, "id")
	if !ok {
		return
	}

	// Execute usecase
	result, err := h.useCases.Book.GetBook.Execute(ctx, operations.GetBookRequest{ID: id})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTO
	response := dto.ToBookResponseFromGet(result)

	logger.Debug("book retrieved", zap.String("id", id))
	h.RespondJSON(w, http.StatusOK, response)
}

// @Summary Update a book
// @Tags books
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Param request body dto.UpdateBookRequest true "Book data"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /books/{id} [put]
func (h *BookHandler) update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "book_handler", "update")

	id, ok := h.GetURLParam(w, r, "id")
	if !ok {
		return
	}

	// Decode request
	var req dto.UpdateBookRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Validate request
	if !h.validator.ValidateStruct(w, req) {
		return
	}

	// Execute usecase
	response, err := h.useCases.Book.UpdateBook.Execute(ctx, operations.UpdateBookRequest{
		ID:      id,
		Name:    req.Name,
		Genre:   req.Genre,
		ISBN:    req.ISBN,
		Authors: req.Authors,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	logger.Info("book updated", zap.String("id", id))
	h.RespondJSON(w, http.StatusOK, response)
}

// @Summary Delete a book
// @Tags books
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Success 204
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /books/{id} [delete]
func (h *BookHandler) delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "book_handler", "delete")

	id, ok := h.GetURLParam(w, r, "id")
	if !ok {
		return
	}

	// Execute usecase
	response, err := h.useCases.Book.DeleteBook.Execute(ctx, operations.DeleteBookRequest{ID: id})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	logger.Info("book deleted", zap.String("id", id))
	h.RespondJSON(w, http.StatusOK, response)
}
