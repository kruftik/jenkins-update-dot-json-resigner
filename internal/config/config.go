package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
)

type AppConfig struct {
	Dbg bool `long:"debug" env:"DEBUG" description:"debug mode"`

	UpdateJSONPath string `long:"update-json-path"  env:"UPDATE_JSON_PATH"`
	UpdateJSONURL  string `long:"update-json-url" env:"UPDATE_JSON_URL"`

	UpdateJSONDownloadTimeout time.Duration `long:"timeout" env:"UPDATE_JSON_DOWNLOAD_TIMEOUT" default:"120s"`

	UpdateJSONCacheTTL time.Duration `long:"cache-ttl" env:"UPDATE_JSON_CACHE_TTL" default:"30m"`

	OriginDownloadURL string `long:"origin-download-uri" env:"ORIGIN_DOWNLOAD_URL" default:"https://updates.jenkins.io/"`
	NewDownloadURL    string `long:"new-download-uri" env:"NEW_DOWNLOAD_URL" required:"true"`

	SignCAPath          string `long:"ca-certificate-path" env:"SIGN_CA_PATH" description:"x509 CA certificates path"`
	SignCertificatePath string `long:"certificate-path" env:"SIGN_CERTIFICATE_PATH" description:"x509-certificate path" required:"true"`
	SignKeyPath         string `long:"key-path" env:"SIGN_KEY_PATH" description:"private key path" required:"true"`
	SignKeyPassword     string `long:"private-key-pass" env:"SIGN_KEY_PASSWORD"`

	ServerAddr string `long:"server-addr" env:"SERVER_ADDR" default:""`
	ServerPort int    `long:"listen-port" env:"LISTEN_PORT" default:"8282"`

	DataDirPath string `long:"data-dir" env:"DATA_DIR" default:"/tmp/update-center-data"`
}

func (cfg AppConfig) validateSource() error {
	if cfg.UpdateJSONURL == "" && cfg.UpdateJSONPath == "" {
		return fmt.Errorf("either update.json URL or path must be configured")
	}

	if cfg.UpdateJSONURL != "" && cfg.UpdateJSONPath != "" {
		return fmt.Errorf("update.json URL and path cannot be used simultaneously")
	}

	return nil
}

func ParseConfig() (AppConfig, error) {
	cfg := AppConfig{}

	if _, err := flags.Parse(&cfg); err != nil {
		return AppConfig{}, err
	}

	cfg.OriginDownloadURL = strings.TrimSuffix(cfg.OriginDownloadURL, "/")
	cfg.NewDownloadURL = strings.TrimSuffix(cfg.NewDownloadURL, "/")

	if err := cfg.validateSource(); err != nil {
		return AppConfig{}, fmt.Errorf("invalid source: %w", err)
	}

	if err := os.MkdirAll(cfg.DataDirPath, 0755); err != nil {
		return AppConfig{}, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return cfg, nil
}
