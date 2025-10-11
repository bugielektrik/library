package usecase

import (
	memberdomain "library-service/internal/members/domain"
	reservationdomain "library-service/internal/reservations/domain"
	reservationops "library-service/internal/reservations/operations"
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
	reservationRepo reservationdomain.Repository,
	memberRepo memberdomain.Repository,
) ReservationUseCases {
	// Create domain service
	reservationService := reservationdomain.NewService()

	return ReservationUseCases{
		CreateReservation:      reservationops.NewCreateReservationUseCase(reservationRepo, memberRepo, reservationService),
		CancelReservation:      reservationops.NewCancelReservationUseCase(reservationRepo, reservationService),
		GetReservation:         reservationops.NewGetReservationUseCase(reservationRepo),
		ListMemberReservations: reservationops.NewListMemberReservationsUseCase(reservationRepo),
	}
}
