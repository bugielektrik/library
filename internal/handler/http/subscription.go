package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"library/internal/domain/member"
	"library/internal/service/subscription"
	"library/pkg/server/status"
)

type SubscriptionHandler struct {
	subscriptionService *subscription.Service
}

func NewSubscriptionHandler(s *subscription.Service) *SubscriptionHandler {
	return &SubscriptionHandler{subscriptionService: s}
}

func (h *SubscriptionHandler) MemberRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.listMembers)
	r.Post("/", h.addMember)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.getMember)
		r.Put("/", h.updateMember)
		r.Delete("/", h.deleteMember)
		r.Get("/books", h.listMemberBooks)
	})

	return r
}

// List of members from the store
//
//	@Summary	List of members from the store
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Success	200			{array}		member.Response
//	@Failure	500			{object}	status.Response
//	@Router		/members 	[get]
func (h *SubscriptionHandler) listMembers(w http.ResponseWriter, r *http.Request) {
	res, err := h.subscriptionService.ListMembers(r.Context())
	if err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}

	render.JSON(w, r, status.OK(res))
}

// Add a new member to the store
//
//	@Summary	Add a new member to the store
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Param		request	body		member.Request	true	"body param"
//	@Success	200		{object}	member.Response
//	@Failure	400		{object}	status.Response
//	@Failure	500		{object}	status.Response
//	@Router		/members [post]
func (h *SubscriptionHandler) addMember(w http.ResponseWriter, r *http.Request) {
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

// Read the member from the store
//
//	@Summary	Read the member from the store
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int	true	"path param"
//	@Success	200	{object}	member.Response
//	@Failure	500	{object}	status.Response
//	@Router		/members/{id} [get]
func (h *SubscriptionHandler) getMember(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.subscriptionService.GetMember(r.Context(), id)
	if err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}

	render.JSON(w, r, status.OK(res))
}

// Update the member in the store
//
//	@Summary	Update the member in the store
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Param		id		path	int				true	"path param"
//	@Param		request	body	member.Request	true	"body param"
//	@Success	200
//	@Failure	400	{object}	status.Response
//	@Failure	500	{object}	status.Response
//	@Router		/members/{id} [put]
func (h *SubscriptionHandler) updateMember(w http.ResponseWriter, r *http.Request) {
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

// Delete the member from the store
//
//	@Summary	Delete the member from the store
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Param		id	path	int	true	"path param"
//	@Success	200
//	@Failure	500	{object}	status.Response
//	@Router		/members/{id} [delete]
func (h *SubscriptionHandler) deleteMember(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.subscriptionService.DeleteMember(r.Context(), id); err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}
}

// List of books from the store
//
//	@Summary	List of books from the store
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int	true	"path param"
//	@Success	200	{array}		book.Response
//	@Failure	500	{object}	status.Response
//	@Router		/members/{id}/books [get]
func (h *SubscriptionHandler) listMemberBooks(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.subscriptionService.ListMemberBooks(r.Context(), id)
	if err != nil {
		render.JSON(w, r, status.InternalServerError(err))
		return
	}

	render.JSON(w, r, status.OK(res))
}
