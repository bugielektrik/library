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

type Configs struct {
	APP        AppConfig
	Store      StoreConfig
	ClickHouse ClickHouseConfig
	JWT        JWTConfig
	NATS       NATSConfig
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

type ClickHouseConfig struct {
	DSN string
}

type JWTConfig struct {
	AccessSecret     string        `required:"true"`
	RefreshSecret    string        `required:"true"`
	AccessTokenTTL   time.Duration `default:"15m"`
	RefreshTokenTTL  time.Duration `default:"168h"`
}

type NATSConfig struct {
	URL           string `required:"true"`
	Subject       string `default:"library.service"`
	StreamName    string `default:"LIBRARY_EVENTS"`
	EnableJetStream bool   `default:"false"`
}

func New() (*Configs, error) {
	cfg := &Configs{}

	root, err := os.Getwd()
	if err != nil {
		logStructured("error", "get_workdir", map[string]interface{}{"error": err.Error()})
		return cfg, fmt.Errorf("unable to get working directory: %w", err)
	}

	envPath := filepath.Join(root, ".env")
	if _, statErr := os.Stat(envPath); statErr == nil {
		if loadErr := godotenv.Load(envPath); loadErr != nil {
			logStructured("error", "load_env", map[string]interface{}{"file": envPath, "error": loadErr.Error()})
			return cfg, fmt.Errorf("failed to load env file %s: %w", envPath, loadErr)
		}
		logStructured("info", "load_env", map[string]interface{}{"file": envPath})
	} else if !os.IsNotExist(statErr) {
		logStructured("error", "stat_env_file", map[string]interface{}{"file": envPath, "error": statErr.Error()})
		return cfg, fmt.Errorf("failed to stat env file %s: %w", envPath, statErr)
	} else {
		logStructured("info", "env_file_missing", map[string]interface{}{"file": envPath})
	}

	cfg.APP = AppConfig{
		Mode:    defaultAppMode,
		Port:    defaultAppPort,
		Host:    defaultAppHost,
		Path:    defaultAppPath,
		Timeout: defaultAppTimeout,
	}

	targets := map[string]interface{}{
		"APP":        &cfg.APP,
		"POSTGRES":   &cfg.Store,
		"CLICKHOUSE": &cfg.ClickHouse,
		"JWT":        &cfg.JWT,
		"NATS":       &cfg.NATS,
	}

	for p, target := range targets {
		if target == nil {
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

func logStructured(level string, action string, params map[string]interface{}) {
	msg := fmt.Sprintf("level=%s component=config action=%s", level, action)
	for k, v := range params {
		msg = fmt.Sprintf("%s %s=%v", msg, k, v)
	}
	log.Println(msg)
}
