// Package infrastructure provides cross-cutting concerns and technical infrastructure.
//
// This package contains the outermost layer of clean architecture, providing
// technical capabilities that support the entire application.
//
// Infrastructure components:
//   - auth: JWT authentication and authorization
//   - config: Application configuration management
//   - database: Database connection pooling and management
//   - http: HTTP server setup and routing
//   - logger: Structured logging with Zap
//   - store: Database and cache store management
//   - app: Application initialization and dependency wiring
//
// Design principles:
//   - Cross-cutting concerns separated from business logic
//   - Configuration loaded from environment variables
//   - Centralized logging and error handling
//   - Database connection management with health checks
//
// Example usage:
//
//	// Application initialization
//	app, err := app.New()
//	if err != nil {
//		log.Fatal(err)
//	}
//	app.Run()
package infrastructure
