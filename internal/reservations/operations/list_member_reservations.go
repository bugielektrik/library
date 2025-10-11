package operations

import (
	"context"

	"go.uber.org/zap"

	reservationdomain "library-service/internal/reservations/domain"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
)

// ListMemberReservationsRequest represents the input for listing a member's reservations
type ListMemberReservationsRequest struct {
	MemberID string
}

// ListMemberReservationsResponse represents the output of listing reservations
type ListMemberReservationsResponse struct {
	Reservations []reservationdomain.Response
}

// ListMemberReservationsUseCase handles listing all reservations for a member
type ListMemberReservationsUseCase struct {
	reservationRepo reservationdomain.Repository
}

// NewListMemberReservationsUseCase creates a new instance of ListMemberReservationsUseCase
func NewListMemberReservationsUseCase(reservationRepo reservationdomain.Repository) *ListMemberReservationsUseCase {
	return &ListMemberReservationsUseCase{
		reservationRepo: reservationRepo,
	}
}

// Execute retrieves all reservations for a member
func (uc *ListMemberReservationsUseCase) Execute(ctx context.Context, req ListMemberReservationsRequest) (ListMemberReservationsResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "reservation", "list_member")

	// Get reservations from repository
	reservations, err := uc.reservationRepo.GetByMemberID(ctx, req.MemberID)
	if err != nil {
		logger.Error("failed to get member reservations", zap.Error(err))
		return ListMemberReservationsResponse{}, errors.Database("database operation", err)
	}

	logger.Info("member reservations retrieved successfully", zap.Int("count", len(reservations)))

	return ListMemberReservationsResponse{
		Reservations: reservationdomain.ParseFromReservations(reservations),
	}, nil
}
