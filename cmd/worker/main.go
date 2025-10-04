package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"library-service/internal/infrastructure/config"
	log "library-service/internal/infrastructure/logger"
)

// Worker handles background jobs and tasks
type Worker struct {
	logger *zap.Logger
	config *config.Config
}

func main() {
	logger := log.New()
	defer logger.Sync()

	logger.Info("starting worker service")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load configuration", zap.Error(err))
	}

	worker := &Worker{
		logger: logger,
		config: cfg,
	}

	// Start worker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start background tasks
	go worker.processJobs(ctx)
	go worker.cleanupExpiredData(ctx)

	logger.Info("worker service started")

	// Wait for shutdown signal
	sig := <-quit
	logger.Info("received shutdown signal", zap.String("signal", sig.String()))

	// Cancel context to stop all workers
	cancel()

	// Give workers time to finish
	time.Sleep(5 * time.Second)

	logger.Info("worker service stopped")
}

// processJobs processes background jobs from a queue
func (w *Worker) processJobs(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	w.logger.Info("job processor started")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("job processor stopping")
			return
		case <-ticker.C:
			// Process jobs from queue
			// In production, this would:
			// - Read from Redis queue or message broker
			// - Process tasks like sending emails, generating reports
			// - Update job status
			w.logger.Debug("processing jobs")
		}
	}
}

// cleanupExpiredData periodically cleans up expired data
func (w *Worker) cleanupExpiredData(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	w.logger.Info("cleanup task started")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("cleanup task stopping")
			return
		case <-ticker.C:
			w.logger.Info("running cleanup task")
			// Clean up expired:
			// - Cache entries
			// - Session data
			// - Temporary files
			// - Old audit logs
			fmt.Println("Cleanup task executed")
		}
	}
}
