package main

import (
	"context"
	"fmt"
	config3 "library-service/internal/infrastructure/config"
	"library-service/internal/payments/provider/epayment"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"library-service/internal/container"
	domainapp "library-service/internal/domain/app"
	"library-service/internal/infrastructure/auth"
	"library-service/internal/infrastructure/log"
	paymentops "library-service/internal/payments/service/payment"
)

// Worker handles background jobs and tasks
type Worker struct {
	logger                   *zap.Logger
	config                   *config3.Config
	usecases                 *container.Container
	expirePaymentsUC         *paymentops.ExpirePaymentsUseCase
	processCallbackRetriesUC *paymentops.ProcessCallbackRetriesUseCase
}

// Validator wraps go-playground/validator
type Validator struct {
	validate *validator.Validate
}

// Validate validates a struct
func (v *Validator) Validate(i interface{}) error {
	if v.validate == nil {
		v.validate = validator.New()
	}
	return v.validate.Struct(i)
}

func main() {
	logger := log.New()
	defer logger.Sync()

	logger.Info("starting worker service")

	// Load configuration
	cfg := config3.MustLoad("")

	// Initialize repositories
	repos, err := domainapp.NewRepositories(domainapp.WithMemoryStore())
	if err != nil {
		logger.Fatal("failed to initialize repositories", zap.Error(err))
	}
	logger.Info("repositories initialized")

	// Initialize caches
	caches, err := domainapp.NewCaches(
		domainapp.Dependencies{Repositories: repos},
		domainapp.WithMemoryCache(),
	)
	if err != nil {
		logger.Fatal("failed to initialize caches", zap.Error(err))
	}
	logger.Info("caches initialized")

	// Initialize auth service (minimal setup for worker)
	authServices := &container.AuthServices{
		JWTService: auth.NewJWTService(
			cfg.JWT.Secret,
			cfg.JWT.AccessTokenTTL,
			cfg.JWT.RefreshTokenTTL,
			cfg.JWT.Issuer,
		),
		PasswordService: auth.NewPasswordService(),
	}
	logger.Info("auth service initialized")

	// Initialize payment provider
	epaymentConfig, err := epayment.LoadConfigFromEnv()
	if err != nil {
		logger.Warn("failed to load epayment config, payment features may not work", zap.Error(err))
	}
	paymentGateway := epayment.NewGateway(epaymentConfig, logger)

	gatewayServices := &container.GatewayServices{
		PaymentGateway: paymentGateway,
	}
	logger.Info("provider service initialized")

	// Initialize validator
	validator := &Validator{}
	logger.Info("validator initialized")

	// Initialize use cases container
	usecaseRepos := &container.Repositories{
		Book:        repos.Book,
		Author:      repos.Author,
		Member:      repos.Member,
		Reservation: repos.Reservation,
		Payment:     repos.Payment,
		SavedCard:   repos.SavedCard,
	}
	usecaseCaches := &container.Caches{
		Book:   caches.Book,
		Author: caches.Author,
	}
	usecases := container.NewContainer(usecaseRepos, usecaseCaches, authServices, gatewayServices, validator)

	worker := &Worker{
		logger:                   logger,
		config:                   cfg,
		usecases:                 usecases,
		expirePaymentsUC:         usecases.Payment.ExpirePayments,
		processCallbackRetriesUC: usecases.Payment.ProcessCallbackRetries,
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
	go worker.expirePayments(ctx)
	go worker.processCallbackRetries(ctx)

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

// expirePayments periodically expires old pending/processing payments
func (w *Worker) expirePayments(ctx context.Context) {
	// Run immediately on startup
	w.runPaymentExpiryJob(ctx)

	// Then run every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	w.logger.Info("payment expiry task started")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("payment expiry task stopping")
			return
		case <-ticker.C:
			w.runPaymentExpiryJob(ctx)
		}
	}
}

// runPaymentExpiryJob executes the payment expiry job
func (w *Worker) runPaymentExpiryJob(ctx context.Context) {
	w.logger.Info("running payment expiry job")

	// Create a context with timeout for this specific job
	jobCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// Execute expiry use case
	result, err := w.expirePaymentsUC.Execute(jobCtx, paymentops.ExpirePaymentsRequest{
		BatchSize: 100, // Process max 100 payments per run
	})

	if err != nil {
		w.logger.Error("payment expiry job failed", zap.Error(err))
		return
	}

	if result.ExpiredCount > 0 || result.FailedCount > 0 {
		w.logger.Info("payment expiry job completed",
			zap.Int("expired_count", result.ExpiredCount),
			zap.Int("failed_count", result.FailedCount),
			zap.Int("error_count", len(result.Errors)),
		)
	}

	// Log errors if any
	for i, errMsg := range result.Errors {
		if i < 5 { // Log first 5 errors only
			w.logger.Error("payment expiry error", zap.String("error", errMsg))
		}
	}
}

// processCallbackRetries periodically processes pending callback retries
func (w *Worker) processCallbackRetries(ctx context.Context) {
	// Run immediately on startup
	w.runCallbackRetryJob(ctx)

	// Then run every 2 minutes
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	w.logger.Info("callback retry task started")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("callback retry task stopping")
			return
		case <-ticker.C:
			w.runCallbackRetryJob(ctx)
		}
	}
}

// runCallbackRetryJob executes the callback retry job
func (w *Worker) runCallbackRetryJob(ctx context.Context) {
	w.logger.Info("running callback retry job")

	// Create a context with timeout for this specific job
	jobCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	// Execute callback retry use case
	result, err := w.processCallbackRetriesUC.Execute(jobCtx, paymentops.ProcessCallbackRetriesRequest{
		BatchSize: 50, // Process max 50 retries per run
	})

	if err != nil {
		w.logger.Error("callback retry job failed", zap.Error(err))
		return
	}

	if result.ProcessedCount > 0 {
		w.logger.Info("callback retry job completed",
			zap.Int("processed_count", result.ProcessedCount),
			zap.Int("success_count", result.SuccessCount),
			zap.Int("failed_count", result.FailedCount),
			zap.Int("error_count", len(result.Errors)),
		)
	}

	// Log errors if any
	for i, errMsg := range result.Errors {
		if i < 5 { // Log first 5 errors only
			w.logger.Warn("callback retry error", zap.String("error", errMsg))
		}
	}
}
