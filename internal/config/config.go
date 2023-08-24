package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

const (
	defaultPort    = "8080"
	defaultHost    = "http://localhost:8080"
	defaultTimeout = 60 * time.Second

	defaultKey     = "IP03O5Ekg91g5jw=="
	defaultExpires = 3600 * time.Second
)

type (
	Configs struct {
		APP      AppConfig
		POSTGRES StoreConfig
	}

	AppConfig struct {
		ServerPort    string        `split_words:"true" required:"true"`
		ServerHost    string        `split_words:"true" required:"true"`
		ServerTimeout time.Duration `split_words:"true" required:"true"`
		TokenKey      string        `split_words:"true"`
		TokenExpires  time.Duration `split_words:"true"`
	}

	ClientConfig struct {
		URL      string `split_words:"true" required:"true"`
		Login    string `split_words:"true"`
		Password string `split_words:"true"`
	}

	StoreConfig struct {
		DSN string `split_words:"true" required:"true"`
	}
)

// New populates Configs struct with values from config file
// located at filepath and environment variables.
func New() (cfg Configs, err error) {
	root, err := os.Getwd()
	if err != nil {
		return
	}
	godotenv.Load(filepath.Join(root, ".env"))

	cfg.APP = AppConfig{
		ServerPort:    defaultPort,
		ServerHost:    defaultHost,
		ServerTimeout: defaultTimeout,

		TokenKey:     defaultKey,
		TokenExpires: defaultExpires,
	}

	if err = envconfig.Process("APP", &cfg.APP); err != nil {
		return
	}

	if err = envconfig.Process("POSTGRES", &cfg.POSTGRES); err != nil {
		return
	}

	return
}
