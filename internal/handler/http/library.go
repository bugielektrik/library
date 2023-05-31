package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"library/internal/domain/author"
	"library/internal/domain/book"
	"library/internal/service/library"
	"library/pkg/server/status"
)

type LibraryHandler struct {
	libraryService *library.Service
}

func NewLibraryHandler(l *library.Service) *LibraryHandler {
	return &LibraryHandler{libraryService: l}
}

func (h *LibraryHandler) AuthorRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.listAuthors)
	r.Post("/", h.addAuthor)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.getAuthor)
		r.Put("/", h.updateAuthor)
		r.Delete("/", h.deleteAuthor)
	})

	return r
}

func (h *LibraryHandler) BookRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.listBooks)
	r.Post("/", h.addBook)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.getBook)
		r.Put("/", h.updateBook)
		r.Delete("/", h.deleteBook)
		r.Get("/authors", h.listBookAuthors)
	})

	return r
}

// List of authors from the store
//
//	@Summary	List of authors from the store
//	@Tags		authors
//	@Accept		json
//	@Produce	json
//	@Success	200	{array}		author.Response
//	@Failure	500	{object}	status.Response
//	@Router		/authors  [get]
func (h *LibraryHandler) listAuthors(w http.ResponseWriter, r *http.Request) {
	res, err := h.libraryService.ListAuthors(r.Context())
	if err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}

	render.JSON(w, r, status.OK(res))
}

// Add a new author to the store
//
//	@Summary	Add a new author to the store
//	@Tags		authors
//	@Accept		json
//	@Produce	json
//	@Param		request	body		author.Request	true	"body param"
//	@Success	200		{object}	author.Response
//	@Failure	400		{object}	status.Response
//	@Failure	500		{object}	status.Response
//	@Router		/authors [post]
func (h *LibraryHandler) addAuthor(w http.ResponseWriter, r *http.Request) {
	req := author.Request{}
	if err := render.Bind(r, &req); err != nil {
		render.JSON(w, r, status.BadRequest(err, req))
		return
	}

	res, err := h.libraryService.AddAuthor(r.Context(), req)
	if err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}

	render.JSON(w, r, status.OK(res))
}

// Read the author from the store
//
//	@Summary	Read the author from the store
//	@Tags		authors
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int	true	"path param"
//	@Success	200	{object}	author.Response
//	@Failure	500	{object}	status.Response
//	@Router		/authors/{id} [get]
func (h *LibraryHandler) getAuthor(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.libraryService.GetAuthor(r.Context(), id)
	if err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}

	render.JSON(w, r, status.OK(res))
}

// Update the author in the store
//
//	@Summary	Update the author in the store
//	@Tags		authors
//	@Accept		json
//	@Produce	json
//	@Param		id		path	int				true	"path param"
//	@Param		request	body	author.Request	true	"body param"
//	@Success	200
//	@Failure	400	{object}	status.Response
//	@Failure	500	{object}	status.Response
//	@Router		/authors/{id} [put]
func (h *LibraryHandler) updateAuthor(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := author.Request{}
	if err := render.Bind(r, &req); err != nil {
		render.JSON(w, r, status.BadRequest(err, req))
		return
	}

	if err := h.libraryService.UpdateAuthor(r.Context(), id, req); err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}
}

// Delete the author from the store
//
//	@Summary	Delete the author from the store
//	@Tags		authors
//	@Accept		json
//	@Produce	json
//	@Param		id	path	int	true	"path param"
//	@Success	200
//	@Failure	500	{object}	status.Response
//	@Router		/authors/{id} [delete]
func (h *LibraryHandler) deleteAuthor(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.libraryService.DeleteAuthor(r.Context(), id); err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}
}

// List of books from the store
//
//	@Summary	List of books from the store
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Success	200	{array}		book.Response
//	@Failure	500	{object}	status.Response
//	@Router		/books  [get]
func (h *LibraryHandler) listBooks(w http.ResponseWriter, r *http.Request) {
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
func (h *LibraryHandler) addBook(w http.ResponseWriter, r *http.Request) {
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
func (h *LibraryHandler) getBook(w http.ResponseWriter, r *http.Request) {
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
func (h *LibraryHandler) updateBook(w http.ResponseWriter, r *http.Request) {
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
func (h *LibraryHandler) deleteBook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.libraryService.DeleteBook(r.Context(), id); err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}
}

// List of authors from the store
//
//	@Summary	List of authors from the store
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int	true	"path param"
//	@Success	200	{array}		author.Response
//	@Failure	500	{object}	status.Response
//	@Router		/books/{id}/authors [get]
func (h *LibraryHandler) listBookAuthors(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.libraryService.ListBookAuthors(r.Context(), id)
	if err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}

	render.JSON(w, r, status.OK(res))
}
