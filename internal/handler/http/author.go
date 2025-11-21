package http

import (
	"errors"
	"library-service/internal/service/interfaces"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"library-service/internal/domain/author"
	"library-service/pkg/server/response"
	"library-service/pkg/store"
)

type AuthorHandler struct {
	authorService interfaces.AuthorService
}

func NewAuthorHandler(s interfaces.AuthorService) *AuthorHandler {
	return &AuthorHandler{authorService: s}
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

// @Summary	list of authors from the repository
// @Tags		authors
// @Accept		json
// @Produce	json
// @Success	200			{array}		author.Response
// @Failure	500			{object}	response.Object
// @Router		/authors 	[get]
func (h *AuthorHandler) list(w http.ResponseWriter, r *http.Request) {
	res, err := h.authorService.ListAuthors(r.Context())
	if err != nil {
		response.InternalServerError(w, r, err, nil)
		return
	}

	response.OK(w, r, res)
}

// @Summary	add a new author to the repository
// @Tags		authors
// @Accept		json
// @Produce	json
// @Param		request	body		author.Request	true	"body param"
// @Success	200		{object}	author.Response
// @Failure	400		{object}	response.Object
// @Failure	500		{object}	response.Object
// @Router		/authors [post]
func (h *AuthorHandler) add(w http.ResponseWriter, r *http.Request) {
	req := author.Request{}
	if err := render.Bind(r, &req); err != nil {
		response.BadRequest(w, r, err, req)
		return
	}

	res, err := h.authorService.AddAuthor(r.Context(), req)
	if err != nil {
		response.InternalServerError(w, r, err, nil)
		return
	}

	response.OK(w, r, res)
}

// @Summary	get the author from the repository
// @Tags		authors
// @Accept		json
// @Produce	json
// @Param		id	path		int	true	"path param"
// @Success	200	{object}	author.Response
// @Failure	404	{object}	response.Object
// @Failure	500	{object}	response.Object
// @Router		/authors/{id} [get]
func (h *AuthorHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.authorService.GetAuthor(r.Context(), id)
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

// @Summary	update the author in the repository
// @Tags		authors
// @Accept		json
// @Produce	json
// @Param		id		path	int				true	"path param"
// @Param		request	body	author.Request	true	"body param"
// @Success	200
// @Failure	400	{object}	response.Object
// @Failure	404	{object}	response.Object
// @Failure	500	{object}	response.Object
// @Router		/authors/{id} [put]
func (h *AuthorHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := author.Request{}
	if err := render.Bind(r, &req); err != nil {
		response.BadRequest(w, r, err, req)
		return
	}

	if err := h.authorService.UpdateAuthor(r.Context(), id, req); err != nil {
		switch {
		case errors.Is(err, store.ErrorNotFound):
			response.NotFound(w, r, err)
		default:
			response.InternalServerError(w, r, err, nil)
		}
		return
	}
}

// @Summary	delete the author from the repository
// @Tags		authors
// @Accept		json
// @Produce	json
// @Param		id	path	string	true	"path param"
// @Success	200
// @Failure	404	{object}	response.Object
// @Failure	500	{object}	response.Object
// @Router		/authors/{id} [delete]
func (h *AuthorHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.authorService.DeleteAuthor(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, store.ErrorNotFound):
			response.NotFound(w, r, err)
		default:
			response.InternalServerError(w, r, err, nil)
		}
		return
	}
}
