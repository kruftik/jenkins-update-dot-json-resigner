package app

import (
	"context"
	"fmt"

	"jenkins-resigner-service/internal/logging"
	"jenkins-resigner-service/internal/server"

	"os"
	"os/signal"
	"syscall"

	"jenkins-resigner-service/internal/config"
	//"fmt"
	//"os"
	//"os/signal"
	//"syscall"
	//
	//"github.com/jessevdk/go-flags"
	//"go.uber.org/zap"
	"jenkins-resigner-service/internal/services/update_center"
	//"time"
)

var (
	//updateJSON *UpdateJSONT
	juc *update_center.JenkinsUCJSONT
)

func App() error {
	var (
		log = logging.GetLogger()
	)

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	update_center.Init(log)

	signInfo, err := update_center.ParseSigningParameters(
		config.Opts.SignCAPath,
		config.Opts.SignCertificatePath,
		config.Opts.SignKeyPath,
		config.Opts.SignKeyPassword,
	)
	if err != nil {
		return fmt.Errorf("cannot parse input args / envs: %w", err)
	}

	locationsOpts, err := update_center.ValidateUpdateJSONLocation(config.Opts.UpdateJSONURL, config.Opts.UpdateJSONPath)
	if err != nil {
		return fmt.Errorf("cannot parse update-center.json location: %w", err)
	}

	if locationsOpts.IsRemoteSource {
		locationsOpts.Timeout = config.Opts.UpdateJSONDownloadTimeout
	}

	jucOpts := update_center.JenkinsUCOpts{
		Src:      locationsOpts,
		CacheTtl: config.Opts.UpdateJSONCacheTTL,
		PatchOpts: update_center.JenkinsPatchOpts{
			From: config.Opts.OriginDownloadURL,
			To:   config.Opts.NewDownloadURL,
		},
		SigningInfo: signInfo,
	}

	juc, err = update_center.NewJenkinsUC(jucOpts)
	if err != nil {
		return fmt.Errorf("cannot initialize JenkinsUC object: %w", err)
	}

	// Shutting down handling...
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-c
		log.Infow("ResignerService shutting down")

		cancelFn()

		juc.Cleanup()

		os.Exit(0)
	}()

	//r, err := initHTTP(logger, juc)

	//log.Info("http server completed")

	srv, err := server.NewServer(ctx, log, server.Services{
		JUCPatcher: juc,
	})
	if err != nil {
		return fmt.Errorf("cannot init web server: %w", err)
	}

	if err = srv.Run(ctx); err != nil {
		return fmt.Errorf("cannot run web server: %w", err)
	}

	return nil
}
