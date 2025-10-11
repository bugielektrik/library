// Package app provides application bootstrap and initialization.
//
// This package orchestrates the application startup sequence:
//  1. Logger initialization
//  2. Configuration loading
//  3. Database connection setup
//  4. Cache initialization
//  5. Authentication services setup
//  6. Use case container wiring
//  7. HTTP server initialization
//  8. Graceful shutdown handling
//
// The app package brings together all infrastructure components and
// dependency injection to create a fully initialized application ready
// to serve requests.
package app
