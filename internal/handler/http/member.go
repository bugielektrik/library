package http

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"library-service/internal/domain/member"
	"library-service/internal/service/subscription"
	"library-service/pkg/server/response"
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
	r.Post("/", h.create)

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
//	@Success	200			{array}		response.Object
//	@Failure	500			{object}	response.Object
//	@Router		/members 	[get]
func (h *MemberHandler) list(w http.ResponseWriter, r *http.Request) {
	res, err := h.subscriptionService.ListMembers(r.Context())
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.OK(w, r, res)
}

// Add a new member to the database
//
//	@Summary	Add a new member to the database
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Param		request	body		member.Request	true	"body param"
//	@Success	200		{object}	response.Object
//	@Failure	400		{object}	response.Object
//	@Failure	500		{object}	response.Object
//	@Router		/members [post]
func (h *MemberHandler) create(w http.ResponseWriter, r *http.Request) {
	req := member.Request{}
	if err := render.Bind(r, &req); err != nil {
		response.BadRequest(w, r, err, req)
		return
	}

	res, err := h.subscriptionService.CreateMember(r.Context(), req)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.OK(w, r, res)
}

// Read the member from the database
//
//	@Summary	Read the member from the database
//	@Tags		members
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int	true	"path param"
//	@Success	200	{object}	response.Object
//	@Failure	404	{object}	response.Object
//	@Failure	500	{object}	response.Object
//	@Router		/members/{id} [get]
func (h *MemberHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.subscriptionService.GetMember(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			response.NotFound(w, r, err)
		default:
			response.InternalServerError(w, r, err)
		}
		return
	}

	response.OK(w, r, res)
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
//	@Failure	400	{object}	response.Object
//	@Failure	404	{object}	response.Object
//	@Failure	500	{object}	response.Object
//	@Router		/members/{id} [put]
func (h *MemberHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := member.Request{}
	if err := render.Bind(r, &req); err != nil {
		response.BadRequest(w, r, err, req)
		return
	}

	if err := h.subscriptionService.UpdateMember(r.Context(), id, req); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			response.NotFound(w, r, err)
		default:
			response.InternalServerError(w, r, err)
		}
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
//	@Failure	404	{object}	response.Object
//	@Failure	500	{object}	response.Object
//	@Router		/members/{id} [delete]
func (h *MemberHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.subscriptionService.DeleteMember(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			response.NotFound(w, r, err)
		default:
			response.InternalServerError(w, r, err)
		}
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
//	@Success	200	{array}		response.Object
//	@Failure	404	{object}	response.Object
//	@Failure	500	{object}	response.Object
//	@Router		/members/{id}/books [get]
func (h *MemberHandler) listBooks(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	res, err := h.subscriptionService.ListMemberBooks(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			response.NotFound(w, r, err)
		default:
			response.InternalServerError(w, r, err)
		}
		return
	}

	response.OK(w, r, res)
}
