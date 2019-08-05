package main

import (
	//"net/http"

	//"time"

	"fmt"
	"github.com/jessevdk/go-flags"
	"go.uber.org/zap"
	"jenkins-resigner-service/pkg/jenkins_update_center"
	"os"
	"time"
)

const (
	UpdateCenterDotJSON = "/update-center.json"
)

var (
	GitCommit = "0.0.1"

	logger *zap.Logger
	log    *zap.SugaredLogger
	// Opts with all cli commands and flags

	Opts = struct {
		Dbg bool `long:"debug" env:"DEBUG" description:"debug mode"`

		UpdateJSONPath string `long:"update-json-path"  env:"UPDATE_JSON_PATH"`
		UpdateJSONURL  string `long:"update-json-url" env:"UPDATE_JSON_URL"`

		UpdateJSONCacheTTL time.Duration `long:"cache-ttl" env:"UPDATE_JSON_CACHE_TTL" default:"30m"`

		OriginDownloadURL string `long:"origin-download-uri" env:"ORIGIN_DOWNLOAD_URL" default:"http://updates.jenkins-ci.org/"`
		NewDownloadURL    string `long:"new-download-uri" env:"NEW_DOWNLOAD_URL" required:"true"`

		SignCAPath          string `long:"ca-certificate-path" env:"SIGN_CA_PATH" description:"x509 CA certificates path"`
		SignCertificatePath string `long:"certificate-path" env:"SIGN_CERTIFICATE_PATH" description:"x509-certificate path" required:"true"`
		SignKeyPath         string `long:"key-path" env:"SIGN_KEY_PATH" description:"private key path" required:"true"`
		SignKeyPassword     string `long:"private-key-pass" env:"SIGN_KEY_PASSWORD"`

		ServerPort int `long:"listen-port" env:"LISTEN_PORT" default:"8282"`
	}{}

	//updateJSON *UpdateJSONT
	juc *jenkins_update_center.JenkinsUCJSONT
)

func main() {
	_, err := flags.Parse(&Opts)
	if err != nil {
		fmt.Println("Can't parse flags: ", err)
		os.Exit(1)
	}

	// Logging...
	if Opts.Dbg {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	defer func() {
		_ = logger.Sync()
	}()

	zap.ReplaceGlobals(logger)
	log = zap.S()

	log.Infof("Jenkins update.json ResignerService (v%s) starting up...", GitCommit)

	jenkins_update_center.Init()

	err = initialize()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
}
