package main

import (
	//"net/http"

	//"time"

	"log"

	"github.com/jessevdk/go-flags"
	"jenkins-resigner-service/internal/logging"

	"jenkins-resigner-service/internal/app"
	"jenkins-resigner-service/internal/config"
)

var (
	GitCommit = "0.0.1"
)

func main() {
	_, err := flags.Parse(&config.Opts)
	if err != nil {
		log.Fatalf("cannot parse flags: %v", err)
	}

	// Logging...
	log, err := logging.Configure(config.Opts)
	if err != nil {
		log.Fatalf("cannot configure logger: %v", err)
	}

	log.Infof("Jenkins update.json ResignerService (v%s) starting up...", GitCommit)

	err = app.App()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
}
