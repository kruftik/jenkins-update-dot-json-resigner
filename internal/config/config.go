package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
)

type ServerConfig struct {
	ListenAddr string `long:"listen-addr" env:"LISTEN_ADDR" default:""`
	ListenPort int    `long:"listen-port" env:"LISTEN_PORT" default:"8282"`

	TLSCertPath string `long:"tlscert" env:"TLS_CERT_PATH" default:""`
	TLSKeyPath  string `long:"tlskey" env:"TLS_KEY_PATH" default:""`
}

type SignerConfig struct {
	CAPath          string `long:"ca-certificate-path" env:"SIGN_CA_PATH" description:"x509 CA certificates path"`
	CertificatePath string `long:"certificate-path" env:"SIGN_CERTIFICATE_PATH" description:"x509-certificate path" required:"true"`
	KeyPath         string `long:"key-path" env:"SIGN_KEY_PATH" description:"private key path" required:"true"`
	KeyPassword     string `long:"private-key-pass" env:"SIGN_KEY_PASSWORD"`
}

type SourceConfig struct {
	Path string `long:"update-json-path"  env:"UPDATE_JSON_PATH"`
	URL  string `long:"update-json-url" env:"UPDATE_JSON_URL"`
}

type PatchConfig struct {
	OriginDownloadURL string `long:"origin-download-uri" env:"ORIGIN_DOWNLOAD_URL" default:"https://updates.jenkins.io/"`
	NewDownloadURL    string `long:"new-download-uri" env:"NEW_DOWNLOAD_URL" required:"true"`
}

type AppConfig struct {
	Dbg bool `long:"debug" env:"DEBUG" description:"debug mode"`

	Source SourceConfig

	RealMirrorURL string `long:"real-mirror-url" env:"REAL_MIRROR_URL" default:"https://ftp.belnet.be/mirror/jenkins/"`

	GetUpdateJSONBodyTimeout time.Duration `long:"timeout" env:"UPDATE_JSON_DOWNLOAD_TIMEOUT" default:"120s"`

	UpdateJSONCacheTTL time.Duration `long:"cache-ttl" env:"UPDATE_JSON_CACHE_TTL" default:"30m"`

	Signer SignerConfig
	Patch  PatchConfig
	Server ServerConfig

	DataDirPath string `long:"data-dir" env:"DATA_DIR" default:"/tmp/update-center-data"`
}

func (cfg AppConfig) validateSource() error {
	if cfg.Source.URL == "" && cfg.Source.Path == "" {
		return fmt.Errorf("either update.json URL or path must be configured")
	}

	if cfg.Source.URL != "" && cfg.Source.Path != "" {
		return fmt.Errorf("update.json URL and path cannot be used simultaneously")
	}

	return nil
}

func ParseConfig() (AppConfig, error) {
	cfg := AppConfig{}

	if _, err := flags.Parse(&cfg); err != nil {
		return AppConfig{}, err
	}

	cfg.Patch.OriginDownloadURL = strings.TrimSuffix(cfg.Patch.OriginDownloadURL, "/")
	cfg.Patch.NewDownloadURL = strings.TrimSuffix(cfg.Patch.NewDownloadURL, "/")

	if err := cfg.validateSource(); err != nil {
		return AppConfig{}, fmt.Errorf("invalid source: %w", err)
	}

	if err := os.MkdirAll(cfg.DataDirPath, 0755); err != nil {
		return AppConfig{}, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return cfg, nil
}
