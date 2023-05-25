package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"library/internal/dto"
	"library/internal/service"
)

type BookHandler struct {
	bookService service.BookService
}

func NewBookHandler(a service.BookService) *BookHandler {
	return &BookHandler{bookService: a}
}

func (h *BookHandler) Routes() chi.Router {
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

func (h *BookHandler) list(w http.ResponseWriter, r *http.Request) {
	res, err := h.bookService.List(r.Context())
	if err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}
	render.JSON(w, r, dto.OK(res))
}

func (h *BookHandler) create(w http.ResponseWriter, r *http.Request) {
	req := dto.BookRequest{}
	if err := render.Bind(r, &req); err != nil {
		render.JSON(w, r, dto.BadRequest(err, req))
		return
	}

	res, err := h.bookService.Create(r.Context(), req)
	if err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}

	render.JSON(w, r, dto.Created(res))
}

func (h *BookHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.bookService.Get(r.Context(), id)
	if err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}

	render.JSON(w, r, dto.OK(res))
}

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
}

func (h *BookHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.bookService.Delete(r.Context(), id); err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}
}
