package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Loader handles configuration loading from various sources using Viper
type Loader struct {
	viper       *viper.Viper
	config      *Config
	configPath  string
	environment string
}

// NewLoader creates a new configuration loader with Viper
func NewLoader() *Loader {
	v := viper.New()

	// Set environment variable prefix and enable automatic env binding
	v.SetEnvPrefix("") // No prefix, match all env vars
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return &Loader{
		viper:       v,
		config:      &Config{},
		environment: getEnvOrDefault("APP_ENV", "development"),
	}
}

// Load loads configuration from all sources with priority:
// 1. Environment variables (highest)
// 2. Environment-specific config file (config.production.yaml)
// 3. Base config file (config.yaml)
// 4. Default values (lowest)
func (l *Loader) Load(configPath string) (*Config, error) {
	l.configPath = configPath

	// 1. Set default values
	l.setDefaults()

	// 2. Load from config file if exists
	if configPath != "" {
		if err := l.loadFromFile(configPath); err != nil {
			// Config file is optional
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("loading config file: %w", err)
			}
		}
	}

	// 3. Load environment-specific config (e.g., config.production.yaml)
	if err := l.loadEnvironmentConfig(); err != nil {
		// Environment-specific config is optional
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("loading environment config: %w", err)
		}
	}

	// 4. Bind environment variables (Viper does this automatically with AutomaticEnv)
	l.bindEnvironmentVariables()

	// 5. Unmarshal into our config struct
	if err := l.viper.Unmarshal(l.config); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	// 6. Validate final configuration
	if err := l.config.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return l.config, nil
}

