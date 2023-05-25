package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"library/internal/dto"
	"library/internal/service"
)

type MemberHandler struct {
	memberService service.MemberService
}

func NewMemberHandler(a service.MemberService) *MemberHandler {
	return &MemberHandler{memberService: a}
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
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}
	render.JSON(w, r, dto.OK(res))
}

func (h *MemberHandler) create(w http.ResponseWriter, r *http.Request) {
	req := dto.MemberRequest{}
	if err := render.Bind(r, &req); err != nil {
		render.JSON(w, r, dto.BadRequest(err, req))
		return
	}

	res, err := h.memberService.Create(r.Context(), req)
	if err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}

	render.JSON(w, r, dto.Created(res))
}

func (h *MemberHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.memberService.Get(r.Context(), id)
	if err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}

	render.JSON(w, r, dto.OK(res))
}

func (h *MemberHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := dto.MemberRequest{}
	if err := render.Bind(r, &req); err != nil {
		render.JSON(w, r, dto.BadRequest(err, req))
		return
	}

	if err := h.memberService.Update(r.Context(), id, req); err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}
}

func (h *MemberHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.memberService.Delete(r.Context(), id); err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}
}
