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

	"library/internal/cache"
	"library/internal/config"
	"library/internal/handler"
	"library/internal/repository"
	"library/internal/service/library"
	"library/internal/service/subscription"
	"library/pkg/log"
	"library/pkg/server"
)

const (
	version     = "1.0.0"
	description = "library-service"
)

// Run initializes whole application.
func Run() {
	logger := log.New(version, description)

	cfg, err := config.New()
	if err != nil {
		logger.Error("ERR_INIT_CONFIG", zap.Error(err))
		return
	}

	repositories, err := repository.New(
		repository.WithMemoryStore())
	if err != nil {
		logger.Error("ERR_INIT_REPOSITORY", zap.Error(err))
		return
	}
	defer repositories.Close()

	caches, err := cache.New(
		cache.Dependencies{
			AuthorRepository: repositories.Author,
			BookRepository:   repositories.Book,
		},
		cache.WithMemoryStore())
	if err != nil {
		logger.Error("ERR_INIT_CACHE", zap.Error(err))
		return
	}
	defer caches.Close()

	libraryService, err := library.New(
		library.WithAuthorRepository(repositories.Author),
		library.WithBookRepository(repositories.Book),
		library.WithAuthorCache(caches.Author),
		library.WithBookCache(caches.Book))
	if err != nil {
		logger.Error("ERR_INIT_LIBRARY_SERVICE", zap.Error(err))
		return
	}

	subscriptionService, err := subscription.New(
		subscription.WithMemberRepository(repositories.Member),
		subscription.WithLibraryService(libraryService))
	if err != nil {
		logger.Error("ERR_INIT_SUBSCRIPTION_SERVICE", zap.Error(err))
		return
	}

	handlers, err := handler.New(
		handler.Dependencies{
			LibraryService:      libraryService,
			SubscriptionService: subscriptionService,
		},
		handler.WithHTTPHandler())
	if err != nil {
		logger.Error("ERR_INIT_HANDLER", zap.Error(err))
		return
	}

	servers, err := server.New(
		server.WithHTTPServer(handlers.HTTP, cfg.HTTP.Port))
	if err != nil {
		logger.Error("ERR_INIT_SERVER", zap.Error(err))
		return
	}

	// Run our server in a goroutine so that it doesn't block.
	if err = servers.Run(logger); err != nil {
		logger.Error("ERR_RUN_SERVER", zap.Error(err))
		return
	}

	// Graceful Shutdown
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the httpServer gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	quit := make(chan os.Signal, 1) // create channel to signify a signal being sent

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM) // When an interrupt or termination signal is sent, notify the channel
	<-quit                                             // This blocks the main thread until an interrupt is received
	fmt.Println("Gracefully shutting down...")

	// create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	if err = servers.Stop(ctx); err != nil {
		panic(err) // failure/timeout shutting down the httpServer gracefully
	}

	fmt.Println("Running cleanup tasks...")
	// Your cleanup tasks go here

	fmt.Println("Server was successful shutdown.")
}
