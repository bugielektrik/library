// Package reservation provides HTTP handlers for book reservation operations.
//
// This package handles reservation-related HTTP requests including:
//   - Create book reservation (POST /reservations)
//   - Get reservation by ID (GET /reservations/{id})
//   - List member reservations (GET /reservations)
//   - Cancel reservation (DELETE /reservations/{id})
//
// Reservations allow members to:
//   - Reserve books for future pickup
//   - Hold books temporarily while deciding
//   - Queue for popular books
//
// Business Rules:
//   - Members can only cancel their own reservations
//   - Reservations expire after configured period
//   - Payment may be required to confirm reservation
//
// All endpoints require authentication (JWT middleware applied in router).
//
// Handler Organization:
//   - handler.go: Handler struct, routes, and constructor
//   - crud.go: Create, Get, Cancel (Delete) operations
//   - query.go: List and search operations
//
// Related Packages:
//   - Use Cases: internal/usecase/reservationops/ (reservation business logic)
//   - Domain: internal/domain/reservation/ (reservation entity and service)
//   - DTOs: internal/adapters/http/dto/reservation.go (request/response types)
//
// Example Usage:
//
//	reservationHandler := reservation.NewReservationHandler(useCases, validator)
//	router.Group(func(r chi.Router) {
//	    r.Use(authMiddleware.Authenticate)
//	    r.Mount("/reservations", reservationHandler.Routes())
//	})
package reservation
