package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"library/internal/domain/author"
	"library/internal/service/library"
	"library/pkg/server/status"
	"library/pkg/store"
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
		r.Get("/", h.get)
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
//	@Success	200			{array}		author.Response
//	@Failure	500			{object}	status.Response
//	@Router		/authors 	[get]
func (h *AuthorHandler) list(w http.ResponseWriter, r *http.Request) {
	res, err := h.libraryService.ListAuthors(r.Context())
	if err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}

	render.JSON(w, r, status.OK(res))
}

// Add a new author to the database
//
//	@Summary	Add a new author to the database
//	@Tags		authors
//	@Accept		json
//	@Produce	json
//	@Param		request	body		author.Request	true	"body param"
//	@Success	200		{object}	author.Response
//	@Failure	400		{object}	status.Response
//	@Failure	500		{object}	status.Response
//	@Router		/authors [post]
func (h *AuthorHandler) add(w http.ResponseWriter, r *http.Request) {
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

// Read the author from the database
//
//	@Summary	Read the author from the database
//	@Tags		authors
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int	true	"path param"
//	@Success	200	{object}	author.Response
//	@Failure	404	{object}	status.Response
//	@Failure	500	{object}	status.Response
//	@Router		/authors/{id} [get]
func (h *AuthorHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.libraryService.GetAuthor(r.Context(), id)
	if err != nil && err != store.ErrorNotFound {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}

	if err == store.ErrorNotFound {
		render.JSON(w, r, status.NotFound(err))
		return
	}

	render.JSON(w, r, status.OK(res))
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
//	@Failure	400	{object}	status.Response
//	@Failure	404	{object}	status.Response
//	@Failure	500	{object}	status.Response
//	@Router		/authors/{id} [put]
func (h *AuthorHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := author.Request{}
	if err := render.Bind(r, &req); err != nil {
		render.JSON(w, r, status.BadRequest(err, req))
		return
	}

	err := h.libraryService.UpdateAuthor(r.Context(), id, req)
	if err != nil && err != store.ErrorNotFound {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}

	if err == store.ErrorNotFound {
		render.JSON(w, r, status.NotFound(err))
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
//	@Failure	404	{object}	status.Response
//	@Failure	500	{object}	status.Response
//	@Router		/authors/{id} [delete]
func (h *AuthorHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.libraryService.DeleteAuthor(r.Context(), id)
	if err != nil && err != store.ErrorNotFound {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}

	if err == store.ErrorNotFound {
		render.JSON(w, r, status.NotFound(err))
		return
	}
}
