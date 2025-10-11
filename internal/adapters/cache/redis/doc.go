// Package redis provides Redis-based cache implementations.
//
// This package contains concrete implementations of cache interfaces
// defined in the domain layer, using Redis as the caching backend.
//
// Features:
//   - Key-value storage with TTL support
//   - JSON serialization for complex types
//   - Connection pooling
//   - Error handling and fallback logic
//
// Cache implementations follow the same interface as memory-based caches,
// allowing easy swapping between caching strategies.
package redis
