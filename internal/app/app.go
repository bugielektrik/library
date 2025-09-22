package app

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"library-service/internal/cache"
	"library-service/internal/config"
	"library-service/internal/handler"
	"library-service/internal/provider/epay"
	"library-service/internal/repository"
	"library-service/internal/service/library"
	"library-service/internal/service/payment"
	"library-service/internal/service/subscription"
	"library-service/pkg/log"
	"library-service/pkg/server"
)

// App holds application-wide dependencies and lifecycle management.
type App struct {
	logger    *zap.Logger
	configs   *config.Configs
	providers struct {
		epay *epay.Client
	}
	repositories *repository.Repositories
	caches       *cache.Caches
	services     struct {
		payment      *payment.Service
		library      *library.Service
		subscription *subscription.Service
	}
	servers  *server.Servers
	handlers *handler.Handlers
}

// Run initializes and runs the application.
func Run() {
	logger := log.GetLogger()
	app, err := newApp(logger)
	if err != nil {
		logger.Error("app_init_error", zap.Error(err))
		return
	}

	// Start servers and handle errors
	if err := app.startServers(); err != nil {
		app.logger.Error("server_start_error", zap.Error(err))
		app.shutdown()
		return
	}

	// Start cron jobs if any
	app.logger.Info("application_started",
		zap.String("time", time.Now().Format("02.01.2006 15:04:05")),
		zap.String("swagger", fmt.Sprintf("http://localhost%s/swagger/index.html", app.configs.APP.Port)),
	)

	// Wait for termination signal and then shutdown gracefully
	wait := parseGracefulTimeout()
	app.waitForShutdown(wait)
}

// newApp builds the application with its dependencies.
// It returns an initialized App or an error if initialization failed.
func newApp(logger *zap.Logger) (app *App, err error) {
	app = &App{logger: logger}

	// Load configuration
	app.configs, err = config.New()
	if err != nil {
		return nil, err
	}

	// Initialize providers/clients
	if err := app.initProviders(); err != nil {
		return nil, err
	}

	// Initialize repositories (DB)
	// small backoff to allow DB to be up (kept as original behavior)
	app.logger.Info("waiting_for_db", zap.Duration("delay", time.Second))
	time.Sleep(time.Second)

	app.repositories, err = repository.New(
		repository.WithPostgresStore(app.configs.POSTGRES.DSN),
	)
	if err != nil {
		return nil, err
	}

	// Initialize caches (in-memory, redis, etc.)
	app.caches, err = cache.New(
		cache.Dependencies{
			AuthorRepository: app.repositories.Author,
			BookRepository:   app.repositories.Book,
		},
		cache.WithMemoryStore())
	if err != nil {
		logger.Error("ERR_INIT_CACHES", zap.Error(err))
		return
	}

	// Initialize services (business logic)
	if err := app.initServices(); err != nil {
		app.repositories.Close()
		return nil, err
	}

	// Initialize handlers (HTTP, gRPC, etc.)
	app.handlers, err = handler.New(
		handler.Dependencies{
			Configs:             app.configs,
			PaymentService:      app.services.payment,
			LibraryService:      app.services.library,
			SubscriptionService: app.services.subscription,
		},
		handler.WithHTTPHandler(),
	)
	if err != nil {
		app.repositories.Close()
		return nil, err
	}

	// Initialize HTTP server (or other servers)
	app.servers, err = server.NewServer(server.WithHTTP(app.handlers.HTTP, app.configs.APP.Port))
	if err != nil {
		app.repositories.Close()
		return nil, err
	}

	return app, nil
}

// initProviders initializes external providers/clients.
func (app *App) initProviders() error {
	cfg := app.configs

	app.providers.epay = epay.New(cfg, epay.Credentials{
		Username: cfg.EPAY.Login,
		Password: cfg.EPAY.Password,
		Endpoint: cfg.EPAY.URL,
		OAuth:    cfg.EPAY.OAuth,
		JS:       cfg.EPAY.JS,
	})

	// Only init token refresher outside of dev mode
	if cfg.APP.Mode != "dev" {
		if err := app.providers.epay.InitTokenRefresher(); err != nil {
			app.logger.Error("epay_token_refresher_init_error", zap.Error(err))
			return err
		}
	}

	return nil
}

// initServices composes domain services from repos and providers.
func (app *App) initServices() (err error) {
	app.services.payment, err = payment.New(
		payment.WithEpayClient(app.providers.epay))
	if err != nil {
		app.logger.Error("payment_service_init_error", zap.Error(err))
		return err
	}

	app.services.library, err = library.New(
		library.WithAuthorRepository(app.repositories.Author),
		library.WithBookRepository(app.repositories.Book),
		library.WithAuthorCache(app.caches.Author),
		library.WithBookCache(app.caches.Book))
	if err != nil {
		app.logger.Error("library_service_init_error", zap.Error(err))
		return err
	}

	app.services.subscription, err = subscription.New(
		subscription.WithMemberRepository(app.repositories.Member),
		subscription.WithLibraryService(app.services.library))
	if err != nil {
		app.logger.Error("subscription_service_init_error", zap.Error(err))
		return
	}

	return nil
}

// startServers runs the configured servers in background goroutines.
func (app *App) startServers() error {
	if err := app.servers.Run(app.logger); err != nil {
		return err
	}
	app.logger.Info("http_server_started", zap.String("addr", fmt.Sprintf("http://localhost%s", app.configs.APP.Port)))
	return nil
}

// waitForShutdown waits for OS signals and triggers graceful shutdown.
func (app *App) waitForShutdown(timeout time.Duration) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	sig := <-quit
	app.logger.Info("shutdown_signal_received", zap.String("signal", sig.String()))

	// Context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := app.servers.Stop(ctx); err != nil {
		app.logger.Error("server_shutdown_error", zap.Error(err))
	} else {
		app.logger.Info("server_stopped_gracefully")
	}

	app.shutdown()
}

// shutdown runs cleanup logic and releases resources.
func (app *App) shutdown() {
	app.logger.Info("running_cleanup_tasks")

	// close repositories
	if app.repositories != nil {
		app.repositories.Close()
		app.logger.Info("repositories_closed")
	}

	// final flush/cleanup for logger (best-effort)
	if app.logger != nil {
		_ = app.logger.Sync()
	}

	app.logger.Info("server_successfully_shutdown")
}

// parseGracefulTimeout reads the graceful-timeout flag or uses default.
func parseGracefulTimeout() time.Duration {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", 15*time.Second, "duration for which the server waits for existing connections to finish")
	flag.Parse()
	return wait
}
