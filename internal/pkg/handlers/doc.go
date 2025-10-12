// Package handler provides HTTP handler for the REST API.
//
// This package contains common handler functionality (BaseHandler) and
// subpackages for each domain entity's HTTP handler.
//
// Subpackages:
//   - auth: Authentication and authorization handler
//   - author: Author management handler
//   - book: Book CRUD and query handler
//   - member: Member management handler
//   - payment: Payment processing handler
//   - receipt: Receipt generation and retrieval handler
//   - reservation: Book reservation handler
//   - savedcard: Saved card management handler
//
// BaseHandler:
// Provides common functionality for all handler including error handling,
// JSON response serialization, and helper methods for extracting data from
// requests (member ID, URL parameters).
//
// All handler subpackages include Swagger/OpenAPI annotations for API documentation.
package handlers
