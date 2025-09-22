package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

const (
	defaultAppMode    = "dev"
	defaultAppPort    = ":80"
	defaultAppHost    = "http://localhost:80"
	defaultAppPath    = "/"
	defaultAppTimeout = 60 * time.Second
)

// Configs holds all application configuration groups loaded from environment.
type Configs struct {
	APP      AppConfig
	EPAY     ClientConfig
	POSTGRES StoreConfig
}

type AppConfig struct {
	Mode    string `required:"true"`
	Port    string
	Host    string
	Path    string
	Timeout time.Duration
}

type ClientConfig struct {
	URL      string
	Login    string
	Password string
	OAuth    string
	JS       string
}

type StoreConfig struct {
	DSN string
}

// New loads configuration from a .env file (if present) and environment variables.
// It returns the populated Configs or an error if required processing fails.
//
// Behavior:
// - If a .env file is not present, it continues without error.
// - Any failure to parse environment variables for a specific component is returned with context.
// - Structured logging is emitted via the standard logger in key=value form.
func New() (*Configs, error) {
	cfg := &Configs{}

	// Load current working directory to locate .env file.
	root, err := os.Getwd()
	if err != nil {
		logStructured("error", "get_workdir", map[string]interface{}{"error": err.Error()})
		return cfg, fmt.Errorf("unable to get working directory: %w", err)
	}

	envPath := filepath.Join(root, ".env")
	// If .env exists, attempt to load it. Missing file is not considered an error.
	if _, statErr := os.Stat(envPath); statErr == nil {
		if loadErr := godotenv.Load(envPath); loadErr != nil {
			logStructured("error", "load_env", map[string]interface{}{"file": envPath, "error": loadErr.Error()})
			return cfg, fmt.Errorf("failed to load env file %s: %w", envPath, loadErr)
		}
		logStructured("info", "load_env", map[string]interface{}{"file": envPath})
	} else if !os.IsNotExist(statErr) {
		// Any stat error other than not-exist is unexpected.
		logStructured("error", "stat_env_file", map[string]interface{}{"file": envPath, "error": statErr.Error()})
		return cfg, fmt.Errorf("failed to stat env file %s: %w", envPath, statErr)
	} else {
		// .env not present; continue with environment variables only.
		logStructured("info", "env_file_missing", map[string]interface{}{"file": envPath})
	}

	// Set sane defaults for the application config before overriding with env vars.
	cfg.APP = AppConfig{
		Mode:    defaultAppMode,
		Port:    defaultAppPort,
		Host:    defaultAppHost,
		Path:    defaultAppPath,
		Timeout: defaultAppTimeout,
	}

	// Map prefixes to the corresponding destination struct pointers.
	targets := map[string]interface{}{
		"APP":      &cfg.APP,
		"EPAY":     &cfg.EPAY,
		"POSTGRES": &cfg.POSTGRES,
	}

	// Process each prefix; return early on error with context.
	for p, target := range targets {
		if target == nil {
			// Defensive: should not happen, but helps catch typos during maintenance.
			logStructured("error", "missing_target", map[string]interface{}{"prefix": p})
			return cfg, fmt.Errorf("internal error: missing target for prefix %q", p)
		}

		if procErr := envconfig.Process(p, target); procErr != nil {
			logStructured("error", "env_process", map[string]interface{}{"prefix": p, "error": procErr.Error()})
			return cfg, fmt.Errorf("failed to process env for %s: %w", p, procErr)
		}
	}

	return cfg, nil
}

// logStructured emits simple key=value structured logs (lowercase keys).
// This is a minimal structured logger using the standard library. Replace with a proper
// logger (e.g., zerolog, zap) if available in the project.
func logStructured(level string, action string, params map[string]interface{}) {
	// Build message: level=<level> action=<action> key1=value1 key2=value2 ...
	msg := fmt.Sprintf("level=%s component=config action=%s", level, action)
	for k, v := range params {
		msg = fmt.Sprintf("%s %s=%v", msg, k, v)
	}
	// Use the standard log package to emit the message.
	log.Println(msg)
}
