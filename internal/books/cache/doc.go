// Package cache provides cache implementations for the books bounded context.
//
// This package contains both memory and Redis cache implementations for
// Book and Author entities, keeping cache infrastructure colocated with
// the domain it serves.
//
// Implementations:
//   - memory.BookCache: In-memory cache for books
//   - memory.AuthorCache: In-memory cache for authors
//   - redis.BookCache: Redis-backed cache for books
//   - redis.AuthorCache: Redis-backed cache for authors
//
// The cache implementations are used by book and author use cases to
// improve performance by reducing database queries for frequently
// accessed entities.
//
// Architecture:
//   - Cache interfaces defined in domain (book.Cache, author.Cache)
//   - Implementations provided here (memory, redis)
//   - Orchestration via adapters/cache (container, warming)
package cache
