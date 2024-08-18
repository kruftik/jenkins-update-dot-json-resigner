package app

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/config"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/patcher"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/signer"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/sourcefileproviders"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/sourcefileproviders/localfile"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/sourcefileproviders/remoteurl"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/server"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins"
)

func App(ctx context.Context, version string) error {
	cfg, err := config.ParseConfig()
	if err != nil {
		return fmt.Errorf("cannot parse config: %w", err)
	}

	var logger *zap.Logger

	// Logging...
	if cfg.Dbg {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	defer func() {
		_ = logger.Sync()
	}()

	log := logger.Sugar()

	log.Infof("Jenkins update.json ResignerService (v%s) starting up...", version)

	var sourceFileProvider sourcefileproviders.SourceFileProvider

	if cfg.UpdateJSONURL != "" {
		sourceFileProvider, err = remoteurl.NewRemoteURLProvider(log, cfg.UpdateJSONURL)
	} else {
		sourceFileProvider, err = localfile.NewLocalFileProvider(cfg.UpdateJSONPath)
	}
	if err != nil {
		return fmt.Errorf("cannot initialize source file provider: %w", err)
	}

	signer, err := signer.NewSignerService(log, cfg.SignCAPath, cfg.SignCertificatePath, cfg.SignKeyPath, cfg.SignKeyPassword)
	if err != nil {
		return fmt.Errorf("cannot initialize signer: %w", err)
	}

	patchers := []patcher.Patcher{
		patcher.NewPatcher(log, cfg.OriginDownloadURL, cfg.NewDownloadURL),
	}

	juc, err := jenkins.NewJenkinsUpdateCenter(ctx, log, cfg, sourceFileProvider, signer, patchers)
	if err != nil {
		return fmt.Errorf("cannot initialize jenkins update center: %w", err)
	}

	srv, err := server.NewServer(log, juc, cfg.ServerAddr+":"+strconv.Itoa(cfg.ServerPort), cfg.DataDirPath, cfg.NewDownloadURL)
	if err != nil {
		return fmt.Errorf("cannot initialize server: %w", err)
	}

	if err := srv.ListenAndServe(ctx); err != nil {
		return errors.Wrapf(err, "ResignerService http server terminated: %s", err)
	}

	if err := juc.CleanUp(context.Background()); err != nil {
		return fmt.Errorf("cannot clean up: %w", err)
	}

	return nil
}
