package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"library/internal/dto"
	"library/internal/service"
)

type BookHandler struct {
	bookService service.BookService
}

func BookRoutes(s service.BookService) chi.Router {
	handler := BookHandler{
		bookService: s,
	}

	r := chi.NewRouter()

	r.Get("/", handler.getAll)
	r.Post("/", handler.create)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", handler.getByID)
		r.Put("/", handler.update)
		r.Delete("/", handler.delete)
	})

	return r
}

func (h *BookHandler) create(w http.ResponseWriter, r *http.Request) {
	req := dto.BookRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(req)
		return
	}

	res, err := h.bookService.Create(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *BookHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.bookService.GetByID(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *BookHandler) getAll(w http.ResponseWriter, r *http.Request) {
	res, err := h.bookService.GetAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *BookHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := dto.BookRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.bookService.Update(id, req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *BookHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.bookService.Delete(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
