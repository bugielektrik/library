// Package service implements use cases for book reservation service.
//
// This package orchestrates book reservation workflows, allowing members to
// reserve books that are currently checked out. When reserved books are returned,
// the system notifies the member with an active reservation.
//
// Use cases implemented:
//   - CreateReservationUseCase: Creates new book reservation for a member
//   - CancelReservationUseCase: Cancels an active or pending reservation
//   - GetReservationUseCase: Retrieves reservation details by ID
//   - ListMemberReservationsUseCase: Returns all reservations for a member
//   - ProcessExpiredReservationsUseCase: Handles reservation expiry (background job)
//
// Dependencies:
//   - reservation.Repository: For reservation persistence
//   - reservation.Service: For reservation business rules and validation
//   - book.Repository: To check book availability
//   - domain.Repository: To verify member eligibility
//
// Example usage:
//
//	createUC := reservationops.NewCreateReservationUseCase(
//	    reservationRepo, reservationService, bookRepo, memberRepo,
//	)
//	response, err := createUC.Execute(ctx, reservationops.CreateReservationRequest{
//	    MemberID: "member-uuid",
//	    BookID:   "book-uuid",
//	})
//	// response contains: ReservationID, Status, ExpiresAt
//
// Reservation lifecycle:
//  1. Pending: Book currently checked out, reservation created
//  2. Ready: Book returned, member notified (48h to pick up)
//  3. Fulfilled: Member checked out the book
//  4. Expired: Member didn't pick up within 48h, next in queue notified
//  5. Cancelled: Member or admin cancelled the reservation
//
// Business rules:
//   - Members can only reserve books that are currently checked out
//   - Maximum 3 active reservations per member
//   - Reservations expire after 48h in "Ready" status
//   - Queue priority: First-come, first-served (FIFO)
//   - Active subscription required to create reservations
//
// Expiry handling:
//   - Background job checks for expired "Ready" reservations every hour
//   - Expired reservations automatically moved to next member in queue
//   - Members notified via email when reservation becomes ready
//
// Architecture:
//   - Package name uses "operations" to align with bounded context structure
//   - Reservation queue logic in domain.Service
//   - Email notifications triggered via event system
//   - Expiry processing suitable for cron job or worker
package service
