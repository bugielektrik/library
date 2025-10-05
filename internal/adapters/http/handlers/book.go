package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"library-service/internal/adapters/http/dto"
	"library-service/internal/adapters/http/middleware"
	"library-service/internal/usecase/bookops"
	"library-service/pkg/errors"

	log "library-service/internal/infrastructure/log"
)

// BookHandler handles HTTP requests for books
type BookHandler struct {
	createBookUC      *bookops.CreateBookUseCase
	getBookUC         *bookops.GetBookUseCase
	listBooksUC       *bookops.ListBooksUseCase
	updateBookUC      *bookops.UpdateBookUseCase
	deleteBookUC      *bookops.DeleteBookUseCase
	listBookAuthorsUC *bookops.ListBookAuthorsUseCase
	validator         *middleware.Validator
}

// NewBookHandler creates a new book handler
func NewBookHandler(
	createBookUC *bookops.CreateBookUseCase,
	getBookUC *bookops.GetBookUseCase,
	listBooksUC *bookops.ListBooksUseCase,
	updateBookUC *bookops.UpdateBookUseCase,
	deleteBookUC *bookops.DeleteBookUseCase,
	listBookAuthorsUC *bookops.ListBookAuthorsUseCase,
) *BookHandler {
	return &BookHandler{
		createBookUC:      createBookUC,
		getBookUC:         getBookUC,
		listBooksUC:       listBooksUC,
		updateBookUC:      updateBookUC,
		deleteBookUC:      deleteBookUC,
		listBookAuthorsUC: listBookAuthorsUC,
		validator:         middleware.NewValidator(),
	}
}

// Routes returns the router for book endpoints
func (h *BookHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.create)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Put("/", h.update)
		r.Delete("/", h.delete)
		r.Get("/authors", h.listAuthors)
	})

	return r
}

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
	logger := log.FromContext(ctx).Named("book_handler.list")

	// Execute usecase
	result, err := h.listBooksUC.Execute(ctx, bookops.ListBooksRequest{})
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	// Convert to DTOs
	books := make([]dto.BookResponse, len(result.Books))
	for i, book := range result.Books {
		books[i] = dto.BookResponse{
			ID:      book.ID,
			Name:    book.Name,
			Genre:   book.Genre,
			ISBN:    book.ISBN,
			Authors: book.Authors,
		}
	}

	logger.Info("books listed", zap.Int("count", len(books)))
	h.respondJSON(w, http.StatusOK, books)
}

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
	logger := log.FromContext(ctx).Named("book_handler.create")

	// Decode request
	var req dto.CreateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, r, errors.ErrInvalidInput.Wrap(err))
		return
	}

	// Validate request
	if !h.validator.ValidateStruct(w, req) {
		return
	}

	// Execute usecase
	result, err := h.createBookUC.Execute(ctx, bookops.CreateBookRequest{
		Name:    req.Name,
		Genre:   req.Genre,
		ISBN:    req.ISBN,
		Authors: req.Authors,
	})
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	// Convert to DTO
	response := dto.BookResponse{
		ID:      result.ID,
		Name:    result.Name,
		Genre:   result.Genre,
		ISBN:    result.ISBN,
		Authors: result.Authors,
	}

	logger.Info("book created", zap.String("id", response.ID))
	h.respondJSON(w, http.StatusCreated, response)
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
	logger := log.FromContext(ctx).Named("book_handler.get")

	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondError(w, r, errors.ErrInvalidInput.WithDetails("field", "id"))
		return
	}

	// Execute usecase
	result, err := h.getBookUC.Execute(ctx, bookops.GetBookRequest{ID: id})
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	// Convert to DTO
	response := dto.BookResponse{
		ID:      result.ID,
		Name:    result.Name,
		Genre:   result.Genre,
		ISBN:    result.ISBN,
		Authors: result.Authors,
	}

	logger.Debug("book retrieved", zap.String("id", id))
	h.respondJSON(w, http.StatusOK, response)
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
	logger := log.FromContext(ctx).Named("book_handler.update")

	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondError(w, r, errors.ErrInvalidInput.WithDetails("field", "id"))
		return
	}

	// Decode request
	var req dto.UpdateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, r, errors.ErrInvalidInput.Wrap(err))
		return
	}

	// Validate request
	if !h.validator.ValidateStruct(w, req) {
		return
	}

	// Execute usecase
	err := h.updateBookUC.Execute(ctx, bookops.UpdateBookRequest{
		ID:      id,
		Name:    req.Name,
		Genre:   req.Genre,
		ISBN:    req.ISBN,
		Authors: req.Authors,
	})
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	logger.Info("book updated", zap.String("id", id))
	w.WriteHeader(http.StatusNoContent)
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
	logger := log.FromContext(ctx).Named("book_handler.delete")

	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondError(w, r, errors.ErrInvalidInput.WithDetails("field", "id"))
		return
	}

	// Execute usecase
	err := h.deleteBookUC.Execute(ctx, bookops.DeleteBookRequest{ID: id})
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	logger.Info("book deleted", zap.String("id", id))
	w.WriteHeader(http.StatusNoContent)
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
	logger := log.FromContext(ctx).Named("book_handler.list_authors")

	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondError(w, r, errors.ErrInvalidInput.WithDetails("field", "id"))
		return
	}

	// Execute usecase
	result, err := h.listBookAuthorsUC.Execute(ctx, bookops.ListBookAuthorsRequest{BookID: id})
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	// Convert to DTOs
	authors := make([]dto.AuthorResponse, len(result.Authors))
	for i, author := range result.Authors {
		authors[i] = dto.AuthorResponse{
			ID:        author.ID,
			FullName:  author.FullName,
			Pseudonym: author.Pseudonym,
			Specialty: author.Specialty,
		}
	}

	logger.Info("book authors listed", zap.String("book_id", id), zap.Int("count", len(authors)))
	h.respondJSON(w, http.StatusOK, authors)
}

// respondJSON writes a JSON response
func (h *BookHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError writes an error response
func (h *BookHandler) respondError(w http.ResponseWriter, r *http.Request, err error) {
	logger := log.FromContext(r.Context())

	status := errors.GetHTTPStatus(err)

	if status >= 500 {
		logger.Error("internal error", zap.Error(err))
	} else {
		logger.Warn("client error", zap.Error(err), zap.Int("status", status))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := dto.FromError(err)
	json.NewEncoder(w).Encode(response)
}
