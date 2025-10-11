package reservation

import (
	"net/http"

	"go.uber.org/zap"

	"library-service/internal/adapters/http/dto"
	"library-service/internal/usecase/reservationops"
	"library-service/pkg/logutil"
)

// This file contains query operations for reservations.

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
	logger := logutil.HandlerLogger(ctx, "reservation_handler", "list_my_reservations")

	// Get member ID from context
	memberID, ok := h.GetMemberID(w, r)
	if !ok {
		return
	}

	// Execute usecase
	result, err := h.useCases.Reservation.ListMemberReservations.Execute(ctx, reservationops.ListMemberReservationsRequest{
		MemberID: memberID,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTOs
	reservations := dto.FromReservationResponses(result.Reservations)

	logger.Info("reservations listed", zap.Int("count", len(reservations)))
	h.RespondJSON(w, http.StatusOK, reservations)
}
