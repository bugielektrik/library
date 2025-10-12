package service

import (
	"context"
	errors2 "library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"

	"go.uber.org/zap"

	reservationdomain "library-service/internal/reservations/domain"
)

// CancelReservationRequest represents the input for cancelling a reservation
type CancelReservationRequest struct {
	ReservationID string
	MemberID      string // To verify ownership
}

// CancelReservationResponse represents the output of cancelling a reservation
type CancelReservationResponse struct {
	reservationdomain.Response
}

// CancelReservationUseCase handles cancelling a reservation
type CancelReservationUseCase struct {
	reservationRepo    reservationdomain.Repository
	reservationService *reservationdomain.Service
}

// NewCancelReservationUseCase creates a new instance of CancelReservationUseCase
func NewCancelReservationUseCase(
	reservationRepo reservationdomain.Repository,
	reservationService *reservationdomain.Service,
) *CancelReservationUseCase {
	return &CancelReservationUseCase{
		reservationRepo:    reservationRepo,
		reservationService: reservationService,
	}
}

// Execute cancels a reservation
func (uc *CancelReservationUseCase) Execute(ctx context.Context, req CancelReservationRequest) (CancelReservationResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "reservation", "cancel")

	// Get the reservation
	reservationEntity, err := uc.reservationRepo.GetByID(ctx, req.ReservationID)
	if err != nil {
		logger.Error("failed to get reservation", zap.Error(err))
		return CancelReservationResponse{}, errors2.ErrNotFound.WithDetails("resource", "reservation")
	}

	// Verify ownership
	if reservationEntity.MemberID != req.MemberID {
		logger.Warn("member does not own this reservation",
			zap.String("owner_id", reservationEntity.MemberID),
		)
		return CancelReservationResponse{}, errors2.ErrForbidden.WithDetails("reason", "you can only cancel your own reservations")
	}

	// Use domain service to cancel the reservation
	if err := uc.reservationService.MarkAsCancelled(&reservationEntity); err != nil {
		logger.Warn("cannot cancel reservation", zap.Error(err))
		return CancelReservationResponse{}, err
	}

	// Update in repository
	if err := uc.reservationRepo.Update(ctx, reservationEntity); err != nil {
		logger.Error("failed to update reservation", zap.Error(err))
		return CancelReservationResponse{}, errors2.Database("database operation", err)
	}

	logger.Info("reservation cancelled successfully")

	return CancelReservationResponse{
		Response: reservationdomain.ParseFromReservation(reservationEntity),
	}, nil
}
