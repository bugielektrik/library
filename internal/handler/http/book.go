package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"library/internal/dto"
	"library/internal/service"
)

type BookHandler struct {
	bookService service.Book
}

func NewBookHandler(a service.Book) *BookHandler {
	return &BookHandler{bookService: a}
}

func (h *BookHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.add)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Put("/", h.update)
		r.Delete("/", h.delete)
		r.Get("/authors", h.listAuthor)
	})

	return r
}

// List of books from the store
//
//	@Summary	List of books from the store
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Success	200		{array}		dto.BookResponse
//	@Failure	500		{object}	dto.Response
//	@Router		/books 	[get]
func (h *BookHandler) list(w http.ResponseWriter, r *http.Request) {
	res, err := h.bookService.List(r.Context())
	if err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}

	render.JSON(w, r, dto.OK(res))
}

// Add a new book to the store
//
//	@Summary	Add a new book to the store
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		request	body		dto.BookRequest	true	"body param"
//	@Success	200		{object}	dto.BookResponse
//	@Failure	400		{object}	dto.Response
//	@Failure	500		{object}	dto.Response
//	@Router		/books [post]
func (h *BookHandler) add(w http.ResponseWriter, r *http.Request) {
	req := dto.BookRequest{}
	if err := render.Bind(r, &req); err != nil {
		render.JSON(w, r, dto.BadRequest(err, req))
		return
	}

	res, err := h.bookService.Add(r.Context(), req)
	if err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}

	render.JSON(w, r, dto.OK(res))
}

// Read the book from the store
//
//	@Summary	Read the book from the store
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int	true	"path param"
//	@Success	200	{object}	dto.BookResponse
//	@Failure	500	{object}	dto.Response
//	@Router		/books/{id} [get]
func (h *BookHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.bookService.Get(r.Context(), id)
	if err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}

	render.JSON(w, r, dto.OK(res))
}

// Update the book in the store
//
//	@Summary	Update the book in the store
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		id		path	int				true	"path param"
//	@Param		request	body	dto.BookRequest	true	"body param"
//	@Success	200
//	@Failure	400	{object}	dto.Response
//	@Failure	500	{object}	dto.Response
//	@Router		/books/{id} [put]
func (h *BookHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := dto.BookRequest{}
	if err := render.Bind(r, &req); err != nil {
		render.JSON(w, r, dto.BadRequest(err, req))
		return
	}

	if err := h.bookService.Update(r.Context(), id, req); err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}

	render.NoContent(w, r)
}

// Delete the book from the store
//
//	@Summary	Delete the book from the store
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		id	path	int	true	"path param"
//	@Success	200
//	@Failure	500	{object}	dto.Response
//	@Router		/books/{id} [delete]
func (h *BookHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.bookService.Delete(r.Context(), id); err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}

	render.NoContent(w, r)
}

// List of authors from the store
//
//	@Summary	List of authors from the store
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int	true	"path param"
//	@Success	200	{array}		dto.AuthorResponse
//	@Failure	500	{object}	dto.Response
//	@Router		/books/{id}/authors [get]
func (h *BookHandler) listAuthor(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.bookService.ListAuthor(r.Context(), id)
	if err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}

	render.JSON(w, r, dto.OK(res))
}
