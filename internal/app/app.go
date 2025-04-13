package app

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/config"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/patcher"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/signer"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/sourcefileproviders"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/sourcefileproviders/cache"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/sourcefileproviders/localfile"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/sourcefileproviders/remoteurl"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/types"
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

	var sourceFileProvider sourcefileproviders.Provider

	if cfg.Source.URL != "" {
		sourceFileProvider, err = remoteurl.NewRemoteURLProvider(log.With("component", "remote-url-provider"), cfg.Source.URL)

		if cfg.UpdateJSONCacheTTL > 0 {
			log.Infof("initializing caching wrapper (cache TTL = %s)", cfg.UpdateJSONCacheTTL)
			sourceFileProvider, err = cache.NewCacheWrapper(ctx, log.With("component", "cache-wrapper"), sourceFileProvider, cfg.UpdateJSONCacheTTL)
			if err != nil {
				return fmt.Errorf("cannot initialize cache wrapper: %w", err)
			}
		}
	} else {
		sourceFileProvider, err = localfile.NewLocalFileProvider(cfg.Source.Path)
	}
	if err != nil {
		return fmt.Errorf("cannot initialize source file provider: %w", err)
	}

	signerSvc, err := signer.NewSignerService(log.With("component", "signer"), cfg.Signer)
	if err != nil {
		return fmt.Errorf("cannot initialize signer: %w", err)
	}

	patchers := []types.Patcher{
		patcher.NewPatcher(log.With("component", "patcher"), cfg.Patch),
	}

	juc := jenkins.NewJenkinsUpdateCenter(log.With("component", "juc"), cfg, sourceFileProvider, signerSvc, patchers)

	if err := juc.RefreshContent(ctx); err != nil {
		return fmt.Errorf("cannot refresh content: %w", err)
	}

	defer func() {
		if err := juc.CleanUp(context.Background()); err != nil {
			log.Warnf(fmt.Sprintf("cannot clean up: %v", err))
		}
	}()

	srv, err := server.NewServer(log.With("component", "server"), cfg.Server, juc, cfg.DataDirPath, cfg.RealMirrorURL)
	if err != nil {
		return fmt.Errorf("cannot initialize server: %w", err)
	}

	if err := srv.ListenAndServe(ctx); err != nil {
		return fmt.Errorf("ResignerService http server terminated: %w", err)
	}

	return nil
}
