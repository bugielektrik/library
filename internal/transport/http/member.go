package http

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

func NewMemberHandler(m service.MemberService) *MemberHandler {
	return &MemberHandler{memberService: m}
}

func (h *MemberHandler) Routes() chi.Router {
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

func (h *MemberHandler) list(w http.ResponseWriter, r *http.Request) {
	res, err := h.memberService.List(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *MemberHandler) create(w http.ResponseWriter, r *http.Request) {
	req := dto.MemberRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(req)
		return
	}

	res, err := h.memberService.Create(r.Context(), req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *MemberHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.memberService.Get(r.Context(), id)
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.memberService.Update(r.Context(), id, req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *MemberHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.memberService.Delete(r.Context(), id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
