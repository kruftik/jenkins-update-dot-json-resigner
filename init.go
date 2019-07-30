package main

import (
	"github.com/pkg/errors"

	//"fmt"
	//"os"
	//"os/signal"
	//"syscall"
	//
	//"github.com/jessevdk/go-flags"
	//"go.uber.org/zap"
	"jenkins-resigner-service/jenkins_update_center"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	//"time"
)

func initialize() error {
	signInfo, err := jenkins_update_center.ParseSigningParameters(
		Opts.SignCAPath,
		Opts.SignCertificatePath,
		Opts.SignKeyPath,
		Opts.SignKeyPassword,
	)
	if err != nil {
		return err
	}

	locationsOpts, err := jenkins_update_center.ValidateUpdateJSONLocation(Opts.UpdateJSONURL, Opts.UpdateJSONPath)
	if err != nil {
		return err
	}

	jucOpts := jenkins_update_center.JenkinsUCOpts{
		Src:      locationsOpts,
		CacheTtl: Opts.UpdateJSONCacheTTL,
		PatchOpts: jenkins_update_center.JenkinsPatchOpts{
			From: Opts.OriginDownloadURI,
			To:   Opts.NewDownloadURI,
		},
		SigningInfo: signInfo,
	}

	juc, err = jenkins_update_center.NewJenkinsUC(jucOpts)
	if err != nil {
		return err
	}

	// Shutting down handling...
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-c
		log.Infow("ResignerService shutting down")

		juc.Cleanup()

		os.Exit(0)
	}()

	r, err := initHTTP(juc)
	if err != nil {
		return err
	}

	if err := http.ListenAndServe(":"+strconv.Itoa(Opts.ServerPort), r); err != nil {
		return errors.Wrapf(err, "ResignerService http server terminated: %s", err)
	}

	log.Info("http server completed")

	return nil
}
