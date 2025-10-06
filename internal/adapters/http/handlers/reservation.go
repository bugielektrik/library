package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"library-service/internal/adapters/http/dto"
	"library-service/internal/adapters/http/middleware"
	"library-service/internal/usecase/reservationops"
	"library-service/pkg/errors"

	"library-service/internal/infrastructure/log"
)

// ReservationHandler handles HTTP requests for reservations
type ReservationHandler struct {
	createReservationUC      *reservationops.CreateReservationUseCase
	cancelReservationUC      *reservationops.CancelReservationUseCase
	getReservationUC         *reservationops.GetReservationUseCase
	listMemberReservationsUC *reservationops.ListMemberReservationsUseCase
	validator                *middleware.Validator
}

// NewReservationHandler creates a new reservation handler
func NewReservationHandler(
	createReservationUC *reservationops.CreateReservationUseCase,
	cancelReservationUC *reservationops.CancelReservationUseCase,
	getReservationUC *reservationops.GetReservationUseCase,
	listMemberReservationsUC *reservationops.ListMemberReservationsUseCase,
) *ReservationHandler {
	return &ReservationHandler{
		createReservationUC:      createReservationUC,
		cancelReservationUC:      cancelReservationUC,
		getReservationUC:         getReservationUC,
		listMemberReservationsUC: listMemberReservationsUC,
		validator:                middleware.NewValidator(),
	}
}

// Routes returns the router for reservation endpoints
func (h *ReservationHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.listMyReservations)
	r.Post("/", h.create)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Delete("/", h.cancel)
	})

	return r
}

// @Summary List my reservations
// @Description Get all reservations for the authenticated member
// @Tags reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.ReservationResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /reservations [get]
func (h *ReservationHandler) listMyReservations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx).Named("reservation_handler.list_my_reservations")

	// Get member ID from context (set by auth middleware)
	memberID, ok := middleware.GetMemberIDFromContext(ctx)
	if !ok {
		h.respondError(w, r, errors.ErrUnauthorized)
		return
	}

	// Execute usecase
	result, err := h.listMemberReservationsUC.Execute(ctx, reservationops.ListMemberReservationsRequest{
		MemberID: memberID,
	})
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	// Convert to DTOs
	reservations := dto.FromReservationResponses(result.Reservations)

	logger.Info("reservations listed", zap.Int("count", len(reservations)))
	h.respondJSON(w, http.StatusOK, reservations)
}

// @Summary Create a new reservation
// @Description Reserve a book for the authenticated member
// @Tags reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateReservationRequest true "Reservation data"
// @Success 201 {object} dto.ReservationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /reservations [post]
func (h *ReservationHandler) create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx).Named("reservation_handler.create")

	// Get member ID from context (set by auth middleware)
	memberID, ok := middleware.GetMemberIDFromContext(ctx)
	if !ok {
		h.respondError(w, r, errors.ErrUnauthorized)
		return
	}

	// Decode request
	var req dto.CreateReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, r, errors.ErrInvalidInput.Wrap(err))
		return
	}

	// Validate request
	if !h.validator.ValidateStruct(w, req) {
		return
	}

	// Execute usecase
	result, err := h.createReservationUC.Execute(ctx, reservationops.CreateReservationRequest{
		BookID:   req.BookID,
		MemberID: memberID,
	})
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	// Convert to DTO
	response := dto.FromReservationResponse(result.Response)

	logger.Info("reservation created", zap.String("id", response.ID))
	h.respondJSON(w, http.StatusCreated, response)
}

// @Summary Get a reservation by ID
// @Description Get details of a specific reservation
// @Tags reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Reservation ID"
// @Success 200 {object} dto.ReservationResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /reservations/{id} [get]
func (h *ReservationHandler) get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx).Named("reservation_handler.get")

	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondError(w, r, errors.ErrInvalidInput.WithDetails("field", "id"))
		return
	}

	// Execute usecase
	result, err := h.getReservationUC.Execute(ctx, reservationops.GetReservationRequest{
		ReservationID: id,
	})
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	// Convert to DTO
	response := dto.FromReservationResponse(result.Response)

	logger.Debug("reservation retrieved", zap.String("id", id))
	h.respondJSON(w, http.StatusOK, response)
}

// @Summary Cancel a reservation
// @Description Cancel a reservation (only the owner can cancel)
// @Tags reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Reservation ID"
// @Success 200 {object} dto.ReservationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /reservations/{id} [delete]
func (h *ReservationHandler) cancel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx).Named("reservation_handler.cancel")

	// Get member ID from context (set by auth middleware)
	memberID, ok := middleware.GetMemberIDFromContext(ctx)
	if !ok {
		h.respondError(w, r, errors.ErrUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondError(w, r, errors.ErrInvalidInput.WithDetails("field", "id"))
		return
	}

	// Execute usecase
	result, err := h.cancelReservationUC.Execute(ctx, reservationops.CancelReservationRequest{
		ReservationID: id,
		MemberID:      memberID,
	})
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	// Convert to DTO
	response := dto.FromReservationResponse(result.Response)

	logger.Info("reservation cancelled", zap.String("id", id))
	h.respondJSON(w, http.StatusOK, response)
}

// respondJSON writes a JSON response
func (h *ReservationHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError writes an error response
func (h *ReservationHandler) respondError(w http.ResponseWriter, r *http.Request, err error) {
	logger := log.FromContext(r.Context())

	status := errors.GetHTTPStatus(err)

	if status >= 500 {
		logger.Error("internal error", zap.Error(err))
	} else {
		logger.Warn("client error", zap.Error(err), zap.Int("status", status))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := dto.FromError(err)
	json.NewEncoder(w).Encode(response)
}
