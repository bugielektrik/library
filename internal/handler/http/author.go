package http

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"library-service/internal/domain/author"
	"library-service/internal/service/library"
	"library-service/pkg/server/response"
)

type AuthorHandler struct {
	libraryService *library.Service
}

func NewAuthorHandler(s *library.Service) *AuthorHandler {
	return &AuthorHandler{libraryService: s}
}

func (h *AuthorHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.add)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.getByID)
		r.Put("/", h.update)
		r.Delete("/", h.delete)
	})

	return r
}

// List of authors from the database
//
//	@Summary	List of authors from the database
//	@Tags		authors
//	@Accept		json
//	@Produce	json
//	@Success	200			{array}		response.Object
//	@Failure	500			{object}	response.Object
//	@Router		/authors 	[get]
func (h *AuthorHandler) list(w http.ResponseWriter, r *http.Request) {
	res, err := h.libraryService.ListAuthors(r.Context())
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.OK(w, r, res)
}

// Add a new author to the database
//
//	@Summary	Add a new author to the database
//	@Tags		authors
//	@Accept		json
//	@Produce	json
//	@Param		request	body		author.Request	true	"body param"
//	@Success	200		{object}	response.Object
//	@Failure	400		{object}	response.Object
//	@Failure	500		{object}	response.Object
//	@Router		/authors [post]
func (h *AuthorHandler) add(w http.ResponseWriter, r *http.Request) {
	req := author.Request{}
	if err := render.Bind(r, &req); err != nil {
		response.BadRequest(w, r, err, req)
		return
	}

	res, err := h.libraryService.AddAuthor(r.Context(), req)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.OK(w, r, res)
}

// Read the author from the database
//
//	@Summary	Read the author from the database
//	@Tags		authors
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int	true	"path param"
//	@Success	200	{object}	response.Object
//	@Failure	404	{object}	response.Object
//	@Failure	500	{object}	response.Object
//	@Router		/authors/{id} [get]
func (h *AuthorHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.libraryService.GetAuthorByID(r.Context(), id)
	if err != nil && err != sql.ErrNoRows {
		response.InternalServerError(w, r, err)
		return
	}

	if err == sql.ErrNoRows {
		response.NotFound(w, r, err)
		return
	}

	response.OK(w, r, res)
}

// Update the author in the database
//
//	@Summary	Update the author in the database
//	@Tags		authors
//	@Accept		json
//	@Produce	json
//	@Param		id		path	int				true	"path param"
//	@Param		request	body	author.Request	true	"body param"
//	@Success	200
//	@Failure	400	{object}	response.Object
//	@Failure	404	{object}	response.Object
//	@Failure	500	{object}	response.Object
//	@Router		/authors/{id} [put]
func (h *AuthorHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := author.Request{}
	if err := render.Bind(r, &req); err != nil {
		response.BadRequest(w, r, err, req)
		return
	}

	err := h.libraryService.UpdateAuthor(r.Context(), id, req)
	if err != nil && err != sql.ErrNoRows {
		response.InternalServerError(w, r, err)
		return
	}

	if err == sql.ErrNoRows {
		response.NotFound(w, r, err)
		return
	}
}

// Delete the author from the database
//
//	@Summary	Delete the author from the database
//	@Tags		authors
//	@Accept		json
//	@Produce	json
//	@Param		id	path	int	true	"path param"
//	@Success	200
//	@Failure	404	{object}	response.Object
//	@Failure	500	{object}	response.Object
//	@Router		/authors/{id} [delete]
func (h *AuthorHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.libraryService.DeleteAuthor(r.Context(), id)
	if err != nil && err != sql.ErrNoRows {
		response.InternalServerError(w, r, err)
		return
	}

	if err == sql.ErrNoRows {
		response.NotFound(w, r, err)
		return
	}
}
