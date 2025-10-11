// Package server provides HTTP server configuration and lifecycle management.
//
// This package handles:
//   - HTTP server initialization and configuration
//   - Graceful shutdown handling
//   - Request timeouts and limits
//   - Middleware setup
//   - Router configuration
//
// The server package wires together all HTTP components (handlers, middleware)
// and manages the server lifecycle from startup to graceful shutdown.
package server
