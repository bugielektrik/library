package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"library/internal/domain/book"
	"library/internal/service/library"
	"library/pkg/server/status"
)

type BookHandler struct {
	libraryService *library.Service
}

func NewBookHandler(s *library.Service) *BookHandler {
	return &BookHandler{libraryService: s}
}

func (h *BookHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.add)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Put("/", h.update)
		r.Delete("/", h.delete)
	})

	return r
}

// List of books from the store
//
//	@Summary	List of books from the store
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Success	200		{array}		book.Response
//	@Failure	500		{object}	status.Response
//	@Router		/books 	[get]
func (h *BookHandler) list(w http.ResponseWriter, r *http.Request) {
	res, err := h.libraryService.ListBooks(r.Context())
	if err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}

	render.JSON(w, r, status.OK(res))
}

// Add a new book to the store
//
//	@Summary	Add a new book to the store
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		request	body		book.Request	true	"body param"
//	@Success	200		{object}	book.Response
//	@Failure	400		{object}	status.Response
//	@Failure	500		{object}	status.Response
//	@Router		/books [post]
func (h *BookHandler) add(w http.ResponseWriter, r *http.Request) {
	req := book.Request{}
	if err := render.Bind(r, &req); err != nil {
		render.JSON(w, r, status.BadRequest(err, req))
		return
	}

	res, err := h.libraryService.AddBook(r.Context(), req)
	if err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}

	render.JSON(w, r, status.OK(res))
}

// Read the book from the store
//
//	@Summary	Read the book from the store
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int	true	"path param"
//	@Success	200	{object}	book.Response
//	@Failure	500	{object}	status.Response
//	@Router		/books/{id} [get]
func (h *BookHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.libraryService.GetBook(r.Context(), id)
	if err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}

	render.JSON(w, r, status.OK(res))
}

// Update the book in the store
//
//	@Summary	Update the book in the store
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		id		path	int				true	"path param"
//	@Param		request	body	book.Request	true	"body param"
//	@Success	200
//	@Failure	400	{object}	status.Response
//	@Failure	500	{object}	status.Response
//	@Router		/books/{id} [put]
func (h *BookHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := book.Request{}
	if err := render.Bind(r, &req); err != nil {
		render.JSON(w, r, status.BadRequest(err, req))
		return
	}

	if err := h.libraryService.UpdateBook(r.Context(), id, req); err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}
}

// Delete the book from the store
//
//	@Summary	Delete the book from the store
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		id	path	int	true	"path param"
//	@Success	200
//	@Failure	500	{object}	status.Response
//	@Router		/books/{id} [delete]
func (h *BookHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.libraryService.DeleteBook(r.Context(), id); err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}
}
