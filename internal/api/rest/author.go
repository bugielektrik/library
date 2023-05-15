package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"library/internal/dto"
	"library/internal/service"
)

type AuthorHandler struct {
	authorService service.AuthorService
}

func AuthorRoutes(s service.AuthorService) chi.Router {
	handler := AuthorHandler{
		authorService: s,
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

func (h *AuthorHandler) create(w http.ResponseWriter, r *http.Request) {
	req := dto.AuthorRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(req)
		return
	}

	res, err := h.authorService.Create(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *AuthorHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.authorService.GetByID(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *AuthorHandler) getAll(w http.ResponseWriter, r *http.Request) {
	res, err := h.authorService.GetAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *AuthorHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := dto.AuthorRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.authorService.Update(id, req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *AuthorHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.authorService.Delete(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
