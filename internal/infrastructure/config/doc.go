// Package config provides application configuration management.
//
// This package handles loading and validation of configuration from
// environment variables and configuration files.
//
// Configuration includes:
//   - Server settings (host, port, timeouts)
//   - Database connections (PostgreSQL DSN)
//   - Cache settings (Redis connection)
//   - Authentication (JWT secrets, token expiry)
//   - External services (payment gateways, email)
//
// Configuration is loaded at application startup and validated before
// the application begins serving requests.
package config
