package usecase

import (
	"library-service/internal/domain/member"
	"library-service/internal/domain/reservation"
	"library-service/internal/usecase/reservationops"
)

// ReservationUseCases contains all reservation-related use cases
type ReservationUseCases struct {
	CreateReservation      *reservationops.CreateReservationUseCase
	CancelReservation      *reservationops.CancelReservationUseCase
	GetReservation         *reservationops.GetReservationUseCase
	ListMemberReservations *reservationops.ListMemberReservationsUseCase
}

// newReservationUseCases creates all reservation-related use cases
func newReservationUseCases(
	reservationRepo reservation.Repository,
	memberRepo member.Repository,
) ReservationUseCases {
	// Create domain service
	reservationService := reservation.NewService()

	return ReservationUseCases{
		CreateReservation:      reservationops.NewCreateReservationUseCase(reservationRepo, memberRepo, reservationService),
		CancelReservation:      reservationops.NewCancelReservationUseCase(reservationRepo, reservationService),
		GetReservation:         reservationops.NewGetReservationUseCase(reservationRepo),
		ListMemberReservations: reservationops.NewListMemberReservationsUseCase(reservationRepo),
	}
}
