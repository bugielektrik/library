// Package memory provides in-memory cache implementations.
//
// This package contains concrete implementations of cache interfaces
// using in-memory storage, suitable for development and testing.
//
// Features:
//   - Fast in-process caching
//   - Thread-safe operations
//   - TTL support with automatic expiration
//   - No external dependencies
//
// Memory caches are useful for:
//   - Development environments
//   - Testing scenarios
//   - Fallback when Redis is unavailable
package memory
