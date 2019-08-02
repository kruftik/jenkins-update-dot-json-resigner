package main

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	l "github.com/treastech/logger"
	"jenkins-resigner-service/pkg/jenkins_update_center"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	"net/http/pprof"
)

const (
	timeoutTotal = 15 * time.Second
)

func initProxy() (*httputil.ReverseProxy, error) {
	originURL, err := url.ParseRequestURI(Opts.NewDownloadURI)
	if err != nil {
		log.Warn("origin URL is incorrect: ", err)
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(originURL)

	return proxy, nil
}

func initHTTP(juc *jenkins_update_center.JenkinsUCJSONT) (*chi.Mux, error) {
	log.Info("Running http server... ")

	r := chi.NewRouter()

	r.Use(middleware.Heartbeat("/healthz"))

	// Регистрация pprof-обработчиков
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	proxy, err := initProxy()
	if err != nil {
		return nil, err
	}

	r.Group(func(r chi.Router) {
		r.Use(middleware.RealIP)
		r.Use(middleware.Recoverer)

		r.Use(middleware.Timeout(timeoutTotal))

		r.Use(l.Logger(logger))

		r.Get("/*", proxy.ServeHTTP)

		r.Get(UpdateCenterDotJSON, func(w http.ResponseWriter, r *http.Request) {
			c, err := juc.GetPatchedAndSigned()
			if err != nil {
				log.Warn(err)
				return
			}

			cl := strconv.Itoa(len(c))
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Length", cl)
			w.Header().Set("Etag", "update-center-json-"+cl)

			if _, err = w.Write(c); err != nil {
				log.Warn(err)
				return
			}
		})
	})

	return r, nil
}