// loadFromFile loads configuration from a specific file
func (l *Loader) loadFromFile(path string) error {
	l.viper.SetConfigFile(path)

	if err := l.viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

// loadEnvironmentConfig loads environment-specific configuration
func (l *Loader) loadEnvironmentConfig() error {
	if l.configPath == "" {
		return nil
	}

	dir := filepath.Dir(l.configPath)
	base := filepath.Base(l.configPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// Try to load environment-specific config (e.g., config.production.yaml)
	envPath := filepath.Join(dir, fmt.Sprintf("%s.%s%s", name, l.environment, ext))

	// Check if file exists
	if _, err := os.Stat(envPath); err != nil {
		return err
	}

	// Merge environment-specific config
	l.viper.SetConfigFile(envPath)
	return l.viper.MergeInConfig()
}

// bindEnvironmentVariables explicitly binds environment variables to config keys
func (l *Loader) bindEnvironmentVariables() {
	// App config
	l.viper.BindEnv("app.env", "APP_ENV")
	l.viper.BindEnv("app.debug", "DEBUG")

	// Server config
	l.viper.BindEnv("server.host", "SERVER_HOST")
	l.viper.BindEnv("server.port", "PORT", "SERVER_PORT")

	// Database config
	l.viper.BindEnv("database.host", "DB_HOST")
	l.viper.BindEnv("database.port", "DB_PORT")
	l.viper.BindEnv("database.database", "DB_NAME")
	l.viper.BindEnv("database.username", "DB_USER")
	l.viper.BindEnv("database.password", "DB_PASSWORD")

	// Redis config
	l.viper.BindEnv("redis.enabled", "REDIS_ENABLED")
	l.viper.BindEnv("redis.host", "REDIS_HOST")
	l.viper.BindEnv("redis.port", "REDIS_PORT")
	l.viper.BindEnv("redis.password", "REDIS_PASSWORD")

	// JWT config
	l.viper.BindEnv("jwt.secret", "JWT_SECRET")
	l.viper.BindEnv("jwt.access_token_ttl", "JWT_ACCESS_TTL")
	l.viper.BindEnv("jwt.refresh_token_ttl", "JWT_REFRESH_TTL")
	l.viper.BindEnv("jwt.issuer", "JWT_ISSUER")

	// Payment config
	l.viper.BindEnv("payment.api_key", "PAYMENT_API_KEY")
	l.viper.BindEnv("payment.secret_key", "PAYMENT_SECRET_KEY")
	l.viper.BindEnv("payment.webhook_secret", "PAYMENT_WEBHOOK_SECRET")

	// Logging config
	l.viper.BindEnv("logging.level", "LOG_LEVEL")
}

// setDefaults sets default values for all configuration fields
func (l *Loader) setDefaults() {
	// App defaults
	l.viper.SetDefault("app.name", "Library Service")
	l.viper.SetDefault("app.version", "1.0.0")
	l.viper.SetDefault("app.env", "development")
	l.viper.SetDefault("app.debug", false)

	// Server defaults
	l.viper.SetDefault("server.host", "0.0.0.0")
	l.viper.SetDefault("server.port", 8080)
	l.viper.SetDefault("server.read_timeout", "30s")
	l.viper.SetDefault("server.write_timeout", "30s")
	l.viper.SetDefault("server.idle_timeout", "60s")
	l.viper.SetDefault("server.shutdown_timeout", "10s")
	l.viper.SetDefault("server.max_request_size", 10485760) // 10MB
	l.viper.SetDefault("server.enable_cors", true)
	l.viper.SetDefault("server.allowed_origins", []string{"*"})
	l.viper.SetDefault("server.enable_rate_limit", true)
	l.viper.SetDefault("server.rate_limit", 100)

	// Database defaults
	l.viper.SetDefault("database.driver", "postgres")
	l.viper.SetDefault("database.host", "localhost")
	l.viper.SetDefault("database.port", 5432)
	l.viper.SetDefault("database.database", "library")
	l.viper.SetDefault("database.username", "library")
	l.viper.SetDefault("database.password", "library123")
	l.viper.SetDefault("database.ssl_mode", "disable")
	l.viper.SetDefault("database.max_open_conns", 25)
	l.viper.SetDefault("database.max_idle_conns", 25)
	l.viper.SetDefault("database.conn_max_lifetime", "5m")
	l.viper.SetDefault("database.conn_max_idle_time", "5m")
	l.viper.SetDefault("database.enable_migration", true)
	l.viper.SetDefault("database.migration_path", "migrations/postgres")

	// Redis defaults
	l.viper.SetDefault("redis.enabled", false)
	l.viper.SetDefault("redis.host", "localhost")
	l.viper.SetDefault("redis.port", 6379)
	l.viper.SetDefault("redis.database", 0)
	l.viper.SetDefault("redis.max_retries", 3)
	l.viper.SetDefault("redis.dial_timeout", "5s")
	l.viper.SetDefault("redis.read_timeout", "3s")
	l.viper.SetDefault("redis.write_timeout", "3s")
	l.viper.SetDefault("redis.pool_size", 10)
	l.viper.SetDefault("redis.ttl", "1h")

	// JWT defaults
	l.viper.SetDefault("jwt.issuer", "library-service")
	l.viper.SetDefault("jwt.access_token_ttl", "24h")
	l.viper.SetDefault("jwt.refresh_token_ttl", "168h") // 7 days
	l.viper.SetDefault("jwt.algorithm", "HS256")

	// Payment defaults
	l.viper.SetDefault("payment.provider", "mock")
	l.viper.SetDefault("payment.currency", "USD")
	l.viper.SetDefault("payment.timeout", "30s")
	l.viper.SetDefault("payment.retry_attempts", 3)
	l.viper.SetDefault("payment.retry_delay", "1s")
	l.viper.SetDefault("payment.sandbox", true)

	// Logging defaults
	l.viper.SetDefault("logging.level", "info")
	l.viper.SetDefault("logging.format", "json")
	l.viper.SetDefault("logging.output", "stdout")
	l.viper.SetDefault("logging.file_path", "logs/app.log")
	l.viper.SetDefault("logging.max_size", 100)
	l.viper.SetDefault("logging.max_backups", 5)
	l.viper.SetDefault("logging.max_age", 30)
	l.viper.SetDefault("logging.compress", false)
	l.viper.SetDefault("logging.enable_caller", true)
	l.viper.SetDefault("logging.enable_stacktrace", false)
	l.viper.SetDefault("logging.skip_paths", []string{"/health", "/metrics"})
	l.viper.SetDefault("logging.slow_query_threshold", "100ms")

	// Metrics defaults
	l.viper.SetDefault("metrics.enabled", true)
	l.viper.SetDefault("metrics.path", "/metrics")
	l.viper.SetDefault("metrics.port", 9090)
	l.viper.SetDefault("metrics.namespace", "library")
	l.viper.SetDefault("metrics.collect_interval", "10s")
	l.viper.SetDefault("metrics.histograms", true)

	// Feature flags defaults
	l.viper.SetDefault("features.enable_swagger", true)
	l.viper.SetDefault("features.enable_graphql", false)
	l.viper.SetDefault("features.enable_websocket", false)
	l.viper.SetDefault("features.enable_notifications", false)
	l.viper.SetDefault("features.enable_search", true)
	l.viper.SetDefault("features.enable_reservations", true)
	l.viper.SetDefault("features.enable_payments", true)
	l.viper.SetDefault("features.maintenance_mode", false)
}

// getEnvOrDefault returns environment variable or default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// MustLoad loads configuration and panics on error
func MustLoad(configPath string) *Config {
	loader := NewLoader()
	config, err := loader.Load(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return config
}

// LoadWithDefaults loads configuration with defaults only
func LoadWithDefaults() *Config {
	loader := NewLoader()
	config, _ := loader.Load("")
	return config
}
