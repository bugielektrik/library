# Cache Warming Implementation

**Date:** October 11, 2025
**Feature:** Automatic cache pre-loading for frequently accessed data

## Overview

Cache warming pre-loads frequently accessed books and authors into cache during application startup, reducing initial response latency for commonly requested items.

## Implementation

### Files Added

1. **`/internal/infrastructure/pkg/cache/warming.go`** - Core cache warming implementation
2. **`/internal/infrastructure/pkg/cache/warming_test.go`** - Comprehensive test coverage

### Files Modified

1. **`/internal/app/app.go`** - Integrated cache warming into application bootstrap

## Architecture

### Strategy

- **Books:** Pre-loads top N books (default: 50) from repository
- **Authors:** Pre-loads top N authors (default: 20) from repository
- **Execution:** Asynchronous (non-blocking startup)
- **Timeout:** 30 seconds maximum to prevent blocking
- **Failure handling:** Non-fatal logging, server continues if warming fails

### Configuration

```go
type WarmingConfig struct {
    Enabled            bool          // Default: true
    PopularBookLimit   int           // Default: 50
    PopularAuthorLimit int           // Default: 20
    Timeout            time.Duration // Default: 30s
    Logger             *zap.Logger
}
```

### Usage

**Automatic (Default):**
Cache warming runs automatically on application startup via `app.go`:

```go
// Warm caches asynchronously (non-blocking)
go cache.WarmCachesAsync(context.Background(), caches, cache.DefaultWarmingConfig(app.logger))
```

**Manual:**
```go
// Synchronous warming (blocks until complete)
err := cache.WarmCaches(ctx, caches, config)

// Asynchronous warming (background)
cache.WarmCachesAsync(ctx, caches, config)
```

## Benefits

1. **Reduced Latency:** First requests for popular items served from cache
2. **Non-Blocking:** Startup time not impacted (runs asynchronously)
3. **Configurable:** Limits and timeout adjustable per environment
4. **Safe:** Timeout and error handling prevent startup failures
5. **Observable:** Comprehensive logging of warming progress

## Performance Impact

- **Startup:** No impact (runs in background goroutine)
- **First requests:** Significantly faster for warmed items
- **Memory:** Minimal (~5-10KB per item × limits)
- **Database:** One-time load on startup (SELECT queries)

## Testing

Comprehensive test coverage including:
- ✅ Synchronous warming with various limits
- ✅ Asynchronous warming (thread-safe)
- ✅ Configuration defaults
- ✅ Disable warming option
- ✅ Timeout handling
- ✅ Race condition testing

## Monitoring

Cache warming logs the following:
- Start of warming with configured limits
- Progress per cache (books/authors warmed)
- Warnings for partial failures
- Final summary with duration and counts

Example log output:
```
INFO    starting cache warming    book_limit=50 author_limit=20 timeout=30s
INFO    cache warming completed   books_warmed=50 authors_warmed=20 duration=125ms
```

## Future Enhancements (Optional)

1. **Analytics-driven warming:** Use actual access patterns instead of "top N"
2. **Background refresh:** Periodic re-warming of popular items
3. **TTL optimization:** Dynamic TTL based on access frequency
4. **Metrics collection:** Track cache hit rates, warming effectiveness
5. **Conditional warming:** Only warm if cache is empty

## Configuration Options

### Environment Variables (Future)

Could add environment variables for runtime configuration:
```bash
CACHE_WARMING_ENABLED=true
CACHE_WARMING_BOOK_LIMIT=100
CACHE_WARMING_AUTHOR_LIMIT=50
CACHE_WARMING_TIMEOUT=60s
```

Currently uses code-level defaults via `DefaultWarmingConfig()`.

## Troubleshooting

**Cache warming fails:**
- Check database connectivity
- Verify repository List() methods work
- Check timeout (may need longer for large datasets)
- Review logs for specific errors

**Warming takes too long:**
- Reduce PopularBookLimit and PopularAuthorLimit
- Increase timeout
- Check database query performance

**Items not cached:**
- Verify warming completed successfully (check logs)
- Ensure cache backend is available (Redis/memory)
- Check cache Set() implementation

## Related Files

- `/internal/infrastructure/pkg/cache/cache.go` - Cache infrastructure
- `/internal/app/app.go` - Application bootstrap
- `/internal/books/domain/book/cache.go` - Book cache interface
- `/internal/books/domain/author/cache.go` - Author cache interface

## ADR Reference

Consider creating ADR 014: Cache Warming Strategy if this feature requires formal documentation of architectural decisions.
