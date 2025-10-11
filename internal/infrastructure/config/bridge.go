package config

import (
	"fmt"
	"strconv"
	"time"

	pkgconfig "library-service/pkg/config"
)

// Bridge provides backwards compatibility with the old config interface
// while using the new pkg/config system underneath

// Load loads configuration using the new system and converts to old format
func Load() (*Config, error) {
	// Load configuration using new system
	newConfig := pkgconfig.MustLoad("")

	// Convert to old format
	oldConfig := &Config{
		App: AppConfig{
			Mode:    newConfig.App.Environment,
			Port:    ":" + strconv.Itoa(newConfig.Server.Port),
			Host:    fmt.Sprintf("http://%s:%d", newConfig.Server.Host, newConfig.Server.Port),
			Path:    "/",
			Timeout: newConfig.Server.ReadTimeout,
		},
		JWT: JWTConfig{
			Secret:          newConfig.JWT.Secret,
			AccessTokenTTL:  newConfig.JWT.AccessTokenTTL,
			RefreshTokenTTL: newConfig.JWT.RefreshTokenTTL,
			Issuer:          newConfig.JWT.Issuer,
		},
	}

	return oldConfig, nil
}

// LoadFromNew converts new config to old format
func LoadFromNew(newConfig *pkgconfig.Config) *Config {
	return &Config{
		App: AppConfig{
			Mode:    newConfig.App.Environment,
			Port:    ":" + strconv.Itoa(newConfig.Server.Port),
			Host:    fmt.Sprintf("http://%s:%d", newConfig.Server.Host, newConfig.Server.Port),
			Path:    "/",
			Timeout: newConfig.Server.ReadTimeout,
		},
		JWT: JWTConfig{
			Secret:          newConfig.JWT.Secret,
			AccessTokenTTL:  newConfig.JWT.AccessTokenTTL,
			RefreshTokenTTL: newConfig.JWT.RefreshTokenTTL,
			Issuer:          newConfig.JWT.Issuer,
		},
	}
}

// ToNew converts old config to new format
func ToNew(oldConfig *Config) *pkgconfig.Config {
	// Parse port from string
	port := 8080
	if oldConfig.App.Port != "" && len(oldConfig.App.Port) > 1 {
		if p, err := strconv.Atoi(oldConfig.App.Port[1:]); err == nil {
			port = p
		}
	}

	// Map old mode to new environment
	environment := oldConfig.App.Mode
	if environment == "dev" {
		environment = "development"
	} else if environment == "prod" {
		environment = "production"
	}

	return &pkgconfig.Config{
		App: pkgconfig.AppConfig{
			Name:        "Library Service",
			Version:     "1.0.0",
			Environment: environment,
			Debug:       environment == "development",
		},
		Server: pkgconfig.ServerConfig{
			Host:            "0.0.0.0",
			Port:            port,
			ReadTimeout:     oldConfig.App.Timeout,
			WriteTimeout:    oldConfig.App.Timeout,
			IdleTimeout:     60 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			MaxRequestSize:  10485760,
			EnableCORS:      true,
			AllowedOrigins:  []string{"*"},
			EnableRateLimit: true,
			RateLimit:       100,
		},
		Database: pkgconfig.DatabaseConfig{
			Driver:          "postgres",
			Host:            "localhost",
			Port:            5432,
			Database:        "library",
			Username:        "library",
			Password:        "library123",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    25,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 5 * time.Minute,
			EnableMigration: true,
			MigrationPath:   "migrations/postgres",
		},
		Redis: pkgconfig.RedisConfig{
			Enabled:      false,
			Host:         "localhost",
			Port:         6379,
			Database:     0,
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolSize:     10,
			TTL:          time.Hour,
		},
		JWT: pkgconfig.JWTConfig{
			Secret:          oldConfig.JWT.Secret,
			Issuer:          oldConfig.JWT.Issuer,
			AccessTokenTTL:  oldConfig.JWT.AccessTokenTTL,
			RefreshTokenTTL: oldConfig.JWT.RefreshTokenTTL,
			Algorithm:       "HS256",
		},
		Payment: pkgconfig.PaymentConfig{
			Provider:      "mock",
			Currency:      "USD",
			Timeout:       30 * time.Second,
			RetryAttempts: 3,
			RetryDelay:    time.Second,
			Sandbox:       true,
		},
		Logging: pkgconfig.LoggingConfig{
			Level:              "debug",
			Format:             "console",
			Output:             "stdout",
			EnableCaller:       true,
			EnableStacktrace:   false,
			SlowQueryThreshold: 100 * time.Millisecond,
		},
		Metrics: pkgconfig.MetricsConfig{
			Enabled:         true,
			Path:            "/metrics",
			Port:            9090,
			Namespace:       "library",
			CollectInterval: 10 * time.Second,
			Histograms:      true,
		},
		Features: pkgconfig.FeatureFlags{
			EnableSwagger:       true,
			EnableGraphQL:       false,
			EnableWebSocket:     false,
			EnableNotifications: false,
			EnableSearch:        true,
			EnableReservations:  true,
			EnablePayments:      true,
			MaintenanceMode:     false,
		},
	}
}
