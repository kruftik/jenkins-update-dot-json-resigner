package main

import (
	//"net/http"

	//"time"

	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"go.uber.org/zap"

	"jenkins-resigner-service/internal/app"
	"jenkins-resigner-service/internal/config"
)

func main() {
	_, err := flags.Parse(&config.Opts)
	if err != nil {
		fmt.Println("Can't parse flags: ", err)
		os.Exit(1)
	}

	logger, _ := zap.NewDevelopment()
	defer func() {
		_ = logger.Sync()
	}()

	zap.ReplaceGlobals(logger)
	log := zap.S()

	log.Infof("Jenkins update.json schema validator starting up...")

	err = app.App(logger)
	log.Infof(err.Error())
}