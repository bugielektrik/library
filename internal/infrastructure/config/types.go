package config

import (
	"fmt"
	"time"
)

// Config represents the complete application configuration
type Config struct {
	App      AppConfig      `yaml:"app" json:"app" validate:"required"`
	Server   ServerConfig   `yaml:"server" json:"server" validate:"required"`
	Database DatabaseConfig `yaml:"database" json:"database" validate:"required"`
	Redis    RedisConfig    `yaml:"redis" json:"redis"`
	JWT      JWTConfig      `yaml:"jwt" json:"jwt" validate:"required"`
	Payment  PaymentConfig  `yaml:"payment" json:"payment"`
	Logging  LoggingConfig  `yaml:"logging" json:"logging"`
	Metrics  MetricsConfig  `yaml:"metrics" json:"metrics"`
	Features FeatureFlags   `yaml:"features" json:"features"`
}

// AppConfig contains application-level settings
type AppConfig struct {
	Name        string `yaml:"name" json:"name" default:"Library Service" validate:"required"`
	Version     string `yaml:"version" json:"version" default:"1.0.0"`
	Environment string `yaml:"env" json:"env" env:"APP_ENV" default:"development" validate:"required,oneof=development staging production"`
	Debug       bool   `yaml:"debug" json:"debug" env:"DEBUG" default:"false"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Host            string        `yaml:"host" json:"host" env:"SERVER_HOST" default:"0.0.0.0"`
	Port            int           `yaml:"port" json:"port" env:"PORT" default:"8080" validate:"min=1,max=65535"`
	ReadTimeout     time.Duration `yaml:"read_timeout" json:"read_timeout" default:"30s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" json:"write_timeout" default:"30s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" json:"idle_timeout" default:"60s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" json:"shutdown_timeout" default:"10s"`
	MaxRequestSize  int64         `yaml:"max_request_size" json:"max_request_size" default:"10485760"` // 10MB
	EnableCORS      bool          `yaml:"enable_cors" json:"enable_cors" default:"true"`
	AllowedOrigins  []string      `yaml:"allowed_origins" json:"allowed_origins" default:"[\"*\"]"`
	TrustedProxies  []string      `yaml:"trusted_proxies" json:"trusted_proxies"`
	EnableRateLimit bool          `yaml:"enable_rate_limit" json:"enable_rate_limit" default:"true"`
	RateLimit       int           `yaml:"rate_limit" json:"rate_limit" default:"100"` // requests per minute
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	Driver          string        `yaml:"driver" json:"driver" default:"postgres" validate:"required,oneof=postgres mysql sqlite"`
	Host            string        `yaml:"host" json:"host" env:"DB_HOST" default:"localhost" validate:"required"`
	Port            int           `yaml:"port" json:"port" env:"DB_PORT" default:"5432" validate:"min=1,max=65535"`
	Database        string        `yaml:"database" json:"database" env:"DB_NAME" default:"library" validate:"required"`
	Username        string        `yaml:"username" json:"username" env:"DB_USER" default:"library" validate:"required"`
	Password        string        `yaml:"password" json:"password" env:"DB_PASSWORD" secret:"true" validate:"required"`
	SSLMode         string        `yaml:"ssl_mode" json:"ssl_mode" default:"disable" validate:"oneof=disable require verify-ca verify-full"`
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns" default:"25"`
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns" default:"25"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime" default:"5m"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" json:"conn_max_idle_time" default:"5m"`
	EnableMigration bool          `yaml:"enable_migration" json:"enable_migration" default:"true"`
	MigrationPath   string        `yaml:"migration_path" json:"migration_path" default:"migrations/postgres"`
}

// GetDSN returns the database connection string
func (db DatabaseConfig) GetDSN() string {
	switch db.Driver {
	case "postgres":
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			db.Username, db.Password, db.Host, db.Port, db.Database, db.SSLMode)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			db.Username, db.Password, db.Host, db.Port, db.Database)
	default:
		return db.Database // SQLite
	}
}

// RedisConfig contains Redis cache settings
type RedisConfig struct {
	Enabled      bool          `yaml:"enabled" json:"enabled" env:"REDIS_ENABLED" default:"false"`
	Host         string        `yaml:"host" json:"host" env:"REDIS_HOST" default:"localhost"`
	Port         int           `yaml:"port" json:"port" env:"REDIS_PORT" default:"6379" validate:"min=1,max=65535"`
	Password     string        `yaml:"password" json:"password" env:"REDIS_PASSWORD" secret:"true"`
	Database     int           `yaml:"database" json:"database" default:"0" validate:"min=0,max=15"`
	MaxRetries   int           `yaml:"max_retries" json:"max_retries" default:"3"`
	DialTimeout  time.Duration `yaml:"dial_timeout" json:"dial_timeout" default:"5s"`
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout" default:"3s"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout" default:"3s"`
	PoolSize     int           `yaml:"pool_size" json:"pool_size" default:"10"`
	TTL          time.Duration `yaml:"ttl" json:"ttl" default:"1h"`
}

// JWTConfig contains JWT authentication settings
type JWTConfig struct {
	Secret          string        `yaml:"secret" json:"secret" env:"JWT_SECRET" secret:"true" validate:"required,min=32"`
	Issuer          string        `yaml:"issuer" json:"issuer" default:"library-service"`
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl" json:"access_token_ttl" default:"24h"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" json:"refresh_token_ttl" default:"168h"` // 7 days
	Algorithm       string        `yaml:"algorithm" json:"algorithm" default:"HS256" validate:"oneof=HS256 HS384 HS512"`
}

