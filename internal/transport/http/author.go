package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"library/internal/dto"
	"library/internal/service"
)

type AuthorHandler struct {
	authorService service.AuthorService
}

func NewAuthorHandler(a service.AuthorService) *AuthorHandler {
	return &AuthorHandler{authorService: a}
}

func (h *AuthorHandler) Routes() chi.Router {

	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.create)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Put("/", h.update)
		r.Delete("/", h.delete)
	})

	return r
}

func (h *AuthorHandler) list(w http.ResponseWriter, r *http.Request) {
	res, err := h.authorService.List(r.Context())
	if err != nil {
		render.Render(w, r, dto.Data{
			Status:  http.StatusInternalServerError,
			Success: false,
			Data:    err,
		})
		return
	}

	render.Render(w, r, dto.Data{
		Status:  http.StatusOK,
		Success: true,
		Data:    res,
	})
}

func (h *AuthorHandler) create(w http.ResponseWriter, r *http.Request) {
	req := dto.AuthorRequest{}
	if err := render.Bind(r, &req); err != nil {
		render.Render(w, r, dto.Data{
			Status:  http.StatusBadRequest,
			Success: false,
			Data:    req,
		})
		return
	}

	res, err := h.authorService.Create(r.Context(), req)
	if err != nil {
		render.Render(w, r, dto.Data{
			Status:  http.StatusInternalServerError,
			Success: false,
			Data:    err,
		})
		return
	}

	render.Render(w, r, dto.Data{
		Status:  http.StatusCreated,
		Success: true,
		Data:    res,
	})
}

func (h *AuthorHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.authorService.Get(r.Context(), id)
	if err != nil {
		render.Render(w, r, dto.Data{
			Status:  http.StatusInternalServerError,
			Success: false,
			Data:    err,
		})
		return
	}

	render.Render(w, r, dto.Data{
		Status:  http.StatusOK,
		Success: true,
		Data:    res,
	})
}

func (h *AuthorHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := dto.AuthorRequest{}
	if err := render.Bind(r, &req); err != nil {
		render.Render(w, r, dto.Data{
			Status:  http.StatusBadRequest,
			Success: false,
			Data:    req,
		})
		return
	}

	if err := h.authorService.Update(r.Context(), id, req); err != nil {
		render.Render(w, r, dto.Data{
			Status:  http.StatusInternalServerError,
			Success: false,
			Data:    err,
		})
		return
	}
}

func (h *AuthorHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.authorService.Delete(r.Context(), id); err != nil {
		render.Render(w, r, dto.Data{
			Status:  http.StatusInternalServerError,
			Success: false,
			Data:    err,
		})
		return
	}
}
