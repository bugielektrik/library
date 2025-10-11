// Package memory provides in-memory repository implementations for testing.
//
// This package contains in-memory implementations of domain repository interfaces,
// storing data in Go maps and slices. These implementations are suitable for:
//   - Unit testing without database dependencies
//   - Local development and rapid prototyping
//   - Integration tests requiring isolated state
//   - Demonstrations and proof-of-concepts
//
// Characteristics:
//   - Data stored in memory (not persisted across restarts)
//   - No external dependencies (no database required)
//   - Fast operations (no network/disk I/O)
//   - Thread-safe with mutex protection
//   - Implements same interfaces as postgres package
//
// Limitations:
//   - Data lost on application restart
//   - No transaction support
//   - No complex query capabilities
//   - Limited to single-instance deployment
//
// Primary Implementation:
// The postgres package is the primary production implementation.
// Use memory package only for testing and development.
//
// Example Usage:
//
//	// In tests
//	bookRepo := memory.NewBookRepository()
//	memberRepo := memory.NewMemberRepository()
//
//	// Use same interface as postgres
//	book, err := bookRepo.Get(ctx, bookID)
package memory
