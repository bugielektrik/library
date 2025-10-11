package config

import (
	"fmt"
	"sync"
	"time"
)

// Global configuration instance
var (
	global     *Config
	globalOnce sync.Once
	globalMu   sync.RWMutex
)

// Init initializes the global configuration
func Init(configPath string) error {
	loader := NewLoader()
	config, err := loader.Load(configPath)
	if err != nil {
		return err
	}

	SetGlobal(config)
	return nil
}

// SetGlobal sets the global configuration instance
func SetGlobal(config *Config) {
	globalMu.Lock()
	defer globalMu.Unlock()
	global = config
}

// Get returns the global configuration instance
func Get() *Config {
	globalMu.RLock()
	defer globalMu.RUnlock()

	if global == nil {
		// Return default configuration if not initialized
		globalOnce.Do(func() {
			global = LoadWithDefaults()
		})
	}

	return global
}

// Helper functions for common configuration values

// GetDatabaseDSN returns the database connection string
func GetDatabaseDSN() string {
	return Get().Database.GetDSN()
}

// GetServerAddress returns the server address
func GetServerAddress() string {
	config := Get()
	return fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
}

// GetRedisAddress returns the Redis address
func GetRedisAddress() string {
	config := Get()
	return fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port)
}

// IsProduction returns true if running in production
func IsProduction() bool {
	return Get().IsProduction()
}

// IsDevelopment returns true if running in development
func IsDevelopment() bool {
	return Get().IsDevelopment()
}

// IsDebug returns true if debug mode is enabled
func IsDebug() bool {
	return Get().IsDebug()
}

// GetLogLevel returns the configured log level
func GetLogLevel() string {
	return Get().Logging.Level
}

// GetJWTSecret returns the JWT secret
func GetJWTSecret() string {
	return Get().JWT.Secret
}

// GetJWTExpiry returns the JWT access token expiry
func GetJWTExpiry() time.Duration {
	return Get().JWT.AccessTokenTTL
}

// IsFeatureEnabled checks if a feature flag is enabled
func IsFeatureEnabled(feature string) bool {
	features := Get().Features
	switch feature {
	case "swagger":
		return features.EnableSwagger
	case "graphql":
		return features.EnableGraphQL
	case "websocket":
		return features.EnableWebSocket
	case "notifications":
		return features.EnableNotifications
	case "search":
		return features.EnableSearch
	case "reservations":
		return features.EnableReservations
	case "payments":
		return features.EnablePayments
	default:
		return false
	}
}

// IsMaintenanceMode returns true if maintenance mode is enabled
func IsMaintenanceMode() bool {
	return Get().Features.MaintenanceMode
}

// GetRateLimit returns the rate limit per minute
func GetRateLimit() int {
	return Get().Server.RateLimit
}

// IsRateLimitEnabled returns true if rate limiting is enabled
func IsRateLimitEnabled() bool {
	return Get().Server.EnableRateLimit
}

// IsCORSEnabled returns true if CORS is enabled
func IsCORSEnabled() bool {
	return Get().Server.EnableCORS
}

// GetAllowedOrigins returns the allowed CORS origins
func GetAllowedOrigins() []string {
	return Get().Server.AllowedOrigins
}

// GetMaxRequestSize returns the maximum request size in bytes
func GetMaxRequestSize() int64 {
	return Get().Server.MaxRequestSize
}

// GetMetricsPort returns the metrics server port
func GetMetricsPort() int {
	return Get().Metrics.Port
}

// IsMetricsEnabled returns true if metrics are enabled
func IsMetricsEnabled() bool {
	return Get().Metrics.Enabled
}

// GetPaymentProvider returns the payment provider
func GetPaymentProvider() string {
	return Get().Payment.Provider
}

// IsPaymentSandbox returns true if payment is in sandbox mode
func IsPaymentSandbox() bool {
	return Get().Payment.Sandbox
}

// GetCacheTTL returns the cache TTL duration
func GetCacheTTL() time.Duration {
	return Get().Redis.TTL
}

// IsCacheEnabled returns true if Redis cache is enabled
func IsCacheEnabled() bool {
	return Get().Redis.Enabled
}

// ConfigOption represents a configuration option for testing
type ConfigOption func(*Config)

// WithTestConfig creates a test configuration with options
func WithTestConfig(opts ...ConfigOption) *Config {
	config := LoadWithDefaults()

	// Apply test defaults
	config.App.Environment = "test"
	config.Database.Database = "library_test"
	config.JWT.Secret = "test-secret-key-for-testing-purposes-only-32ch"
	config.Logging.Level = "error"
	config.Redis.Enabled = false
	config.Metrics.Enabled = false

	// Apply custom options
	for _, opt := range opts {
		opt(config)
	}

	return config
}

// WithDatabase sets database configuration
func WithDatabase(host string, port int, name string) ConfigOption {
	return func(c *Config) {
		c.Database.Host = host
		c.Database.Port = port
		c.Database.Database = name
	}
}

// WithJWT sets JWT configuration
func WithJWT(secret string, expiry time.Duration) ConfigOption {
	return func(c *Config) {
		c.JWT.Secret = secret
		c.JWT.AccessTokenTTL = expiry
	}
}

// WithRedis sets Redis configuration
func WithRedis(enabled bool, host string, port int) ConfigOption {
	return func(c *Config) {
		c.Redis.Enabled = enabled
		c.Redis.Host = host
		c.Redis.Port = port
	}
}

// WithFeatures sets feature flags
func WithFeatures(features FeatureFlags) ConfigOption {
	return func(c *Config) {
		c.Features = features
	}
}
