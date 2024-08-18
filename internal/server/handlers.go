package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/http/pprof"
	"net/url"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	l "github.com/treastech/logger"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins"
)

const (
	timeoutTotal = 15 * time.Second
)

func (s Server) httpProxy(proxyToURL string) (*httputil.ReverseProxy, error) {
	originURL, err := url.ParseRequestURI(proxyToURL)
	if err != nil {
		return nil, fmt.Errorf("origin URL is incorrect: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(originURL)

	return proxy, nil
}

func (s Server) getHandlers() (*chi.Mux, error) {
	r := chi.NewRouter()

	r.Use(middleware.Heartbeat("/healthz"))

	// Регистрация pprof-обработчиков
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	proxy, err := s.httpProxy(s.proxyToURL)
	if err != nil {
		return nil, err
	}

	r.Group(func(r chi.Router) {
		r.Use(middleware.RealIP)
		r.Use(middleware.Recoverer)

		r.Use(middleware.Timeout(timeoutTotal))

		r.Use(l.Logger(s.log.Desugar()))

		r.Get("/*", proxy.ServeHTTP)

		fsHandler := http.FileServer(http.Dir(s.dataDir))

		r.Group(func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if err := s.patchedFileProvider.RefreshContent(r.Context()); err != nil {
						s.log.Errorf("failed to refresh content: %v", err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					next.ServeHTTP(w, r)
				})
			})

			r.Get("/"+jenkins.UpdateCenterDotJSON, fsHandler.ServeHTTP)
			r.Get("/"+jenkins.UpdateCenterDotHTML, fsHandler.ServeHTTP)
		})

		//r.Get(updateCenterDotJSON, func(w http.ResponseWriter, r *http.Request) {
		//	length, body, err := s.patchedFileProvider.GetJSONP(r.Context())
		//	if err != nil {
		//		s.log.Errorf("cannot get patched file: %v", err)
		//		w.WriteHeader(http.StatusInternalServerError)
		//		return
		//	}
		//
		//	cl := strconv.Itoa(length)
		//	w.Header().Set("Content-Type", "application/json")
		//	w.Header().Set("Content-Length", cl)
		//	w.Header().Set("Etag", "update-center-json-"+cl)
		//
		//	if _, err := io.Copy(w, body); err != nil {
		//		s.log.Warnf("cannot write response: %v", err)
		//		w.WriteHeader(http.StatusInternalServerError)
		//		return
		//	}
		//})

		//r.Get(updateCenterDotHTML, func(w http.ResponseWriter, r *http.Request) {
		//	length, body, err := s.patchedFileProvider.GetHTML(r.Context())
		//	if err != nil {
		//		s.log.Errorf("cannot get patched file: %v", err)
		//		w.WriteHeader(http.StatusInternalServerError)
		//		return
		//	}
		//
		//	cl := strconv.Itoa(length)
		//	w.Header().Set("Content-Type", "text/html")
		//	w.Header().Set("Content-Length", cl)
		//	w.Header().Set("Etag", "update-center-json-html-"+cl)
		//
		//	if _, err := io.Copy(w, body); err != nil {
		//		s.log.Warnf("cannot write response: %v", err)
		//		w.WriteHeader(http.StatusInternalServerError)
		//		return
		//	}
		//})

		r.Get("/updates/hudson.tasks.*", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		r.Get("/updates/hudson.tools.*", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
	})

	return r, nil
}
