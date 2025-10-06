package reservationops

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/domain/reservation"
	"library-service/internal/infrastructure/log"
	"library-service/pkg/errors"
)

// ListMemberReservationsRequest represents the input for listing a member's reservations
type ListMemberReservationsRequest struct {
	MemberID string
}

// ListMemberReservationsResponse represents the output of listing reservations
type ListMemberReservationsResponse struct {
	Reservations []reservation.Response
}

// ListMemberReservationsUseCase handles listing all reservations for a member
type ListMemberReservationsUseCase struct {
	reservationRepo reservation.Repository
}

// NewListMemberReservationsUseCase creates a new instance of ListMemberReservationsUseCase
func NewListMemberReservationsUseCase(reservationRepo reservation.Repository) *ListMemberReservationsUseCase {
	return &ListMemberReservationsUseCase{
		reservationRepo: reservationRepo,
	}
}

// Execute retrieves all reservations for a member
func (uc *ListMemberReservationsUseCase) Execute(ctx context.Context, req ListMemberReservationsRequest) (ListMemberReservationsResponse, error) {
	logger := log.FromContext(ctx).Named("list_member_reservations_usecase").With(
		zap.String("member_id", req.MemberID),
	)

	// Get reservations from repository
	reservations, err := uc.reservationRepo.GetByMemberID(ctx, req.MemberID)
	if err != nil {
		logger.Error("failed to get member reservations", zap.Error(err))
		return ListMemberReservationsResponse{}, errors.ErrDatabase.Wrap(err)
	}

	logger.Info("member reservations retrieved successfully", zap.Int("count", len(reservations)))

	return ListMemberReservationsResponse{
		Reservations: reservation.ParseFromReservations(reservations),
	}, nil
}
