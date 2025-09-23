// Package app provides the main application structure and lifecycle management
// for the library service. It handles initialization, startup, and graceful shutdown
// of all application components including servers, services, repositories, and caches.
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
	"library-service/internal/repository"
	"library-service/internal/service"
	"library-service/pkg/log"
	"library-service/pkg/server"
)

// App holds application-wide dependencies and lifecycle management.
// It encapsulates all the components needed to run the library service
// and provides methods for initialization, startup, and shutdown.
type App struct {
	logger       *zap.Logger
	configs      *config.Configs
	repositories *repository.Repositories
	caches       *cache.Caches
	services     *service.Services
	servers      *server.Servers
	handlers     *handler.Handlers
}

// Run initializes and runs the application with proper error handling
// and graceful shutdown capabilities. It serves as the main entry point
// for the application lifecycle.
func Run() {
	logger := log.GetLogger()
	app, err := initApp(logger)
	if err != nil {
		logger.Error("app init error", zap.Error(err))
		return
	}

	// Start all configured servers
	if err := app.startServers(); err != nil {
		app.logger.Error("server start error", zap.Error(err))
		app.shutdown()
		return
	}

	app.logStartupInfo()

	// Wait for termination signal and shutdown gracefully
	wait := parseGracefulTimeout()
	app.waitForShutdown(wait)
}

// startServers initializes and starts all configured servers in background goroutines.
// It ensures servers are ready to accept connections and logs startup information.
func (app *App) startServers() error {
	if err := app.servers.Run(app.logger); err != nil {
		app.logger.Error("server startup failed",
			zap.Error(err),
			zap.String("port", app.configs.APP.Port),
		)
		return err
	}

	app.logServerStarted()
	return nil
}

// logServerStarted logs detailed information about started servers
// including addresses and useful endpoints for development and monitoring
func (app *App) logServerStarted() {
	port := app.configs.APP.Port
	baseURL := fmt.Sprintf("http://localhost%s", port)

	app.logger.Info("http server started",
		zap.String("address", baseURL),
		zap.String("port", port),
		zap.String("mode", app.configs.APP.Mode),
	)

	// Log useful development endpoints
	if app.configs.APP.Mode == "dev" {
		app.logger.Info("development endpoints",
			zap.String("swagger", fmt.Sprintf("%s/swagger/index.html", baseURL)),
			zap.String("health", fmt.Sprintf("%s/health", baseURL)),
			zap.String("metrics", fmt.Sprintf("%s/metrics", baseURL)),
		)
	}
}

// logStartupInfo logs application startup information including
// server addresses and useful links for development
func (app *App) logStartupInfo() {
	app.logger.Info("application started",
		zap.String("time", time.Now().Format("02.01.2006 15:04:05")),
		zap.String("mode", app.configs.APP.Mode),
		zap.String("swagger", fmt.Sprintf("http://localhost%s/swagger/index.html", app.configs.APP.Port)),
		zap.String("health", fmt.Sprintf("http://localhost%s/health", app.configs.APP.Port)),
	)
}

// waitForShutdown waits for OS termination signals and triggers graceful shutdown.
// It handles SIGINT and SIGTERM signals to allow for clean application termination.
func (app *App) waitForShutdown(timeout time.Duration) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	sig := <-quit
	app.logger.Info("shutdown signal received",
		zap.String("signal", sig.String()),
		zap.Duration("timeout", timeout),
	)

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Stop servers with timeout
	if err := app.servers.Stop(ctx); err != nil {
		app.logger.Error("server shutdown error", zap.Error(err))
	} else {
		app.logger.Info("server stopped gracefully")
	}

	app.shutdown()
}

// shutdown performs cleanup of all application resources.
// It ensures proper resource cleanup and logging synchronization.
func (app *App) shutdown() {
	app.logger.Info("running cleanup tasks")

	// Close repositories and database connections
	if app.repositories != nil {
		app.repositories.Close()
		app.logger.Info("repositories closed")
	}

	// Flush logger buffers (best-effort)
	if app.logger != nil {
		_ = app.logger.Sync()
	}

	app.logger.Info("application shutdown complete")
}

// parseGracefulTimeout reads the graceful-timeout flag from command line
// or returns the default timeout for server shutdown operations
func parseGracefulTimeout() time.Duration {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", 15*time.Second,
		"duration for which the server waits for existing connections to finish")
	flag.Parse()
	return wait
}
