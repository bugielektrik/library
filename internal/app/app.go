package app

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"library/internal/config"
	"library/internal/service/library"
	"library/internal/transport/rest"
	"library/pkg/database"
)

const (
	version     = "1.0.0"
	description = "library-service"
)

// Run initializes whole application.
func Run() {
	// Dependencies
	logger := setupLogger()

	cfg, err := config.Init()
	if err != nil {
		logger.Error("ERR_INIT_CONFIG", zap.Error(err))
		return
	}

	postgres, err := database.New(cfg.POSTGRES.URL)
	if err != nil {
		logger.Error("ERR_INIT_POSTGRES", zap.Error(err))
		return
	}
	defer postgres.Close()

	err = database.Migrate(cfg.POSTGRES.URL)
	if err != nil {
		logger.Error("ERR_MIGRATE_POSTGRES", zap.Error(err))
		return
	}

	// Stores, Services & API Handlers
	libraryService := library.NewService(
		library.WithPostgresRepository(postgres),
	)
	handler := rest.NewHandler(libraryService).Init()

	// Run our server in a goroutine so that it doesn't block.
	server := &http.Server{
		Addr:    ":" + cfg.HTTP.Port,
		Handler: handler,
	}

	go func() {
		// service connections
		if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	logger.Info("Server started on http://localhost:" + cfg.HTTP.Port)

	// Graceful Shutdown
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	quit := make(chan os.Signal, 1) // Create channel to signify a signal being sent

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM) // When an interrupt or termination signal is sent, notify the channel
	<-quit                                             // This blocks the main thread until an interrupt is received
	fmt.Println("Gracefully shutting down...")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	if err = server.Shutdown(ctx); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}

	fmt.Println("Running cleanup tasks...")
	// Your cleanup tasks go here

	fmt.Println("Server was successful shutdown.")
}
