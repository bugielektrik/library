// Package handlers provides HTTP handlers for the REST API.
//
// This package contains common handler functionality (BaseHandler) and
// subpackages for each domain entity's HTTP handlers.
//
// Subpackages:
//   - auth: Authentication and authorization handlers
//   - author: Author management handlers
//   - book: Book CRUD and query handlers
//   - member: Member management handlers
//   - payment: Payment processing handlers
//   - receipt: Receipt generation and retrieval handlers
//   - reservation: Book reservation handlers
//   - savedcard: Saved card management handlers
//
// BaseHandler:
// Provides common functionality for all handlers including error handling,
// JSON response serialization, and helper methods for extracting data from
// requests (member ID, URL parameters).
//
// All handler subpackages include Swagger/OpenAPI annotations for API documentation.
package handlers
