// Package adapters contains implementations of external interfaces and integrations.
//
// This package implements the adapters layer (also known as interface adapters) in clean
// architecture, translating between the application core and external systems.
//
// The adapters layer includes:
//   - Inbound adapters (driving): HTTP handlers, gRPC services, CLI commands
//   - Outbound adapters (driven): Repository implementations, external service clients
//   - DTOs and request/response transformations
//
// Subpackages:
//   - http: REST API handlers and middleware
//   - grpc: gRPC service implementations
//   - repository: Database repository implementations (PostgreSQL, MongoDB, in-memory)
//   - cache: Cache implementations (Redis, in-memory)
//   - email: Email service adapters (SMTP)
//   - payment: Payment gateway integrations (Stripe, PayPal)
//   - storage: File storage adapters (S3, local)
//
// Design principles:
//   - Adapters implement domain interfaces
//   - Dependencies point inward (toward domain)
//   - Technology-specific implementation details isolated here
//   - Easy to swap implementations without affecting business logic
package adapters
