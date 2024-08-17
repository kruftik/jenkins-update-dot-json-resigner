package app

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/config"
	//"fmt"
	//"os"
	//"os/signal"
	//"syscall"
	//
	//"github.com/jessevdk/go-flags"
	//"go.uber.org/zap"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins_update_center"
	//"time"
)

var (
	log *zap.SugaredLogger

	//updateJSON *UpdateJSONT
	juc *jenkins_update_center.JenkinsUCJSONT
)

func App(logger *zap.Logger) error {
	zap.ReplaceGlobals(logger)
	log = zap.S()

	jenkins_update_center.Init(log)

	signInfo, err := jenkins_update_center.ParseSigningParameters(
		config.Opts.SignCAPath,
		config.Opts.SignCertificatePath,
		config.Opts.SignKeyPath,
		config.Opts.SignKeyPassword,
	)
	if err != nil {
		return errors.Wrap(err, "cannot parse input args / envs")
	}

	locationsOpts, err := jenkins_update_center.ValidateUpdateJSONLocation(config.Opts.UpdateJSONURL, config.Opts.UpdateJSONPath)
	if err != nil {
		return fmt.Errorf("cannot parse update-center.json location: %w", err)
	}

	if locationsOpts.IsRemoteSource {
		locationsOpts.Timeout = config.Opts.UpdateJSONDownloadTimeout
	}

	jucOpts := jenkins_update_center.JenkinsUCOpts{
		Src:      locationsOpts,
		CacheTtl: config.Opts.UpdateJSONCacheTTL,
		PatchOpts: jenkins_update_center.JenkinsPatchOpts{
			From: config.Opts.OriginDownloadURL,
			To:   config.Opts.NewDownloadURL,
		},
		SigningInfo: signInfo,
	}

	juc, err = jenkins_update_center.NewJenkinsUC(jucOpts)
	if err != nil {
		return fmt.Errorf("cannot initialize JenkinsUC object: %w", err)
	}

	// Shutting down handling...
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-c
		log.Infow("ResignerService shutting down")

		juc.Cleanup()

		os.Exit(0)
	}()

	r, err := initHTTP(logger, juc)
	if err != nil {
		return errors.Wrap(err, "cannot initialize HTTP-server")
	}

	if err := http.ListenAndServe(":"+strconv.Itoa(config.Opts.ServerPort), r); err != nil {
		return errors.Wrapf(err, "ResignerService http server terminated: %s", err)
	}

	log.Info("http server completed")

	return nil
}
