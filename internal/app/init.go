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
	dbStartupDelay = time.Second
)

func initApp(logger *zap.Logger) (*App, error) {
	app := &App{logger: logger}

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

func (app *App) initializeRepositories() error {
	app.logger.Info("initializing repositories",
		zap.Duration("db startup delay", dbStartupDelay))
	time.Sleep(dbStartupDelay)

	repositories, err := repository.New(
		repository.WithPostgresStore(app.configs.Store.DSN),
	)
	if err != nil {
		app.logger.Error("repository init error", zap.Error(err))
		return err
	}

	app.repositories = repositories
	app.logger.Info("repositories initialized")
	return nil
}

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

func (app *App) cleanup() {
	if app.repositories != nil {
		app.repositories.Close()
		app.logger.Info("repositories cleanup complete")
	}
}
