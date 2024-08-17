package main

import (
	//"net/http"

	//"time"

	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"go.uber.org/zap"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/app"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/config"
)

var (
	GitCommit = "0.0.1"

	logger *zap.Logger
	log    *zap.SugaredLogger
)

func main() {
	_, err := flags.Parse(&config.Opts)
	if err != nil {
		fmt.Println("Can't parse flags: ", err)
		os.Exit(1)
	}

	// Logging...
	if config.Opts.Dbg {
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

	err = app.App(logger)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
}
