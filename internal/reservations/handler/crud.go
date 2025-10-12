package http

import (
	"library-service/internal/pkg/httputil"
	"library-service/internal/pkg/logutil"
	"net/http"

	"go.uber.org/zap"

	reservationops "library-service/internal/reservations/service"
)

// This file contains CRUD operations for reservations.

// @Summary Create a new reservation
// @Description Reserve a book for the authenticated member
// @Tags reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateReservationRequest true "Reservation data"
// @Success 201 {object} ReservationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reservations [post]
func (h *ReservationHandler) create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "reservation_handler", "create")

	// Get member ID from context
	memberID, ok := h.GetMemberID(w, r)
	if !ok {
		return
	}

	// Decode request
	var req CreateReservationRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Validate request
	if !h.validator.ValidateStruct(w, req) {
		return
	}

	// Execute usecase
	result, err := h.useCases.Reservation.CreateReservation.Execute(ctx, reservationops.CreateReservationRequest{
		BookID:   req.BookID,
		MemberID: memberID,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTO
	response := FromReservationResponse(result.Response)

	logger.Info("reservation created", zap.String("id", response.ID))
	h.RespondJSON(w, http.StatusCreated, response)
}

// @Summary Get a reservation by ID
// @Description Get details of a specific reservation
// @Tags reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Reservation ID"
// @Success 200 {object} ReservationResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reservations/{id} [get]
func (h *ReservationHandler) get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "reservation_handler", "get")

	id, ok := h.GetURLParam(w, r, "id")
	if !ok {
		return
	}

	// Execute usecase
	result, err := h.useCases.Reservation.GetReservation.Execute(ctx, reservationops.GetReservationRequest{
		ReservationID: id,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTO
	response := FromReservationResponse(result.Response)

	logger.Debug("reservation retrieved", zap.String("id", id))
	h.RespondJSON(w, http.StatusOK, response)
}

// @Summary Cancel a reservation
// @Description Cancel a reservation (only the owner can cancel)
// @Tags reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Reservation ID"
// @Success 200 {object} ReservationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reservations/{id} [delete]
func (h *ReservationHandler) cancel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "reservation_handler", "cancel")

	// Get member ID from context
	memberID, ok := h.GetMemberID(w, r)
	if !ok {
		return
	}

	id, ok := h.GetURLParam(w, r, "id")
	if !ok {
		return
	}

	// Execute usecase
	result, err := h.useCases.Reservation.CancelReservation.Execute(ctx, reservationops.CancelReservationRequest{
		ReservationID: id,
		MemberID:      memberID,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTO
	response := FromReservationResponse(result.Response)

	logger.Info("reservation cancelled", zap.String("id", id))
	h.RespondJSON(w, http.StatusOK, response)
}
