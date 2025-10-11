// Package dto provides data transfer objects for HTTP request and response handling.
//
// This package contains all the structs used for serializing and deserializing
// HTTP request and response bodies. DTOs follow the naming convention:
// {Operation}{Entity}Request/Response (e.g., CreateBookRequest, BookResponse).
//
// DTOs handle:
//   - JSON marshaling/unmarshaling
//   - Request validation tags
//   - Conversion to/from domain entities
//
// DTOs are independent of business logic and serve as the boundary between
// HTTP handlers and use cases.
package dto
