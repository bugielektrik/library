package app

import (
	"context"
	"flag"
	"fmt"
	"library-service/config"
	"library-service/internal/provider/epay"
	jetstream2 "library-service/pkg/broker/nats/jetstream"
	natsServer "library-service/pkg/broker/nats/nats_rpc/server"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"library-service/internal/cache"
	"library-service/internal/handler"
	"library-service/internal/repository"
	"library-service/internal/service"
	"library-service/pkg/log"
	"library-service/pkg/server"
)

type App struct {
	logger         *zap.Logger
	configs        *config.Configs
	repositories   *repository.Repositories
	caches         *cache.Caches
	services       *service.Services
	servers        *server.Servers
	handlers       *handler.Handlers
	natsServer     *natsServer.Server
	jetStream      *jetstream2.JetStream
	eventPublisher *jetstream2.Publisher
	eventConsumer  *jetstream2.Consumer
}

func (app *App) initEPAYClient() {
	_ = epay.New(*app.configs, epay.Credentials{
		Username: "",
		Password: "",
		Endpoint: "",
		OAuth:    "",
		JS:       "",
	})
}
func Run() {
	logger := log.GetLogger()
	app, err := initApp(logger)
	if err != nil {
		logger.Error("app init error", zap.Error(err))
		return
	}

	if err := app.startServers(); err != nil {
		app.logger.Error("server start error", zap.Error(err))
		app.shutdown()
		return
	}

	app.logStartupInfo()

	wait := parseGracefulTimeout()
	app.waitForShutdown(wait)
}

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

func (app *App) logServerStarted() {
	port := app.configs.APP.Port
	baseURL := fmt.Sprintf("http://localhost%s", port)

	app.logger.Info("http server started",
		zap.String("address", baseURL),
		zap.String("port", port),
		zap.String("mode", app.configs.APP.Mode),
	)

	if app.configs.APP.Mode == "dev" {
		app.logger.Info("development endpoints",
			zap.String("swagger", fmt.Sprintf("%s/swagger/index.html", baseURL)),
			zap.String("health", fmt.Sprintf("%s/health", baseURL)),
			zap.String("metrics", fmt.Sprintf("%s/metrics", baseURL)),
		)
	}
}

func (app *App) logStartupInfo() {
	app.logger.Info("application started",
		zap.String("time", time.Now().Format("02.01.2006 15:04:05")),
		zap.String("mode", app.configs.APP.Mode),
		zap.String("swagger", fmt.Sprintf("http://localhost%s/swagger/index.html", app.configs.APP.Port)),
		zap.String("health", fmt.Sprintf("http://localhost%s/health", app.configs.APP.Port)),
	)
}

func (app *App) waitForShutdown(timeout time.Duration) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	sig := <-quit
	app.logger.Info("shutdown signal received",
		zap.String("signal", sig.String()),
		zap.Duration("timeout", timeout),
	)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := app.servers.Stop(ctx); err != nil {
		app.logger.Error("server shutdown error", zap.Error(err))
	} else {
		app.logger.Info("server stopped gracefully")
	}

	app.shutdown()
}

func (app *App) shutdown() {
	app.logger.Info("running cleanup tasks")

	if app.natsServer != nil {
		if err := app.natsServer.Shutdown(); err != nil {
			app.logger.Error("nats server shutdown error", zap.Error(err))
		} else {
			app.logger.Info("nats server stopped")
		}
	}

	if app.repositories != nil {
		app.repositories.Close()
		app.logger.Info("repositories closed")
	}

	if app.logger != nil {
		_ = app.logger.Sync()
	}

	app.logger.Info("application shutdown complete")
}

func parseGracefulTimeout() time.Duration {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", 15*time.Second,
		"duration for which the server waits for existing connections to finish")
	flag.Parse()
	return wait
}
