// Package log provides structured logging configuration.
//
// This package configures and initializes the logging system using Zap.
//
// Features:
//   - Structured logging with JSON output in production
//   - Console-friendly output in development
//   - Log level configuration (debug, info, warn, error)
//   - Request context propagation
//   - Performance optimized logging
//
// The logger is initialized at application startup and made available
// throughout the application via the logutil package helpers.
package log
