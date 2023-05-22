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

	"library/config"
	"library/internal/repository"
	"library/internal/service"
	"library/internal/transport/http"
	"library/pkg/log"
)

const (
	version     = "1.0.0"
	description = "library-service"
)

// Run initializes whole application.
func Run() {
	// Dependencies
	logger := log.New(version, description)

	configs, err := config.New()
	if err != nil {
		logger.Error("ERR_INIT_CONFIG", zap.Error(err))
		return
	}

	// Repositories, Services & Handlers
	repositories, err := repository.New(repository.WithMemoryRepository())
	if err != nil {
		logger.Error("ERR_INIT_REPOSITORY", zap.Error(err))
		return
	}

	services := service.New(service.Dependencies{
		AuthorRepository: repositories.Author,
		BookRepository:   repositories.Book,
		MemberRepository: repositories.Member,
	})

	httpServer := http.New(http.Dependencies{
		Configs:       configs,
		AuthorService: services.Author,
		BookService:   services.Book,
		MemberService: services.Member,
	})

	// Run our httpServer in a goroutine so that it doesn't block.
	go func() {
		// service connections
		if err = httpServer.Run(); err != nil {
			logger.Error("ERR_INIT_REST", zap.Error(err))
			return
		}
	}()
	logger.Info("Server started on http://localhost:" + configs.HTTP.Port)

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
	if err = httpServer.Stop(ctx); err != nil {
		panic(err) // failure/timeout shutting down the httpServer gracefully
	}

	fmt.Println("Running cleanup tasks...")
	// Your cleanup tasks go here

	fmt.Println("Server was successful shutdown.")
}
