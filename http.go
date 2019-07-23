package main

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	l "github.com/treastech/logger"
	"jenkins-resigner-service/jenkins_update_center"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"
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

func initHTTP(juc *jenkins_update_center.JenkinsUCJSONT) error {
	log.Info("Running http server... ")

	r := chi.NewRouter()

	r.Use(middleware.Heartbeat("/healthz"))

	proxy, err := initProxy()
	if err != nil {
		return err
	}

	r.Group(func(r chi.Router) {
		r.Use(middleware.RealIP)
		r.Use(middleware.Recoverer)

		r.Use(middleware.Timeout(timeoutTotal))

		r.Use(l.Logger(logger))

		r.Get("/*", proxy.ServeHTTP)

		r.Get(UpdateCenterDotJSON, func(w http.ResponseWriter, r *http.Request) {

		})
	})

	port := strconv.Itoa(Opts.ServerPort)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		return fmt.Errorf("ResignerService http server terminated: %s", err)
	}

	log.Info("http server completed")

	return nil
}
