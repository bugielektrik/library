// Package mongo provides MongoDB repository implementations.
//
// Status: EXPERIMENTAL / NOT ACTIVELY USED
//
// This package contains MongoDB implementations of domain repository interfaces.
// It was created as an alternative to the PostgreSQL implementation but is not
// currently used in production.
//
// Primary Implementation:
// The postgres package (internal/adapters/repository/postgres) is the primary
// and recommended repository implementation for this project.
//
// If MongoDB support is needed:
//  1. Complete repository implementations for all domain entities
//  2. Add comprehensive integration tests
//  3. Update application bootstrap (internal/infrastructure/app/app.go)
//  4. Add MongoDB connection management
//  5. Update documentation and deployment guides
//
// Current State:
//   - Partial implementations may exist
//   - Not tested in production
//   - May be out of sync with current domain model
//   - No active maintenance
//
// Future Considerations:
// If polyglot persistence is required (using both PostgreSQL and MongoDB),
// consider implementing the repository interfaces for specific domains that
// benefit from document storage (e.g., audit logs, analytics data).
package mongo
