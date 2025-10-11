package member

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"library-service/internal/adapters/http/dto"
	"library-service/internal/adapters/http/handlers"
	"library-service/internal/usecase"
	"library-service/internal/usecase/memberops"
	"library-service/pkg/logutil"
)

// MemberHandler handles HTTP requests for members
type MemberHandler struct {
	handlers.BaseHandler
	useCases *usecase.Container
}

// NewMemberHandler creates a new member handler
func NewMemberHandler(
	useCases *usecase.Container,
) *MemberHandler {
	return &MemberHandler{
		useCases: useCases,
	}
}

// Routes returns the router for member endpoints
func (h *MemberHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.getProfile)
	})

	return r
}

// @Summary List all members
// @Description Retrieves a list of all members in the system
// @Tags members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.MemberResponse "List of members"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /members [get]
func (h *MemberHandler) list(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "member_handler", "list")

	// Execute usecase
	result, err := h.useCases.Member.ListMembers.Execute(ctx, memberops.ListMembersRequest{})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTOs
	members := dto.FromMemberEntities(result.Members)

	logger.Info("members listed", zap.Int("count", len(members)))
	h.RespondJSON(w, http.StatusOK, members)
}

// @Summary Get member profile
// @Description Retrieves detailed profile information for a specific member
// @Tags members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Member ID"
// @Success 200 {object} dto.MemberResponse "Member profile"
// @Failure 400 {object} dto.ErrorResponse "Invalid member ID"
// @Failure 404 {object} dto.ErrorResponse "Member not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /members/{id} [get]
func (h *MemberHandler) getProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "member_handler", "get_profile")

	id, ok := h.GetURLParam(w, r, "id")
	if !ok {
		return
	}

	// Execute usecase
	result, err := h.useCases.Member.GetMemberProfile.Execute(ctx, memberops.GetMemberProfileRequest{
		MemberID: id,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTO
	memberResp := dto.FromMemberEntity(result.Member)

	logger.Info("member profile retrieved", zap.String("id", id))
	h.RespondJSON(w, http.StatusOK, memberResp)
}