// PaymentConfig contains payment provider settings
type PaymentConfig struct {
	Provider      string            `yaml:"provider" json:"provider" default:"mock" validate:"oneof=mock stripe razorpay epayments"`
	APIKey        string            `yaml:"api_key" json:"api_key" env:"PAYMENT_API_KEY" secret:"true"`
	SecretKey     string            `yaml:"secret_key" json:"secret_key" env:"PAYMENT_SECRET_KEY" secret:"true"`
	WebhookSecret string            `yaml:"webhook_secret" json:"webhook_secret" env:"PAYMENT_WEBHOOK_SECRET" secret:"true"`
	Currency      string            `yaml:"currency" json:"currency" default:"USD"`
	Timeout       time.Duration     `yaml:"timeout" json:"timeout" default:"30s"`
	RetryAttempts int               `yaml:"retry_attempts" json:"retry_attempts" default:"3"`
	RetryDelay    time.Duration     `yaml:"retry_delay" json:"retry_delay" default:"1s"`
	Sandbox       bool              `yaml:"sandbox" json:"sandbox" default:"true"`
	CallbackURL   string            `yaml:"callback_url" json:"callback_url"`
	SuccessURL    string            `yaml:"success_url" json:"success_url"`
	FailureURL    string            `yaml:"failure_url" json:"failure_url"`
	Metadata      map[string]string `yaml:"metadata" json:"metadata"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level              string        `yaml:"level" json:"level" env:"LOG_LEVEL" default:"info" validate:"oneof=debug info warn error fatal"`
	Format             string        `yaml:"format" json:"format" default:"json" validate:"oneof=json console"`
	Output             string        `yaml:"output" json:"output" default:"stdout" validate:"oneof=stdout stderr file"`
	FilePath           string        `yaml:"file_path" json:"file_path" default:"logs/app.log"`
	MaxSize            int           `yaml:"max_size" json:"max_size" default:"100"` // megabytes
	MaxBackups         int           `yaml:"max_backups" json:"max_backups" default:"5"`
	MaxAge             int           `yaml:"max_age" json:"max_age" default:"30"` // days
	Compress           bool          `yaml:"compress" json:"compress" default:"false"`
	EnableCaller       bool          `yaml:"enable_caller" json:"enable_caller" default:"true"`
	EnableStacktrace   bool          `yaml:"enable_stacktrace" json:"enable_stacktrace" default:"false"`
	SkipPaths          []string      `yaml:"skip_paths" json:"skip_paths" default:"[\"/health\",\"/metrics\"]"`
	SlowQueryThreshold time.Duration `yaml:"slow_query_threshold" json:"slow_query_threshold" default:"100ms"`
}

// MetricsConfig contains metrics and monitoring settings
type MetricsConfig struct {
	Enabled         bool          `yaml:"enabled" json:"enabled" default:"true"`
	Path            string        `yaml:"path" json:"path" default:"/metrics"`
	Port            int           `yaml:"port" json:"port" default:"9090" validate:"min=1,max=65535"`
	Namespace       string        `yaml:"namespace" json:"namespace" default:"library"`
	CollectInterval time.Duration `yaml:"collect_interval" json:"collect_interval" default:"10s"`
	Histograms      bool          `yaml:"histograms" json:"histograms" default:"true"`
}

// FeatureFlags contains feature toggle settings
type FeatureFlags struct {
	EnableSwagger       bool `yaml:"enable_swagger" json:"enable_swagger" default:"true"`
	EnableGraphQL       bool `yaml:"enable_graphql" json:"enable_graphql" default:"false"`
	EnableWebSocket     bool `yaml:"enable_websocket" json:"enable_websocket" default:"false"`
	EnableNotifications bool `yaml:"enable_notifications" json:"enable_notifications" default:"false"`
	EnableSearch        bool `yaml:"enable_search" json:"enable_search" default:"true"`
	EnableReservations  bool `yaml:"enable_reservations" json:"enable_reservations" default:"true"`
	EnablePayments      bool `yaml:"enable_payments" json:"enable_payments" default:"true"`
	MaintenanceMode     bool `yaml:"maintenance_mode" json:"maintenance_mode" default:"false"`
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate required fields
	if c.App.Name == "" {
		return fmt.Errorf("app name is required")
	}

	if c.JWT.Secret == "" || len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters")
	}

	if c.Database.Host == "" || c.Database.Database == "" {
		return fmt.Errorf("database host and name are required")
	}

	// Validate port ranges
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}

	if c.Redis.Enabled && (c.Redis.Port < 1 || c.Redis.Port > 65535) {
		return fmt.Errorf("redis port must be between 1 and 65535")
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsDebug returns true if debug mode is enabled
func (c *Config) IsDebug() bool {
	return c.App.Debug || c.IsDevelopment()
}
