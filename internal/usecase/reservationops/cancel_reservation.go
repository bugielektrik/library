package reservationops

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/domain/reservation"
	"library-service/internal/infrastructure/log"
	"library-service/pkg/errors"
)

// CancelReservationRequest represents the input for cancelling a reservation
type CancelReservationRequest struct {
	ReservationID string
	MemberID      string // To verify ownership
}

// CancelReservationResponse represents the output of cancelling a reservation
type CancelReservationResponse struct {
	reservation.Response
}

// CancelReservationUseCase handles cancelling a reservation
type CancelReservationUseCase struct {
	reservationRepo    reservation.Repository
	reservationService *reservation.Service
}

// NewCancelReservationUseCase creates a new instance of CancelReservationUseCase
func NewCancelReservationUseCase(
	reservationRepo reservation.Repository,
	reservationService *reservation.Service,
) *CancelReservationUseCase {
	return &CancelReservationUseCase{
		reservationRepo:    reservationRepo,
		reservationService: reservationService,
	}
}

// Execute cancels a reservation
func (uc *CancelReservationUseCase) Execute(ctx context.Context, req CancelReservationRequest) (CancelReservationResponse, error) {
	logger := log.FromContext(ctx).Named("cancel_reservation_usecase").With(
		zap.String("reservation_id", req.ReservationID),
		zap.String("member_id", req.MemberID),
	)

	// Get the reservation
	reservationEntity, err := uc.reservationRepo.GetByID(ctx, req.ReservationID)
	if err != nil {
		logger.Error("failed to get reservation", zap.Error(err))
		return CancelReservationResponse{}, errors.ErrNotFound.WithDetails("resource", "reservation")
	}

	// Verify ownership
	if reservationEntity.MemberID != req.MemberID {
		logger.Warn("member does not own this reservation",
			zap.String("owner_id", reservationEntity.MemberID),
		)
		return CancelReservationResponse{}, errors.ErrForbidden.WithDetails("reason", "you can only cancel your own reservations")
	}

	// Use domain service to cancel the reservation
	if err := uc.reservationService.MarkAsCancelled(&reservationEntity); err != nil {
		logger.Warn("cannot cancel reservation", zap.Error(err))
		return CancelReservationResponse{}, err
	}

	// Update in repository
	if err := uc.reservationRepo.Update(ctx, reservationEntity); err != nil {
		logger.Error("failed to update reservation", zap.Error(err))
		return CancelReservationResponse{}, errors.ErrDatabase.Wrap(err)
	}

	logger.Info("reservation cancelled successfully")

	return CancelReservationResponse{
		Response: reservation.ParseFromReservation(reservationEntity),
	}, nil
}
