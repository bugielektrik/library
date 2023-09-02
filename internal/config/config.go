package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

const (
	defaultAppMode    = "dev"
	defaultAppPort    = "8080"
	defaultAppPath    = "/api/v1"
	defaultAppTimeout = 60 * time.Second

	defaultTokenSalt    = "IP03O5Ekg91g5jw=="
	defaultTokenExpires = 3600 * time.Second
)

type (
	Configs struct {
		APP      AppConfig
		TOKEN    TokenConfig
		MONGO    StoreConfig
		POSTGRES StoreConfig
	}

	AppConfig struct {
		Mode    string `required:"true"`
		Port    string
		Path    string
		Timeout time.Duration
	}

	TokenConfig struct {
		Salt    string
		Expires time.Duration
	}

	ClientConfig struct {
		URL      string `required:"true"`
		Login    string
		Password string
	}

	StoreConfig struct {
		DSN string `required:"true"`
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
		Mode:    defaultAppMode,
		Port:    defaultAppPort,
		Path:    defaultAppPath,
		Timeout: defaultAppTimeout,
	}

	cfg.TOKEN = TokenConfig{
		Salt:    defaultTokenSalt,
		Expires: defaultTokenExpires,
	}

	if err = envconfig.Process("APP", &cfg.APP); err != nil {
		return
	}

	if err = envconfig.Process("MONGO", &cfg.MONGO); err != nil {
		return
	}

	if err = envconfig.Process("POSTGRES", &cfg.POSTGRES); err != nil {
		return
	}

	return
}
