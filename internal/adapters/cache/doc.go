// Package cache provides caching layer implementations for domain entities.
//
// This package implements cache adapters that optimize performance by reducing
// database queries for frequently accessed data. Multiple backend implementations
// are available with automatic fallback.
//
// Implementations:
//   - redis/: Redis cache implementation (primary, production)
//   - memory/: In-memory cache with LRU eviction (fallback, development)
//
// Cached entities:
//   - Books: Cache by ID and ISBN (TTL: 1 hour)
//   - Members: Cache by ID (TTL: 30 minutes)
//   - Authors: Cache by ID (TTL: 1 hour)
//   - JWT tokens: Cache for validation (TTL: token expiry)
//
// Cache patterns:
//   - Cache-aside: Check cache first, fallback to database
//   - Write-through: Update cache on write operations
//   - Cache invalidation: Clear on updates/deletes
//   - TTL-based expiration: Automatic cleanup
//
// Example usage:
//
//	// In use case
//	func (uc *GetBookUseCase) Execute(ctx context.Context, req Request) (Response, error) {
//	    // Try cache first
//	    book, err := uc.bookCache.Get(ctx, req.ID)
//	    if err == nil {
//	        return toResponse(book), nil
//	    }
//	    // Cache miss, query database
//	    book, err = uc.bookRepo.Get(ctx, req.ID)
//	    if err != nil {
//	        return Response{}, err
//	    }
//	    // Update cache
//	    _ = uc.bookCache.Set(ctx, req.ID, book)
//	    return toResponse(book), nil
//	}
//
// Redis implementation:
//   - Connection pooling for performance
//   - Serialization: JSON encoding for complex types
//   - Key naming: Consistent prefix pattern (e.g., "book:ID", "member:EMAIL")
//   - Cluster support: Redis Cluster for high availability
//
// Memory implementation:
//   - LRU eviction policy (max 1000 items by default)
//   - Thread-safe using sync.RWMutex
//   - Used when Redis unavailable (automatic fallback)
//   - No persistence: Cleared on application restart
//
// Error handling:
//   - Cache errors are non-critical (logged, not propagated)
//   - Application continues if cache unavailable
//   - Automatic fallback to memory cache on Redis failure
//
// Configuration:
//   - Redis host/port via environment variables
//   - TTL configurable per entity type
//   - Max memory size for in-memory cache
//   - Enable/disable caching globally
//
// Monitoring:
//   - Cache hit/miss metrics
//   - Eviction rate tracking
//   - Connection pool statistics
package cache
