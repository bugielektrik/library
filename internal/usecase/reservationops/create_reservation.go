package reservationops

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/domain/member"
	"library-service/internal/domain/reservation"
	"library-service/internal/infrastructure/log"
	"library-service/pkg/errors"
)

// CreateReservationRequest represents the input for creating a reservation
type CreateReservationRequest struct {
	BookID   string
	MemberID string
}

// CreateReservationResponse represents the output of creating a reservation
type CreateReservationResponse struct {
	reservation.Response
}

// CreateReservationUseCase handles the creation of a new reservation
type CreateReservationUseCase struct {
	reservationRepo    reservation.Repository
	memberRepo         member.Repository
	reservationService *reservation.Service
}

// NewCreateReservationUseCase creates a new instance of CreateReservationUseCase
func NewCreateReservationUseCase(
	reservationRepo reservation.Repository,
	memberRepo member.Repository,
	reservationService *reservation.Service,
) *CreateReservationUseCase {
	return &CreateReservationUseCase{
		reservationRepo:    reservationRepo,
		memberRepo:         memberRepo,
		reservationService: reservationService,
	}
}

// Execute creates a new reservation in the system
func (uc *CreateReservationUseCase) Execute(ctx context.Context, req CreateReservationRequest) (CreateReservationResponse, error) {
	logger := log.FromContext(ctx).Named("create_reservation_usecase").With(
		zap.String("book_id", req.BookID),
		zap.String("member_id", req.MemberID),
	)

	// Get member to verify existence and get borrowed books
	memberEntity, err := uc.memberRepo.Get(ctx, req.MemberID)
	if err != nil {
		logger.Error("failed to get member", zap.Error(err))
		return CreateReservationResponse{}, errors.ErrMemberNotFound
	}

	// Get existing reservations for this member and book
	existingReservations, err := uc.reservationRepo.GetActiveByMemberAndBook(ctx, req.MemberID, req.BookID)
	if err != nil {
		logger.Error("failed to get existing reservations", zap.Error(err))
		return CreateReservationResponse{}, errors.ErrDatabase.Wrap(err)
	}

	// Check if member can reserve this book
	if err := uc.reservationService.CanMemberReserveBook(req.MemberID, req.BookID, existingReservations, memberEntity.Books); err != nil {
		logger.Warn("member cannot reserve book", zap.Error(err))
		return CreateReservationResponse{}, err
	}

	// Create reservation entity
	reservationEntity := reservation.New(reservation.Request{
		BookID:   req.BookID,
		MemberID: req.MemberID,
	})

	// Validate reservation using domain service
	if err := uc.reservationService.ValidateReservation(reservationEntity); err != nil {
		logger.Warn("validation failed", zap.Error(err))
		return CreateReservationResponse{}, err
	}

	// Save to repository
	id, err := uc.reservationRepo.Create(ctx, reservationEntity)
	if err != nil {
		logger.Error("failed to create reservation", zap.Error(err))
		return CreateReservationResponse{}, errors.ErrDatabase.Wrap(err)
	}
	reservationEntity.ID = id

	logger.Info("reservation created successfully", zap.String("id", id))

	return CreateReservationResponse{
		Response: reservation.ParseFromReservation(reservationEntity),
	}, nil
}
