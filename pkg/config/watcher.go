package config

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

// Watcher watches configuration files for changes
type Watcher struct {
	config     *Config
	configPath string
	loader     *Loader
	validator  *ConfigValidator
	logger     *zap.Logger
	mu         sync.RWMutex
	callbacks  []func(*Config)
	watcher    *fsnotify.Watcher
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewWatcher creates a new configuration watcher
func NewWatcher(configPath string, logger *zap.Logger) (*Watcher, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating file watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	w := &Watcher{
		configPath: configPath,
		loader:     NewLoader(),
		validator:  NewValidator(),
		logger:     logger,
		callbacks:  []func(*Config){},
		watcher:    fsWatcher,
		ctx:        ctx,
		cancel:     cancel,
	}

	// Load initial configuration
	if err := w.reload(); err != nil {
		fsWatcher.Close()
		return nil, fmt.Errorf("loading initial config: %w", err)
	}

	// Add config file to watcher
	if err := fsWatcher.Add(configPath); err != nil {
		fsWatcher.Close()
		return nil, fmt.Errorf("watching config file: %w", err)
	}

	// Start watching
	go w.watch()

	return w, nil
}

// Get returns the current configuration
func (w *Watcher) Get() *Config {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.config
}

// OnChange registers a callback to be called when configuration changes
func (w *Watcher) OnChange(callback func(*Config)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.callbacks = append(w.callbacks, callback)
}

// reload reloads the configuration from file
func (w *Watcher) reload() error {
	newConfig, err := w.loader.Load(w.configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Validate new configuration
	if err := w.validator.Validate(newConfig); err != nil {
		return fmt.Errorf("validating config: %w", err)
	}

	// Update configuration
	w.mu.Lock()
	oldConfig := w.config
	w.config = newConfig
	callbacks := w.callbacks
	w.mu.Unlock()

	// Call callbacks if config changed
	if oldConfig != nil && !configEqual(oldConfig, newConfig) {
		w.logger.Info("configuration reloaded",
			zap.String("path", w.configPath),
			zap.Any("changes", getConfigChanges(oldConfig, newConfig)),
		)

		for _, callback := range callbacks {
			go callback(newConfig)
		}
	}

	return nil
}

// watch watches for configuration file changes
func (w *Watcher) watch() {
	debounce := time.NewTimer(0)
	debounce.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Check if it's a write or create event
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				w.logger.Debug("config file changed",
					zap.String("file", event.Name),
					zap.String("operation", event.Op.String()),
				)

				// Debounce rapid changes
				debounce.Stop()
				debounce = time.AfterFunc(100*time.Millisecond, func() {
					if err := w.reload(); err != nil {
						w.logger.Error("failed to reload config",
							zap.String("file", w.configPath),
							zap.Error(err),
						)
					}
				})
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			w.logger.Error("config watcher error", zap.Error(err))
		}
	}
}

// Close stops watching and cleans up resources
func (w *Watcher) Close() error {
	w.cancel()
	return w.watcher.Close()
}

// configEqual compares two configurations for equality
func configEqual(a, b *Config) bool {
	// Simple comparison - can be enhanced
	return fmt.Sprintf("%+v", a) == fmt.Sprintf("%+v", b)
}

// getConfigChanges returns the differences between two configurations
func getConfigChanges(old, new *Config) map[string]interface{} {
	changes := make(map[string]interface{})

	// Check app config
	if old.App.Environment != new.App.Environment || old.App.Debug != new.App.Debug {
		changes["app"] = map[string]interface{}{
			"old": map[string]interface{}{
				"environment": old.App.Environment,
				"debug":       old.App.Debug,
			},
			"new": map[string]interface{}{
				"environment": new.App.Environment,
				"debug":       new.App.Debug,
			},
		}
	}

	// Check server config (compare specific fields since struct contains slices)
	if old.Server.Host != new.Server.Host || old.Server.Port != new.Server.Port {
		changes["server"] = map[string]interface{}{
			"host": map[string]interface{}{
				"old": fmt.Sprintf("%s:%d", old.Server.Host, old.Server.Port),
				"new": fmt.Sprintf("%s:%d", new.Server.Host, new.Server.Port),
			},
		}
	}

	// Check database config (hide password)
	if old.Database.Host != new.Database.Host || old.Database.Port != new.Database.Port {
		changes["database"] = map[string]interface{}{
			"host": map[string]interface{}{
				"old": fmt.Sprintf("%s:%d", old.Database.Host, old.Database.Port),
				"new": fmt.Sprintf("%s:%d", new.Database.Host, new.Database.Port),
			},
		}
	}

	// Check feature flags
	if old.Features != new.Features {
		changes["features"] = map[string]interface{}{
			"old": old.Features,
			"new": new.Features,
		}
	}

	// Check logging config
	if old.Logging.Level != new.Logging.Level {
		changes["logging.level"] = map[string]interface{}{
			"old": old.Logging.Level,
			"new": new.Logging.Level,
		}
	}

	return changes
}

// Manager manages configuration with hot reload support
type Manager struct {
	watcher   *Watcher
	mu        sync.RWMutex
	listeners map[string][]func(*Config)
}

// NewManager creates a new configuration manager
func NewManager(configPath string, logger *zap.Logger) (*Manager, error) {
	watcher, err := NewWatcher(configPath, logger)
	if err != nil {
		return nil, err
	}

	return &Manager{
		watcher:   watcher,
		listeners: make(map[string][]func(*Config)),
	}, nil
}

// Get returns the current configuration
func (m *Manager) Get() *Config {
	return m.watcher.Get()
}

// Subscribe subscribes to configuration changes for a specific component
func (m *Manager) Subscribe(component string, callback func(*Config)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.listeners[component] = append(m.listeners[component], callback)

	// Register with watcher
	m.watcher.OnChange(func(config *Config) {
		m.mu.RLock()
		listeners := m.listeners[component]
		m.mu.RUnlock()

		for _, listener := range listeners {
			listener(config)
		}
	})
}

// UpdateLogLevel dynamically updates the log level
func (m *Manager) UpdateLogLevel(logger *zap.AtomicLevel) {
	m.Subscribe("logging", func(config *Config) {
		level, err := zap.ParseAtomicLevel(config.Logging.Level)
		if err != nil {
			log.Printf("Invalid log level: %s", config.Logging.Level)
			return
		}
		logger.SetLevel(level.Level())
		log.Printf("Log level updated to: %s", config.Logging.Level)
	})
}

// UpdateFeatureFlags dynamically updates feature flags
func (m *Manager) UpdateFeatureFlags(handler func(FeatureFlags)) {
	m.Subscribe("features", func(config *Config) {
		handler(config.Features)
	})
}

// Close stops the configuration manager
func (m *Manager) Close() error {
	return m.watcher.Close()
}
