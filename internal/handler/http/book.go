package http

import (
	"errors"
	"library-service/internal/service/interfaces"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"library-service/internal/domain/book"
	"library-service/pkg/server/response"
	"library-service/pkg/store"
)

type BookHandler struct {
	bookService interfaces.BookService
}

func NewBookHandler(s interfaces.BookService) *BookHandler {
	return &BookHandler{bookService: s}
}

func (h *BookHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.add)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Put("/", h.update)
		r.Delete("/", h.delete)
		r.Get("/authors", h.listAuthors)
	})

	return r
}

// @Summary	list of books from the repository
// @Tags		books
// @Accept		json
// @Produce	json
// @Success	200		{array}		book.Response
// @Failure	401		{object}	response.Object
// @Failure	500		{object}	response.Object
// @Security	BearerAuth
// @Router		/books 	[get]
func (h *BookHandler) list(w http.ResponseWriter, r *http.Request) {
	res, err := h.bookService.ListBooks(r.Context())
	if err != nil {
		response.InternalServerError(w, r, err, nil)
		return
	}

	response.OK(w, r, res)
}

// @Summary	add a new book to the repository
// @Tags		books
// @Accept		json
// @Produce	json
// @Param		request	body		book.Request	true	"body param"
// @Success	200		{object}	book.Response
// @Failure	400		{object}	response.Object
// @Failure	401		{object}	response.Object
// @Failure	500		{object}	response.Object
// @Security	BearerAuth
// @Router		/books [post]
func (h *BookHandler) add(w http.ResponseWriter, r *http.Request) {
	req := book.Request{}
	if err := render.Bind(r, &req); err != nil {
		response.BadRequest(w, r, err, req)
		return
	}

	res, err := h.bookService.CreateBook(r.Context(), req)
	if err != nil {
		response.InternalServerError(w, r, err, nil)
		return
	}

	response.OK(w, r, res)
}

// @Summary	get the book from the repository
// @Tags		books
// @Accept		json
// @Produce	json
// @Param		id	path		string	true	"path param"
// @Success	200	{object}	book.Response
// @Failure	401	{object}	response.Object
// @Failure	404	{object}	response.Object
// @Failure	500	{object}	response.Object
// @Security	BearerAuth
// @Router		/books/{id} [get]
func (h *BookHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.bookService.GetBook(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrorNotFound):
			response.NotFound(w, r, err)
		default:
			response.InternalServerError(w, r, err, nil)
		}
		return
	}

	response.OK(w, r, res)
}

// @Summary	update the book in the repository
// @Tags		books
// @Accept		json
// @Produce	json
// @Param		id		path	string				true	"path param"
// @Param		request	body	book.Request	true	"body param"
// @Success	200
// @Failure	400	{object}	response.Object
// @Failure	401	{object}	response.Object
// @Failure	404	{object}	response.Object
// @Failure	500	{object}	response.Object
// @Security	BearerAuth
// @Router		/books/{id} [put]
func (h *BookHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := book.Request{}
	if err := render.Bind(r, &req); err != nil {
		response.BadRequest(w, r, err, req)
		return
	}

	if err := h.bookService.UpdateBook(r.Context(), id, req); err != nil {
		switch {
		case errors.Is(err, store.ErrorNotFound):
			response.NotFound(w, r, err)
		default:
			response.InternalServerError(w, r, err, nil)
		}
		return
	}
}

// @Summary	delete the book from the repository
// @Tags		books
// @Accept		json
// @Produce	json
// @Param		id	path	string	true	"path param"
// @Success	200
// @Failure	401	{object}	response.Object
// @Failure	404	{object}	response.Object
// @Failure	500	{object}	response.Object
// @Security	BearerAuth
// @Router		/books/{id} [delete]
func (h *BookHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.bookService.DeleteBook(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, store.ErrorNotFound):
			response.NotFound(w, r, err)
		default:
			response.InternalServerError(w, r, err, nil)
		}
		return
	}
}

// @Summary	list of authors for a specific book
// @Tags		books
// @Accept		json
// @Produce	json
// @Param		id	path		string	true	"book id"
// @Success	200	{array}		author.Response
// @Failure	401	{object}	response.Object
// @Failure	404	{object}	response.Object
// @Failure	500	{object}	response.Object
// @Security	BearerAuth
// @Router		/books/{id}/authors [get]
func (h *BookHandler) listAuthors(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.bookService.ListBookAuthors(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrorNotFound):
			response.NotFound(w, r, err)
		default:
			response.InternalServerError(w, r, err, nil)
		}
		return
	}

	response.OK(w, r, res)
}
