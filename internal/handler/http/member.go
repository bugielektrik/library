package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"library/internal/domain/member"
	"library/internal/service/subscription"
	"library/pkg/server/status"
)

type MemberHandler struct {
	subscriptionService *subscription.Service
}

func NewMemberHandler(s *subscription.Service) *MemberHandler {
	return &MemberHandler{subscriptionService: s}
}

func (h *MemberHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.add)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Put("/", h.update)
		r.Delete("/", h.delete)
		r.Get("/books", h.listBooks)
	})

	return r
}

// List of members from the database
//
//	@Summary	List of members from the database
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Success	200			{array}		member.Response
//	@Failure	500			{object}	status.Response
//	@Router		/members 	[get]
func (h *MemberHandler) list(w http.ResponseWriter, r *http.Request) {
	res, err := h.subscriptionService.ListMembers(r.Context())
	if err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}

	render.JSON(w, r, status.OK(res))
}

// Add a new member to the database
//
//	@Summary	Add a new member to the database
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Param		request	body		member.Request	true	"body param"
//	@Success	200		{object}	member.Response
//	@Failure	400		{object}	status.Response
//	@Failure	500		{object}	status.Response
//	@Router		/members [post]
func (h *MemberHandler) add(w http.ResponseWriter, r *http.Request) {
	req := member.Request{}
	if err := render.Bind(r, &req); err != nil {
		render.JSON(w, r, status.BadRequest(err, req))
		return
	}

	res, err := h.subscriptionService.AddMember(r.Context(), req)
	if err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}

	render.JSON(w, r, status.OK(res))
}

// Read the member from the database
//
//	@Summary	Read the member from the database
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int	true	"path param"
//	@Success	200	{object}	member.Response
//	@Failure	500	{object}	status.Response
//	@Router		/members/{id} [get]
func (h *MemberHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.subscriptionService.GetMember(r.Context(), id)
	if err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}

	render.JSON(w, r, status.OK(res))
}

// Update the member in the database
//
//	@Summary	Update the member in the database
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Param		id		path	int				true	"path param"
//	@Param		request	body	member.Request	true	"body param"
//	@Success	200
//	@Failure	400	{object}	status.Response
//	@Failure	500	{object}	status.Response
//	@Router		/members/{id} [put]
func (h *MemberHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := member.Request{}
	if err := render.Bind(r, &req); err != nil {
		render.JSON(w, r, status.BadRequest(err, req))
		return
	}

	if err := h.subscriptionService.UpdateMember(r.Context(), id, req); err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}
}

// Delete the member from the database
//
//	@Summary	Delete the member from the database
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Param		id	path	int	true	"path param"
//	@Success	200
//	@Failure	500	{object}	status.Response
//	@Router		/members/{id} [delete]
func (h *MemberHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.subscriptionService.DeleteMember(r.Context(), id); err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}
}

// List of books from the database
//
//	@Summary	List of books from the database
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int	true	"path param"
//	@Success	200	{array}		book.Response
//	@Failure	500	{object}	status.Response
//	@Router		/members/{id}/books [get]
func (h *MemberHandler) listBooks(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.subscriptionService.ListMemberBooks(r.Context(), id)
	if err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}

	render.JSON(w, r, status.OK(res))
}
