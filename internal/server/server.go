package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins"
)

type Server struct {
	log *zap.SugaredLogger

	patchedFileProvider jenkins.PatchedFileRefresher

	dataDir    string
	proxyToURL string

	srv *http.Server
}

func NewServer(log *zap.SugaredLogger, jsonFileProvider jenkins.PatchedFileRefresher, addr, dataDir, proxyToURL string) (Server, error) {
	s := Server{
		log:                 log,
		patchedFileProvider: jsonFileProvider,
		dataDir:             dataDir,
		proxyToURL:          proxyToURL,
	}

	handlers, err := s.getHandlers()
	if err != nil {
		return Server{}, fmt.Errorf("could not initialize handlers: %w", err)
	}

	s.srv = &http.Server{
		Addr:    addr,
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

	if err := s.srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	return nil
}
