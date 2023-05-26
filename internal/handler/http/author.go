package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"library/internal/dto"
	"library/internal/service"
)

type AuthorHandler struct {
	authorService service.Author
}

func NewAuthorHandler(a service.Author) *AuthorHandler {
	return &AuthorHandler{authorService: a}
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

// List of authors from the store
//
//	@Summary	List of authors from the store
//	@Tags		authors
//	@Accept		json
//	@Produce	json
//	@Success	200			{array}		dto.AuthorResponse
//	@Failure	500			{object}	dto.Response
//	@Router		/authors 	[get]
func (h *AuthorHandler) list(w http.ResponseWriter, r *http.Request) {
	res, err := h.authorService.List(r.Context())
	if err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}

	render.JSON(w, r, dto.OK(res))
}

// Add a new author to the store
//
//	@Summary	Add a new author to the store
//	@Tags		authors
//	@Accept		json
//	@Produce	json
//	@Param		request	body		dto.AuthorRequest	true	"body param"
//	@Success	200		{object}	dto.AuthorResponse
//	@Failure	400		{object}	dto.Response
//	@Failure	500		{object}	dto.Response
//	@Router		/authors [post]
func (h *AuthorHandler) add(w http.ResponseWriter, r *http.Request) {
	req := dto.AuthorRequest{}
	if err := render.Bind(r, &req); err != nil {
		render.JSON(w, r, dto.BadRequest(err, req))
		return
	}

	res, err := h.authorService.Add(r.Context(), req)
	if err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}

	render.JSON(w, r, dto.OK(res))
}

// Read the author from the store
//
//	@Summary	Read the author from the store
//	@Tags		authors
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int	true	"path param"
//	@Success	200	{object}	dto.AuthorResponse
//	@Failure	500	{object}	dto.Response
//	@Router		/authors/{id} [get]
func (h *AuthorHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.authorService.Get(r.Context(), id)
	if err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}

	render.JSON(w, r, dto.OK(res))
}

// Update the author in the store
//
//	@Summary	Update the author in the store
//	@Tags		authors
//	@Accept		json
//	@Produce	json
//	@Param		id		path	int					true	"path param"
//	@Param		request	body	dto.AuthorRequest	true	"body param"
//	@Success	200
//	@Failure	400	{object}	dto.Response
//	@Failure	500	{object}	dto.Response
//	@Router		/authors/{id} [put]
func (h *AuthorHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := dto.AuthorRequest{}
	if err := render.Bind(r, &req); err != nil {
		render.JSON(w, r, dto.BadRequest(err, req))
		return
	}

	if err := h.authorService.Update(r.Context(), id, req); err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
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
//	@Failure	500	{object}	dto.Response
//	@Router		/authors/{id} [delete]
func (h *AuthorHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.authorService.Delete(r.Context(), id); err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}
}
