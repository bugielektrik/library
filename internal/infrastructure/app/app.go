// Package app provides application lifecycle management following clean architecture
package app

import (
	"context"
	"fmt"
	"library-service/internal/adapters/http"
	"library-service/internal/payments/gateway/epayment"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"library-service/internal/adapters/cache"
	"library-service/internal/adapters/repository"
	"library-service/internal/infrastructure/auth"
	"library-service/internal/infrastructure/config"
	"library-service/internal/infrastructure/log"
	"library-service/internal/usecase"
)

// App represents the application with all its dependencies
type App struct {
	logger       *zap.Logger
	config       *config.Config
	repositories *repository.Repositories
	caches       *cache.Caches
	authServices *usecase.AuthServices
	usecases     *usecase.Container
	server       *http.Server
}

// New creates a new application instance.
//
// Bootstrap Order (CRITICAL - must follow this sequence):
//  1. Logger - First so all subsequent steps can log
//  2. Config - Load environment variables and settings
//  3. Repositories - PostgreSQL/memory implementations
//  4. Caches - Redis/memory cache layer
//  5. Auth Services - JWT + Password (infrastructure services)
//  6. Gateway Services - Payment gateway client
//  7. Use Case Container - Wires everything together
//  8. HTTP Server - Routes and middleware
//
// See Also:
//   - Counterpart: internal/usecase/container.go (domain service creation)
//   - ADR: .claude/adr/003-domain-services-vs-infrastructure.md (where to create services)
//   - ADR: .claude/adr/002-clean-architecture-boundaries.md (layer dependencies)
//   - Example: internal/infrastructure/auth/jwt.go (infrastructure service)
//   - Documentation: cmd/api/main.go (application entry point)
func New() (*App, error) {
	app := &App{}

	// Initialize logger first
	logger, err := log.NewLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	app.logger = logger

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		app.logger.Error("failed to load config", zap.Error(err))
		return nil, err
	}
	app.config = cfg
	app.logger.Info("configuration loaded", zap.String("mode", cfg.App.Mode))

	// Initialize repositories
	repos, err := repository.NewRepositories(repository.WithMemoryStore())
	if err != nil {
		app.logger.Error("failed to initialize repositories", zap.Error(err))
		return nil, err
	}
	app.repositories = repos
	app.logger.Info("repositories initialized")

	// Initialize caches
	caches, err := cache.NewCaches(
		cache.Dependencies{Repositories: repos},
		cache.WithMemoryStore(),
	)
	if err != nil {
		app.logger.Error("failed to initialize caches", zap.Error(err))
		return nil, err
	}
	app.caches = caches
	app.logger.Info("caches initialized")

	// Initialize auth services
	authServices := &usecase.AuthServices{
		JWTService: auth.NewJWTService(
			cfg.JWT.Secret,
			cfg.JWT.AccessTokenTTL,
			cfg.JWT.RefreshTokenTTL,
			cfg.JWT.Issuer,
		),
		PasswordService: auth.NewPasswordService(),
	}
	app.authServices = authServices
	app.logger.Info("auth services initialized")

	// Initialize payment gateway
	epaymentConfig, err := epayment.LoadConfigFromEnv()
	if err != nil {
		app.logger.Warn("failed to load epayment config, payment features may not work", zap.Error(err))
	}
	paymentGateway := epayment.NewGateway(epaymentConfig, app.logger)

	gatewayServices := &usecase.GatewayServices{
		PaymentGateway: paymentGateway,
	}
	app.logger.Info("gateway services initialized")

	// Initialize usecases
	usecaseRepos := &usecase.Repositories{
		Book:        repos.Book,
		Author:      repos.Author,
		Member:      repos.Member,
		Reservation: repos.Reservation,
		Payment:     repos.Payment,
	}
	usecaseCaches := &usecase.Caches{
		Book:   caches.Book,
		Author: caches.Author,
	}
	usecases := usecase.NewContainer(usecaseRepos, usecaseCaches, authServices, gatewayServices)
	app.usecases = usecases
	app.logger.Info("usecases initialized")

	// Initialize HTTP server
	srv, err := http.NewHTTPServer(cfg, usecases, authServices, app.logger)
	if err != nil {
		app.logger.Error("failed to initialize server", zap.Error(err))
		return nil, err
	}
	app.server = srv
	app.logger.Info("server initialized")

	return app, nil
}

// Run starts the application and handles graceful shutdown
func (a *App) Run() error {
	// Start server
	if err := a.server.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	a.logger.Info("application started",
		zap.String("port", a.config.App.Port),
		zap.String("mode", a.config.App.Mode),
	)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	a.logger.Info("shutting down application...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("server shutdown error", zap.Error(err))
	}

	if a.repositories != nil {
		a.repositories.Close()
	}

	a.logger.Info("application stopped")
	return nil
}
