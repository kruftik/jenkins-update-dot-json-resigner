package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
	"jenkins-resigner-service/internal/config"
	"jenkins-resigner-service/internal/logging"
	"jenkins-resigner-service/internal/server/middleware"
	"jenkins-resigner-service/internal/services/httpproxy"
	"jenkins-resigner-service/internal/services/update_center"
)

var (
	shutdownTimeout = 5 * time.Second
)

type Services struct {
	JUCPatcher update_center.IJenkinsUpdateCenterPatcher
}

type Server struct {
	log *zap.SugaredLogger

	svc Services

	listen     int
	httpServer *http.Server
	router     *chi.Mux
}

const (
	UpdateCenterDotJSON = "/update-center.json"
	UpdateCenterDotHTML = "/update-center.json.html"
)

func NewServer(ctx context.Context, log *zap.SugaredLogger, svc Services) (*Server, error) {
	srv := &Server{
		log: log,

		svc: svc,

		listen: config.Opts.ServerPort,
		router: chi.NewRouter(),
	}

	if err := srv.internalRoutes(ctx); err != nil {
		return nil, fmt.Errorf("cannot configure internal routes: %w", err)
	}

	srv.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", srv.listen),
		Handler: srv.SetupRoutes(),
		//ReadHeaderTimeout: 5 * time.Second,
		//IdleTimeout:       30 * time.Second,
	}

	return srv, nil
}

func (s *Server) SetupRoutes() http.Handler {
	s.router.Get("/updates/hudson.tasks.*", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	s.router.Get("/updates/hudson.tools.*", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	s.router.Get(UpdateCenterDotJSON, func(w http.ResponseWriter, r *http.Request) {
		c, err := s.svc.JUCPatcher.GetPatchedAndSignedJSONP()
		if err != nil {
			s.log.Warn(err)
			return
		}

		cl := strconv.Itoa(len(c))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", cl)
		w.Header().Set("Etag", "update-center-json-"+cl)

		if _, err = w.Write(c); err != nil {
			s.log.Warn(err)
			return
		}
	})

	s.router.Get(UpdateCenterDotHTML, func(w http.ResponseWriter, r *http.Request) {
		c, err := s.svc.JUCPatcher.GetPatchedAndSignedHTML()
		if err != nil {
			s.log.Warn(err)
			return
		}

		cl := strconv.Itoa(len(c))
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Content-Length", cl)
		w.Header().Set("Etag", "update-center-json-html-"+cl)

		if _, err = w.Write(c); err != nil {
			s.log.Warn(err)
			return
		}
	})

	return s.router
}

func (s *Server) internalRoutes(ctx context.Context) error {
	s.router.Use(chimiddleware.Heartbeat("/healthz"))

	s.router.Use(middleware.Logger(s.log.Desugar()))

	s.router.Use(chimiddleware.RealIP)
	s.router.Use(chimiddleware.Recoverer)

	s.router.Use(chimiddleware.Timeout(timeoutTotal))

	// Регистрация pprof-обработчиков
	s.router.HandleFunc("/debug/pprof/", pprof.Index)
	s.router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	s.router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	s.router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	s.router.HandleFunc("/debug/pprof/trace", pprof.Trace)

	proxy, err := httpproxy.NewHTTPProxy(config.Opts.OriginDownloadURL)
	if err != nil {
		return fmt.Errorf("cannot init http proxy: %w to upstream", err)
	}

	s.router.Get("/*", proxy.ServeHTTP)

	return nil
}

func (s *Server) Run(ctx context.Context) error {
	var (
		log = logging.GetLogger()
	)

	log.Infow("server started",
		"listen", s.listen,
	)

	go func() {
		<-ctx.Done()

		log.Debug("shutdown initiated")

		shutdownCtx, done := context.WithTimeout(context.Background(), shutdownTimeout)

		defer done()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			log.Error(fmt.Errorf("http shutdown error: %w", err))
		}

		log.Debug("shutdown completed")
	}()

	s.httpServer.Handler = s.router

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to serve: %w", err)
	}

	log.Info("stop http server")

	return nil
}
