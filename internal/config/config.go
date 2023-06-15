package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

const (
	defaultHTTPPort               = "80"
	defaultHTTPHost               = "localhost"
	defaultHTTPSchema             = "http"
	defaultHTTPReadTimeout        = 15 * time.Second
	defaultHTTPWriteTimeout       = 15 * time.Second
	defaultHTTPIdleTimeout        = 60 * time.Second
	defaultHTTPMaxHeaderMegabytes = 1

	defaultOAUTHSecret  = "IP03O5Ekg91g5jw=="
	defaultOAUTHExpires = 1200 * time.Second
)

type (
	Configs struct {
		HTTP     HTTPConfig
		OAUTH    OAuthConfig
		POSTGRES DatabaseConfig
	}

	HTTPConfig struct {
		Port               string
		Host               string
		Schema             string
		ReadTimeout        time.Duration
		WriteTimeout       time.Duration
		IdleTimeout        time.Duration
		MaxHeaderMegabytes int
	}

	ClientConfig struct {
		Endpoint string
		Username string
		Password string
	}

	OAuthConfig struct {
		Secret   string
		Expires  time.Duration
		Duration string
	}

	DatabaseConfig struct {
		DSN string
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

	cfg.HTTP = HTTPConfig{
		Port:               defaultHTTPPort,
		Host:               defaultHTTPHost,
		Schema:             defaultHTTPSchema,
		ReadTimeout:        defaultHTTPReadTimeout,
		WriteTimeout:       defaultHTTPWriteTimeout,
		IdleTimeout:        defaultHTTPIdleTimeout,
		MaxHeaderMegabytes: defaultHTTPMaxHeaderMegabytes,
	}

	err = envconfig.Process("HTTP", &cfg.HTTP)
	if err != nil {
		return
	}

	cfg.OAUTH = OAuthConfig{
		Secret:  defaultOAUTHSecret,
		Expires: defaultOAUTHExpires,
	}

	err = envconfig.Process("OAUTH", &cfg.OAUTH)
	if err != nil {
		return
	}

	duration, err := time.ParseDuration(cfg.OAUTH.Duration)
	if err == nil {
		return
	}
	cfg.OAUTH.Expires = duration

	err = envconfig.Process("POSTGRES", &cfg.POSTGRES)
	if err != nil {
		return
	}

	return
}
