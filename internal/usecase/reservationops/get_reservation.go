package reservationops

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/domain/reservation"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
)

// GetReservationRequest represents the input for getting a reservation
type GetReservationRequest struct {
	ReservationID string
}

// GetReservationResponse represents the output of getting a reservation
type GetReservationResponse struct {
	reservation.Response
}

// GetReservationUseCase handles retrieving a reservation by ID
type GetReservationUseCase struct {
	reservationRepo reservation.Repository
}

// NewGetReservationUseCase creates a new instance of GetReservationUseCase
func NewGetReservationUseCase(reservationRepo reservation.Repository) *GetReservationUseCase {
	return &GetReservationUseCase{
		reservationRepo: reservationRepo,
	}
}

// Execute retrieves a reservation by ID
func (uc *GetReservationUseCase) Execute(ctx context.Context, req GetReservationRequest) (GetReservationResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "reservation", "get")

	// Get the reservation from repository
	reservationEntity, err := uc.reservationRepo.GetByID(ctx, req.ReservationID)
	if err != nil {
		logger.Error("failed to get reservation", zap.Error(err))
		return GetReservationResponse{}, errors.ErrNotFound.WithDetails("resource", "reservation")
	}

	logger.Info("reservation retrieved successfully")

	return GetReservationResponse{
		Response: reservation.ParseFromReservation(reservationEntity),
	}, nil
}
