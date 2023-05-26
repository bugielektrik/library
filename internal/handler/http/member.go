package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"library/internal/dto"
	"library/internal/service"
)

type MemberHandler struct {
	memberService service.Member
}

func NewMemberHandler(a service.Member) *MemberHandler {
	return &MemberHandler{memberService: a}
}

func (h *MemberHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.add)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Put("/", h.update)
		r.Delete("/", h.delete)
		r.Get("/books", h.listBook)
	})

	return r
}

// List of members from the store
//
//	@Summary	List of members from the store
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Success	200			{array}		dto.MemberResponse
//	@Failure	500			{object}	dto.Response
//	@Router		/members 	[get]
func (h *MemberHandler) list(w http.ResponseWriter, r *http.Request) {
	res, err := h.memberService.List(r.Context())
	if err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}

	render.JSON(w, r, dto.OK(res))
}

// Add a new member to the store
//
//	@Summary	Add a new member to the store
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Param		request	body		dto.MemberRequest	true	"body param"
//	@Success	200		{object}	dto.MemberResponse
//	@Failure	400		{object}	dto.Response
//	@Failure	500		{object}	dto.Response
//	@Router		/members [post]
func (h *MemberHandler) add(w http.ResponseWriter, r *http.Request) {
	req := dto.MemberRequest{}
	if err := render.Bind(r, &req); err != nil {
		render.JSON(w, r, dto.BadRequest(err, req))
		return
	}

	res, err := h.memberService.Add(r.Context(), req)
	if err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}

	render.JSON(w, r, dto.OK(res))
}

// Read the member from the store
//
//	@Summary	Read the member from the store
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int	true	"path param"
//	@Success	200	{object}	dto.MemberResponse
//	@Failure	500	{object}	dto.Response
//	@Router		/members/{id} [get]
func (h *MemberHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.memberService.Get(r.Context(), id)
	if err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}

	render.JSON(w, r, dto.OK(res))
}

// Update the member in the store
//
//	@Summary	Update the member in the store
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Param		id		path	int					true	"path param"
//	@Param		request	body	dto.MemberRequest	true	"body param"
//	@Success	200
//	@Failure	400	{object}	dto.Response
//	@Failure	500	{object}	dto.Response
//	@Router		/members/{id} [put]
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

	render.NoContent(w, r)
}

// Delete the member from the store
//
//	@Summary	Delete the member from the store
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Param		id	path	int	true	"path param"
//	@Success	200
//	@Failure	500	{object}	dto.Response
//	@Router		/members/{id} [delete]
func (h *MemberHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.memberService.Delete(r.Context(), id); err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}

	render.NoContent(w, r)
}

// List of books from the store
//
//	@Summary	List of books from the store
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int	true	"path param"
//	@Success	200	{array}		dto.BookResponse
//	@Failure	500	{object}	dto.Response
//	@Router		/members/{id}/books [get]
func (h *MemberHandler) listBook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.memberService.ListBook(r.Context(), id)
	if err != nil {
		render.JSON(w, r, dto.InternalServerError(err))
		return
	}

	render.JSON(w, r, dto.OK(res))
}
