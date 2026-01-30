package app

import (
	"context"
	"library-service/config"
	jetstream2 "library-service/pkg/broker/nats/jetstream"
	"library-service/pkg/telemetry"
	"time"

	"github.com/nats-io/nats.go"
	natsJetstream "github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"

	"library-service/internal/cache"
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

	if err := app.initializeTracer(); err != nil {
		app.cleanup()
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

	//if err := app.initializeNATSServer(); err != nil {
	//	app.cleanup()
	//	return nil, err
	//}

	if app.configs.NATS.EnableJetStream {
		if err := app.initializeJetStream(); err != nil {
			app.cleanup()
			return nil, err
		}
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

func (app *App) initializeTracer() error {
	tracerShutdown, err := telemetry.InitTracer(context.Background(), telemetry.Config{
		ServiceName:    app.configs.Telemetry.ServiceName,
		ServiceVersion: app.configs.Telemetry.ServiceVersion,
		Environment:    app.configs.Telemetry.Environment,
		TempoEndpoint:  app.configs.Telemetry.TempoEndpoint,
		Enabled:        app.configs.Telemetry.Enabled,
	}, app.logger)
	if err != nil {
		app.logger.Error("tracer init error", zap.Error(err))
		return err
	}

	app.tracerShutdown = tracerShutdown
	app.logger.Info("tracer initialized",
		zap.Bool("enabled", app.configs.Telemetry.Enabled),
		zap.String("service", app.configs.Telemetry.ServiceName),
	)

	return nil
}

func (app *App) initializeRepositories() error {
	app.logger.Info("initializing repositories",
		zap.Duration("db startup delay", dbStartupDelay))
	time.Sleep(dbStartupDelay)

	configs := []repository.Configuration{
		repository.WithPostgresStore(app.configs.Store.DSN),
	}

	if app.configs.ClickHouse.DSN != "" {
		configs = append(configs, repository.WithClickHouseStore())
		app.logger.Info("clickhouse enabled", zap.String("dsn", app.configs.ClickHouse.DSN))
	}

	repositories, err := repository.New(configs...)
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
			Configs:      app.configs,
		},
		service.WithLibraryService(),
		service.WithSubscriptionService(),
		service.WithAuthService(),
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

//func (app *App) initializeNATSServer() error {
//	bookHandler := natsHandler.NewBookHandler(app.logger, app.services.Book)
//	authorHandler := natsHandler.NewAuthorHandler(app.logger, app.services.Author)
//	memberHandler := natsHandler.NewMemberHandler(app.logger, app.services.Member)
//
//	router := map[string]natsServer.CallHandler{
//		"health": app.natsHealthHandler,
//
//		// Book handlers
//		"book.get":          bookHandler.GetBook,
//		"book.list":         bookHandler.ListBooks,
//		"book.create":       bookHandler.CreateBook,
//		"book.update":       bookHandler.UpdateBook,
//		"book.delete":       bookHandler.DeleteBook,
//		"book.authors.list": bookHandler.ListBookAuthors,
//
//		// Author handlers
//		"author.get":    authorHandler.GetAuthor,
//		"author.list":   authorHandler.ListAuthors,
//		"author.create": authorHandler.CreateAuthor,
//		"author.update": authorHandler.UpdateAuthor,
//		"author.delete": authorHandler.DeleteAuthor,
//
//		// Member handlers
//		"member.get":    memberHandler.GetMember,
//		"member.list":   memberHandler.ListMembers,
//		"member.create": memberHandler.CreateMember,
//		"member.update": memberHandler.UpdateMember,
//		"member.delete": memberHandler.DeleteMember,
//	}
//
//	natsServerInstance, err := natsServer.New(
//		app.configs.NATS.URL,
//		app.configs.NATS.Subject,
//		router,
//	)
//	if err != nil {
//		app.logger.Error("nats server init error", zap.Error(err))
//		return err
//	}
//
//	app.natsServer = natsServerInstance
//	app.natsServer.Start()
//
//	app.logger.Info("nats server initialized",
//		zap.String("url", app.configs.NATS.URL),
//		zap.String("subject", app.configs.NATS.Subject),
//		zap.Int("handlers", len(router)),
//	)
//
//	return nil
//}

func (app *App) natsHealthHandler(msg *nats.Msg) (interface{}, error) {
	return map[string]string{
		"status":  "healthy",
		"service": "library-service",
	}, nil
}

func (app *App) initializeJetStream() error {
	js, err := jetstream2.New(jetstream2.Config{
		URL:           app.configs.NATS.URL,
		StreamName:    app.configs.NATS.StreamName,
		Subjects:      []string{"events.>"},
		MaxAge:        24 * time.Hour * 7,
		MaxBytes:      1024 * 1024 * 1024,
		Replicas:      1,
		StorageType:   natsJetstream.FileStorage,
		RetentionType: natsJetstream.LimitsPolicy,
	})
	if err != nil {
		app.logger.Error("jetstream init error", zap.Error(err))
		return err
	}

	app.jetStream = js
	app.eventPublisher = jetstream2.NewPublisher(js, app.logger, "library-service")
	app.eventConsumer = jetstream2.NewConsumer(js, app.logger)

	app.registerEventHandlers()

	go func() {
		ctx := context.Background()
		err := app.eventConsumer.Start(ctx, app.configs.NATS.StreamName, "library-consumer", []string{"events.>"})
		if err != nil {
			app.logger.Error("event consumer error", zap.Error(err))
		}
	}()

	app.logger.Info("jetstream initialized",
		zap.String("stream", app.configs.NATS.StreamName),
		zap.String("url", app.configs.NATS.URL),
	)

	return nil
}

func (app *App) registerEventHandlers() {
	app.eventConsumer.RegisterHandler("book.created", func(event jetstream2.Event) error {
		app.logger.Info("book created event",
			zap.String("event_id", event.ID),
			zap.Any("data", event.Data),
		)
		return nil
	})

	app.eventConsumer.RegisterHandler("book.updated", func(event jetstream2.Event) error {
		app.logger.Info("book updated event",
			zap.String("event_id", event.ID),
			zap.Any("data", event.Data),
		)
		return nil
	})

	app.eventConsumer.RegisterHandler("book.deleted", func(event jetstream2.Event) error {
		app.logger.Info("book deleted event",
			zap.String("event_id", event.ID),
			zap.Any("data", event.Data),
		)
		return nil
	})
}

func (app *App) cleanup() {
	if app.jetStream != nil {
		app.jetStream.Close()
		app.logger.Info("jetstream stopped")
	}

	if app.natsServer != nil {
		if err := app.natsServer.Shutdown(); err != nil {
			app.logger.Error("nats server shutdown error", zap.Error(err))
		} else {
			app.logger.Info("nats server stopped")
		}
	}

	if app.repositories != nil {
		app.repositories.Close()
		app.logger.Info("repositories cleanup complete")
	}
}
