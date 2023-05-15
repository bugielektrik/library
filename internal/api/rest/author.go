package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"library/internal/dto"
	"library/internal/service"
)

// AuthorHandler represents the author handler
type AuthorHandler struct {
	authorService service.AuthorService
}

// AuthorRoutes creates a new instance of the router
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

// create creates a new author
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

// getByID retrieves an author by ID
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

// getAll retrieves all authors
func (h *AuthorHandler) getAll(w http.ResponseWriter, r *http.Request) {
	res, err := h.authorService.GetAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// update updates an existing author
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

// delete deletes a author
func (h *AuthorHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.authorService.Delete(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
