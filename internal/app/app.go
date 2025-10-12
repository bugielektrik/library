// Package app provides application lifecycle management following clean architecture
package app

import (
	"context"
	"fmt"
	"library-service/internal/container"
	domainapp "library-service/internal/domain/app"
	epayment2 "library-service/internal/payments/provider/epayment"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"library-service/internal/infrastructure/auth"
	"library-service/internal/infrastructure/config"
	"library-service/internal/infrastructure/log"
	"library-service/internal/infrastructure/shutdown"
)

// App represents the application with all its dependencies
type App struct {
	logger       *zap.Logger
	config       *config.Config
	repositories *domainapp.Repositories
	caches       *domainapp.Caches
	authServices *container.AuthServices
	usecases     *container.Container
	httpServer   *Server
}

// Validator wraps go-playground/validator
type Validator struct {
	validate *validator.Validate
}

// Validate validates a struct
func (v *Validator) Validate(i interface{}) error {
	if v.validate == nil {
		v.validate = validator.New()
	}
	return v.validate.Struct(i)
}

// New creates a new application instance.
//
// Bootstrap Order (CRITICAL - must follow this sequence):
//  1. Logger - First so all subsequent steps can log
//  2. Config - Load environment variables and settings
//  3. Repositories - PostgreSQL/memory implementations
//  4. Caches - Redis/memory cache layer
//  5. Auth Services - JWT + Password (infrastructure service)
//  6. Gateway Services - Payment provider client
//  7. Use Case Container - Wires everything together
//  8. HTTP Server - Routes and middleware
//
// See Also:
//   - Counterpart: internal/usecase/container.go (domain service creation)
//   - ADR: .claude/adr/003-domain-service-vs-infrastructure.md (where to create service)
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
	cfg := config.MustLoad("")
	app.config = cfg
	app.logger.Info("configuration loaded", zap.String("environment", cfg.App.Environment))

	// Initialize repositories
	repos, err := domainapp.NewRepositories(domainapp.WithMemoryStore())
	if err != nil {
		app.logger.Error("failed to initialize repositories", zap.Error(err))
		return nil, err
	}
	app.repositories = repos
	app.logger.Info("repositories initialized")

	// Initialize caches
	caches, err := domainapp.NewCaches(
		domainapp.Dependencies{Repositories: repos},
		domainapp.WithMemoryCache(),
	)
	if err != nil {
		app.logger.Error("failed to initialize caches", zap.Error(err))
		return nil, err
	}
	app.caches = caches
	app.logger.Info("caches initialized")

	// Warm caches asynchronously (non-blocking)
	go domainapp.WarmCachesAsync(context.Background(), caches, domainapp.DefaultWarmingConfig(app.logger))

	// Initialize auth service
	authServices := &container.AuthServices{
		JWTService: auth.NewJWTService(
			cfg.JWT.Secret,
			cfg.JWT.AccessTokenTTL,
			cfg.JWT.RefreshTokenTTL,
			cfg.JWT.Issuer,
		),
		PasswordService: auth.NewPasswordService(),
	}
	app.authServices = authServices
	app.logger.Info("auth service initialized")

	// Initialize payment provider
	epaymentConfig, err := epayment2.LoadConfigFromEnv()
	if err != nil {
		app.logger.Warn("failed to load epayment config, payment features may not work", zap.Error(err))
	}
	paymentGateway := epayment2.NewGateway(epaymentConfig, app.logger)

	gatewayServices := &container.GatewayServices{
		PaymentGateway: paymentGateway,
	}
	app.logger.Info("provider service initialized")

	// Initialize validator
	validator := &Validator{}
	app.logger.Info("validator initialized")

	// Initialize usecases
	usecaseRepos := &container.Repositories{
		Book:          repos.Book,
		Author:        repos.Author,
		Member:        repos.Member,
		Reservation:   repos.Reservation,
		Payment:       repos.Payment,
		SavedCard:     repos.SavedCard,
		CallbackRetry: repos.CallbackRetry,
		Receipt:       repos.Receipt,
	}
	usecaseCaches := &container.Caches{
		Book:   caches.Book,
		Author: caches.Author,
	}
	usecases := container.NewContainer(usecaseRepos, usecaseCaches, authServices, gatewayServices, validator)
	app.usecases = usecases
	app.logger.Info("usecases initialized")

	// Initialize HTTP server
	httpSrv, err := NewHTTPServer(cfg, usecases, authServices, app.logger)
	if err != nil {
		app.logger.Error("failed to initialize server", zap.Error(err))
		return nil, err
	}
	app.httpServer = httpSrv
	app.logger.Info("server initialized")

	return app, nil
}

// Run starts the application and handles graceful shutdown with phased execution.
//
// Shutdown Phases:
//  1. Pre-shutdown: Mark service unhealthy, prepare for shutdown
//  2. Stop accepting: Stop accepting new connections
//  3. Drain connections: Wait for in-flight requests (10s max)
//  4. Cleanup: Close DB, cache, external connections
//  5. Post-shutdown: Flush logs, final cleanup
//
// Total shutdown time: ~20 seconds maximum
//
// See Also:
//   - Shutdown manager: internal/infrastructure/shutdown/shutdown.go
func (a *App) Run() error {
	// Start server
	if err := a.httpServer.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	a.logger.Info("application started",
		zap.Int("port", a.config.Server.Port),
		zap.String("environment", a.config.App.Environment),
	)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	sig := <-quit

	a.logger.Info("received shutdown signal",
		zap.String("signal", sig.String()),
	)

	// Create shutdown manager and register hooks
	shutdownMgr := shutdown.NewManager(a.logger)
	shutdownMgr.RegisterDefaultHooks(a.httpServer, a.repositories)

	// Register custom cache cleanup hook
	if a.caches != nil {
		shutdownMgr.RegisterHook(shutdown.PhaseCleanup, "close_caches", func(ctx context.Context) error {
			a.logger.Info("closing cache connections")
			// Caches will be closed when repositories close
			return nil
		})
	}

	// Execute graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := shutdownMgr.Shutdown(ctx); err != nil {
		a.logger.Error("graceful shutdown completed with errors", zap.Error(err))
		return err
	}

	a.logger.Info("application stopped gracefully")
	return nil
}
