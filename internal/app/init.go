package app

import (
	"time"

	"go.uber.org/zap"

	"library-service/internal/cache"
	"library-service/internal/config"
	"library-service/internal/handler"
	"library-service/internal/repository"
	"library-service/internal/service"
	"library-service/pkg/server"
)

const (
	// dbStartupDelay is a small delay to allow database to be ready
	// This follows the original behavior for backward compatibility
	dbStartupDelay = time.Second
)

// initApp builds and initializes the application with all its dependencies.
// It follows the dependency injection pattern and ensures proper initialization order.
// Returns a fully initialized App instance or an error if any component fails to initialize.
func initApp(logger *zap.Logger) (*App, error) {
	app := &App{logger: logger}

	// Initialize components in dependency order
	if err := app.loadConfiguration(); err != nil {
		return nil, err
	}

	if err := app.initializeRepositories(); err != nil {
		return nil, err
	}

	if err := app.initializeCaches(); err != nil {
		app.cleanup()
		return nil, err
	}

	if err := app.initializeServices(); err != nil {
		app.cleanup()
		return nil, err
	}

	if err := app.initializeHandlers(); err != nil {
		app.cleanup()
		return nil, err
	}

	if err := app.initializeServers(); err != nil {
		app.cleanup()
		return nil, err
	}

	return app, nil
}

// loadConfiguration loads and validates application configuration
func (app *App) loadConfiguration() error {
	configs, err := config.New()
	if err != nil {
		app.logger.Error("config load error", zap.Error(err))
		return err
	}

	app.configs = configs
	app.logger.Info("configuration loaded",
		zap.String("mode", configs.APP.Mode),
		zap.String("port", configs.APP.Port),
	)

	return nil
}

// initializeRepositories sets up database connections and repositories
func (app *App) initializeRepositories() error {
	// Small backoff to allow DB to be ready (preserved from original)
	app.logger.Info("initializing repositories",
		zap.Duration("db startup delay", dbStartupDelay))
	time.Sleep(dbStartupDelay)

	repositories, err := repository.New(
		repository.WithMemoryStore(),
	)
	if err != nil {
		app.logger.Error("repository init error", zap.Error(err))
		return err
	}

	app.repositories = repositories
	app.logger.Info("repositories initialized")
	return nil
}

// initializeCaches sets up caching layers with repository dependencies
func (app *App) initializeCaches() error {
	caches, err := cache.New(
		cache.Dependencies{
			Repositories: app.repositories,
		},
		cache.WithMemoryStore(),
	)
	if err != nil {
		app.logger.Error("cache init error", zap.Error(err))
		return err
	}

	app.caches = caches
	app.logger.Info("caches initialized")
	return nil
}

// initializeServices composes business logic services from repositories and providers
func (app *App) initializeServices() error {
	services, err := service.New(
		service.Dependencies{
			Repositories: app.repositories,
			Caches:       app.caches,
		},
		service.WithLibraryService(),
		service.WithSubscriptionService(),
	)
	if err != nil {
		app.logger.Error("service init error", zap.Error(err))
		return err
	}

	app.services = services
	app.logger.Info("services initialized")
	return nil
}

// initializeHandlers sets up HTTP and other protocol handlers
func (app *App) initializeHandlers() error {
	handlers, err := handler.New(
		handler.Dependencies{
			Configs:  app.configs,
			Services: app.services,
		},
		handler.WithHTTPHandler(),
	)
	if err != nil {
		app.logger.Error("handler init error", zap.Error(err))
		return err
	}

	app.handlers = handlers
	app.logger.Info("handlers initialized")
	return nil
}

// initializeServers sets up HTTP and other protocol servers
func (app *App) initializeServers() error {
	servers, err := server.NewServer(
		server.WithHTTP(app.handlers.HTTP, app.configs.APP.Port),
	)
	if err != nil {
		app.logger.Error("server init error", zap.Error(err))
		return err
	}

	app.servers = servers
	app.logger.Info("servers initialized")
	return nil
}

// cleanup performs partial cleanup when initialization fails
// This ensures resources are properly released even if setup is incomplete
func (app *App) cleanup() {
	if app.repositories != nil {
		app.repositories.Close()
		app.logger.Info("repositories cleanup complete")
	}
}
