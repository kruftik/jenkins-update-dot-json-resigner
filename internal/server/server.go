package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/config"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins"
)

type Server struct {
	log *zap.SugaredLogger

	cfg config.ServerConfig

	patchedFileProvider jenkins.PatchedFileRefresher

	dataDir    string
	proxyToURL string

	srv *http.Server
}

func NewServer(log *zap.SugaredLogger, cfg config.ServerConfig, jsonFileProvider jenkins.PatchedFileRefresher, dataDir, proxyToURL string) (Server, error) {
	s := Server{
		log:                 log,
		cfg:                 cfg,
		patchedFileProvider: jsonFileProvider,
		dataDir:             dataDir,
		proxyToURL:          proxyToURL,
	}

	handlers, err := s.getHandlers()
	if err != nil {
		return Server{}, fmt.Errorf("could not initialize handlers: %w", err)
	}

	s.srv = &http.Server{
		Addr:    cfg.ListenAddr + ":" + strconv.Itoa(cfg.ListenPort),
		Handler: handlers,
	}

	return s, nil
}

func (s Server) ListenAndServe(ctx context.Context) error {
	go func(ctx context.Context) {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			s.log.Warnf("cannot gracefully shutdown http server: %v", err)
		}
	}(ctx)

	if s.cfg.TLSCertPath != "" && s.cfg.TLSKeyPath != "" {
		s.log.Infof("starting https server on %s:%d", s.cfg.ListenAddr, s.cfg.ListenPort)

		if err := s.srv.ListenAndServeTLS(s.cfg.TLSCertPath, s.cfg.TLSKeyPath); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				return err
			}
		}
		return nil
	}

	s.log.Infof("starting http server on %s:%d", s.cfg.ListenAddr, s.cfg.ListenPort)

	if err := s.srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	return nil
}
