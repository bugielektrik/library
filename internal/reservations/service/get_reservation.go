package service

import (
	"context"
	"library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"

	"go.uber.org/zap"

	reservationdomain "library-service/internal/reservations/domain"
)

// GetReservationRequest represents the input for getting a reservation
type GetReservationRequest struct {
	ReservationID string
}

// GetReservationResponse represents the output of getting a reservation
type GetReservationResponse struct {
	reservationdomain.Response
}

// GetReservationUseCase handles retrieving a reservation by ID
type GetReservationUseCase struct {
	reservationRepo reservationdomain.Repository
}

// NewGetReservationUseCase creates a new instance of GetReservationUseCase
func NewGetReservationUseCase(reservationRepo reservationdomain.Repository) *GetReservationUseCase {
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
		Response: reservationdomain.ParseFromReservation(reservationEntity),
	}, nil
}
