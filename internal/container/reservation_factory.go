package container

import (
	memberdomain "library-service/internal/members/domain"
	reservationdomain "library-service/internal/reservations/domain"
	reservationservice "library-service/internal/reservations/service"
)

// ================================================================================
// Factory Functions - Reservation Domain
// ================================================================================

// newReservationUseCases creates all reservation-related use cases
func newReservationUseCases(
	reservationRepo reservationdomain.Repository,
	memberRepo memberdomain.Repository,
) ReservationUseCases {
	// Create domain service
	reservationService := reservationdomain.NewService()

	return ReservationUseCases{
		CreateReservation:      reservationservice.NewCreateReservationUseCase(reservationRepo, memberRepo, reservationService),
		CancelReservation:      reservationservice.NewCancelReservationUseCase(reservationRepo, reservationService),
		GetReservation:         reservationservice.NewGetReservationUseCase(reservationRepo),
		ListMemberReservations: reservationservice.NewListMemberReservationsUseCase(reservationRepo),
	}
}
