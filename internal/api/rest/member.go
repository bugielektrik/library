package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"library/internal/dto"
	"library/internal/service"
)

// MemberHandler represents the member handler
type MemberHandler struct {
	memberService service.MemberService
}

// MemberRoutes creates a new instance of the router
func MemberRoutes(s service.MemberService) chi.Router {
	handler := MemberHandler{
		memberService: s,
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

// create creates a new member
func (h *MemberHandler) create(w http.ResponseWriter, r *http.Request) {
	req := dto.MemberRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(req)
		return
	}

	res, err := h.memberService.Create(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

// getByID retrieves an member by ID
func (h *MemberHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.memberService.GetByID(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// getAll retrieves all members
func (h *MemberHandler) getAll(w http.ResponseWriter, r *http.Request) {
	res, err := h.memberService.GetAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// update updates an existing member
func (h *MemberHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := dto.MemberRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.memberService.Update(id, req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// delete deletes a member
func (h *MemberHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.memberService.Delete(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
