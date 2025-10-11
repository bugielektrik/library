// Package logutil provides logging utilities and helper functions.
//
// This package simplifies common logging patterns throughout the application
// by providing standardized logger initialization functions for different
// architectural layers.
//
// # Key Features
//
//   - Consistent logger naming across use cases, handlers, and repositories
//   - Automatic context extraction with FromContext
//   - Layer-specific field conventions
//   - Reduced boilerplate for logger initialization
//
// # Usage Patterns
//
// Use Case Layer:
//
//	logger := logutil.UseCaseLogger(ctx, "create_book",
//	    zap.String("isbn", req.ISBN),
//	    zap.String("title", req.Title),
//	)
//
// HTTP Handler Layer:
//
//	logger := logutil.HandlerLogger(ctx, "book_handler", "create")
//
// Repository Layer:
//
//	logger := logutil.RepositoryLogger(ctx, "book_repository", "create")
//
// # Benefits
//
//   - 3-line logger initialization reduced to 1 line
//   - Consistent naming conventions (e.g., "create_book_usecase")
//   - Automatic context propagation
//   - Easier to update logging strategy globally
//
// # Design Philosophy
//
// This package follows the principle of "convention over configuration"
// by encoding logging best practices into simple, reusable functions.
// It reduces cognitive load and ensures consistency across 72+ logger
// initialization points in the codebase.
package logutil
