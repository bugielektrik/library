package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"library/internal/dto"
	"library/internal/service"
)

type MemberHandler struct {
	memberService service.MemberService
}

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

func (h *MemberHandler) create(w http.ResponseWriter, r *http.Request) {
	req := dto.MemberRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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

func (h *MemberHandler) getAll(w http.ResponseWriter, r *http.Request) {
	res, err := h.memberService.GetAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *MemberHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := dto.MemberRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = h.memberService.Update(id, req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *MemberHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.memberService.Delete(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
