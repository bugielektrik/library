package app

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// WarmingConfig configures cache warming behavior
type WarmingConfig struct {
	// Enabled controls whether cache warming runs
	Enabled bool

	// PopularBookLimit is the number of popular books to warm (0 = all)
	PopularBookLimit int

	// PopularAuthorLimit is the number of popular authors to warm (0 = all)
	PopularAuthorLimit int

	// Timeout is the maximum time to spend warming caches
	Timeout time.Duration

	// Logger for warming progress
	Logger *zap.Logger
}

// DefaultWarmingConfig returns sensible defaults
func DefaultWarmingConfig(logger *zap.Logger) WarmingConfig {
	return WarmingConfig{
		Enabled:            true,
		PopularBookLimit:   50, // Top 50 books
		PopularAuthorLimit: 20, // Top 20 authors
		Timeout:            30 * time.Second,
		Logger:             logger,
	}
}

// WarmCaches pre-loads frequently accessed data into caches.
//
// This function runs during application startup to populate caches with
// commonly accessed books and authors, reducing initial response latency
// for frequently requested items.
//
// Strategy:
//   - Loads most recently added books (proxy for popularity)
//   - Loads authors with most books (active authors)
//   - Runs with timeout to prevent blocking startup
//   - Logs progress and errors (non-fatal)
//
// Performance:
//   - Runs asynchronously relative to server startup
//   - Uses context timeout for safety
//   - Batch loads from repository
//
// Example usage:
//
//	cfg := cache.DefaultWarmingConfig(logger)
//	if err := cache.WarmCaches(ctx, caches, cfg); err != nil {
//	    logger.Warn("cache warming failed", zap.Error(err))
//	}
func WarmCaches(ctx context.Context, caches *Caches, config WarmingConfig) error {
	if !config.Enabled {
		config.Logger.Info("cache warming disabled")
		return nil
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	startTime := time.Now()
	config.Logger.Info("starting cache warming",
		zap.Int("book_limit", config.PopularBookLimit),
		zap.Int("author_limit", config.PopularAuthorLimit),
		zap.Duration("timeout", config.Timeout),
	)

	var warmedBooks, warmedAuthors int
	var bookErr, authorErr error

	// Warm books cache
	warmedBooks, bookErr = warmBooksCache(ctx, caches, config)
	if bookErr != nil {
		config.Logger.Warn("book cache warming incomplete", zap.Error(bookErr))
	}

	// Warm authors cache (continue even if books failed)
	warmedAuthors, authorErr = warmAuthorsCache(ctx, caches, config)
	if authorErr != nil {
		config.Logger.Warn("author cache warming incomplete", zap.Error(authorErr))
	}

	duration := time.Since(startTime)
	config.Logger.Info("cache warming completed",
		zap.Int("books_warmed", warmedBooks),
		zap.Int("authors_warmed", warmedAuthors),
		zap.Duration("duration", duration),
	)

	// Return error only if both failed
	if bookErr != nil && authorErr != nil {
		return fmt.Errorf("cache warming failed: books=%v, authors=%v", bookErr, authorErr)
	}

	return nil
}

// warmBooksCache loads popular books into cache
func warmBooksCache(ctx context.Context, caches *Caches, config WarmingConfig) (int, error) {
	bookRepo := caches.dependencies.Repositories.Book
	bookCache := caches.Book

	// Get all books
	books, err := bookRepo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list books: %w", err)
	}

	// Limit to configured number
	limit := config.PopularBookLimit
	if limit == 0 || limit > len(books) {
		limit = len(books)
	}

	warmed := 0
	for i := 0; i < limit; i++ {
		b := books[i]

		// Check context timeout
		if ctx.Err() != nil {
			return warmed, fmt.Errorf("timeout warming books: %w", ctx.Err())
		}

		// Warm cache
		if err := bookCache.Set(ctx, b.ID, b); err != nil {
			config.Logger.Debug("failed to warm book cache",
				zap.String("book_id", b.ID),
				zap.Error(err),
			)
			continue
		}
		warmed++
	}

	return warmed, nil
}

// warmAuthorsCache loads popular authors into cache
func warmAuthorsCache(ctx context.Context, caches *Caches, config WarmingConfig) (int, error) {
	authorRepo := caches.dependencies.Repositories.Author
	authorCache := caches.Author

	// Get all authors
	authors, err := authorRepo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list authors: %w", err)
	}

	// Limit to configured number
	limit := config.PopularAuthorLimit
	if limit == 0 || limit > len(authors) {
		limit = len(authors)
	}

	warmed := 0
	for i := 0; i < limit; i++ {
		a := authors[i]

		// Check context timeout
		if ctx.Err() != nil {
			return warmed, fmt.Errorf("timeout warming authors: %w", ctx.Err())
		}

		// Warm cache
		if err := authorCache.Set(ctx, a.ID, a); err != nil {
			config.Logger.Debug("failed to warm author cache",
				zap.String("author_id", a.ID),
				zap.Error(err),
			)
			continue
		}
		warmed++
	}

	return warmed, nil
}

// WarmCachesAsync runs cache warming in the background without blocking startup.
//
// This is the recommended approach for production systems where startup time
// is critical. Cache warming runs asynchronously and logs any errors.
//
// Example usage:
//
//	go cache.WarmCachesAsync(ctx, caches, cache.DefaultWarmingConfig(logger))
func WarmCachesAsync(ctx context.Context, caches *Caches, config WarmingConfig) {
	go func() {
		if err := WarmCaches(ctx, caches, config); err != nil {
			config.Logger.Error("async cache warming failed", zap.Error(err))
		}
	}()
}
