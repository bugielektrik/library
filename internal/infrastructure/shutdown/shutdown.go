// Package shutdown provides enhanced graceful shutdown with phased shutdown and hooks
package shutdown

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Phase represents a shutdown phase
type Phase string

const (
	// PhasePreShutdown runs before shutdown begins (e.g., health check updates)
	PhasePreShutdown Phase = "pre_shutdown"

	// PhaseStopAcceptingRequests stops accepting new connections
	PhaseStopAcceptingRequests Phase = "stop_accepting_requests"

	// PhaseDrainConnections waits for in-flight requests to complete
	PhaseDrainConnections Phase = "drain_connections"

	// PhaseCleanup closes external connections (DB, cache, etc.)
	PhaseCleanup Phase = "cleanup"

	// PhasePostShutdown runs final cleanup tasks
	PhasePostShutdown Phase = "post_shutdown"
)

// Hook is a function that runs during a specific shutdown phase
type Hook func(ctx context.Context) error

// Manager manages graceful shutdown with phased execution
type Manager struct {
	logger *zap.Logger
	phases map[Phase][]Hook
	mu     sync.RWMutex
}

// NewManager creates a new shutdown manager
func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		logger: logger,
		phases: make(map[Phase][]Hook),
	}
}

// RegisterHook registers a shutdown hook for a specific phase
func (m *Manager) RegisterHook(phase Phase, name string, hook Hook) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Wrap hook with logging
	wrappedHook := func(ctx context.Context) error {
		m.logger.Info("executing shutdown hook",
			zap.String("phase", string(phase)),
			zap.String("hook", name),
		)

		start := time.Now()
		err := hook(ctx)
		duration := time.Since(start)

		if err != nil {
			m.logger.Error("shutdown hook failed",
				zap.String("phase", string(phase)),
				zap.String("hook", name),
				zap.Duration("duration", duration),
				zap.Error(err),
			)
			return fmt.Errorf("hook %s failed: %w", name, err)
		}

		m.logger.Info("shutdown hook completed",
			zap.String("phase", string(phase)),
			zap.String("hook", name),
			zap.Duration("duration", duration),
		)
		return nil
	}

	m.phases[phase] = append(m.phases[phase], wrappedHook)
}

// Shutdown executes all shutdown hooks in order with phased timeouts
func (m *Manager) Shutdown(ctx context.Context) error {
	m.logger.Info("starting graceful shutdown")
	startTime := time.Now()

	// Define phases with their timeouts
	phasesWithTimeouts := []struct {
		phase   Phase
		timeout time.Duration
	}{
		{PhasePreShutdown, 2 * time.Second},
		{PhaseStopAcceptingRequests, 1 * time.Second},
		{PhaseDrainConnections, 10 * time.Second},
		{PhaseCleanup, 5 * time.Second},
		{PhasePostShutdown, 2 * time.Second},
	}

	var shutdownErrors []error

	// Execute each phase
	for _, pt := range phasesWithTimeouts {
		if err := m.executePhase(ctx, pt.phase, pt.timeout); err != nil {
			m.logger.Error("shutdown phase failed",
				zap.String("phase", string(pt.phase)),
				zap.Error(err),
			)
			shutdownErrors = append(shutdownErrors, err)
			// Continue with other phases even if one fails
		}
	}

	totalDuration := time.Since(startTime)
	m.logger.Info("graceful shutdown completed",
		zap.Duration("total_duration", totalDuration),
		zap.Int("error_count", len(shutdownErrors)),
	)

	if len(shutdownErrors) > 0 {
		return fmt.Errorf("shutdown completed with %d errors", len(shutdownErrors))
	}

	return nil
}

// executePhase executes all hooks for a specific phase with timeout
func (m *Manager) executePhase(parentCtx context.Context, phase Phase, timeout time.Duration) error {
	m.mu.RLock()
	hooks := m.phases[phase]
	m.mu.RUnlock()

	if len(hooks) == 0 {
		m.logger.Debug("no hooks registered for phase", zap.String("phase", string(phase)))
		return nil
	}

	m.logger.Info("executing shutdown phase",
		zap.String("phase", string(phase)),
		zap.Int("hook_count", len(hooks)),
		zap.Duration("timeout", timeout),
	)

	// Create context with timeout for this phase
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	// Execute hooks in parallel with error collection
	var wg sync.WaitGroup
	errChan := make(chan error, len(hooks))

	for _, hook := range hooks {
		wg.Add(1)
		go func(h Hook) {
			defer wg.Done()
			if err := h(ctx); err != nil {
				errChan <- err
			}
		}(hook)
	}

	// Wait for all hooks to complete or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All hooks completed
		close(errChan)

		// Collect errors
		var errors []error
		for err := range errChan {
			errors = append(errors, err)
		}

		if len(errors) > 0 {
			return fmt.Errorf("phase %s: %d hooks failed", phase, len(errors))
		}

		return nil

	case <-ctx.Done():
		// Timeout or cancellation
		m.logger.Warn("shutdown phase timed out",
			zap.String("phase", string(phase)),
			zap.Duration("timeout", timeout),
		)
		return fmt.Errorf("phase %s timed out after %s", phase, timeout)
	}
}

// RegisterDefaultHooks registers common shutdown hooks
func (m *Manager) RegisterDefaultHooks(server ShutdownableServer, repos ShutdownableRepos) {
	// Pre-shutdown: Mark service as unhealthy
	m.RegisterHook(PhasePreShutdown, "mark_unhealthy", func(ctx context.Context) error {
		m.logger.Info("marking service as unhealthy")
		// In a real system, this would update health check endpoint
		return nil
	})

	// Stop accepting requests: Stop HTTP server from accepting new connections
	if server != nil {
		m.RegisterHook(PhaseStopAcceptingRequests, "stop_http_server", func(ctx context.Context) error {
			return server.Shutdown(ctx)
		})
	}

	// Drain connections: Already handled by HTTP server's Shutdown method

	// Cleanup: Close database and cache connections
	if repos != nil {
		m.RegisterHook(PhaseCleanup, "close_repositories", func(ctx context.Context) error {
			repos.Close()
			return nil
		})
	}

	// Post-shutdown: Flush logs and final cleanup
	m.RegisterHook(PhasePostShutdown, "flush_logs", func(ctx context.Context) error {
		if m.logger != nil {
			_ = m.logger.Sync() // Ignore error as it's common on stdout/stderr
		}
		return nil
	})
}

// ShutdownableServer interface for components that need shutdown
type ShutdownableServer interface {
	Shutdown(ctx context.Context) error
}

// ShutdownableRepos interface for repositories that need cleanup
type ShutdownableRepos interface {
	Close()
}
