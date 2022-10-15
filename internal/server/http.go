package server

import (
	"time"
)

const (
	timeoutTotal = 180 * time.Second
)

//func initHTTP(logger *zap.Logger, juc *update_center.JenkinsUCJSONT) (*chi.Mux, error) {
//	var (
//		log = logging.GetLogger()
//	)
//
//	log.Info("Running http server... ")
//
//	r := chi.NewRouter()
//
//	r.Use(middleware.Heartbeat("/healthz"))
//
//	// Регистрация pprof-обработчиков
//	r.HandleFunc("/debug/pprof/", pprof.Index)
//	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
//	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
//	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
//	r.HandleFunc("/debug/pprof/trace", pprof.Trace)
//
//	proxy, err := initProxy()
//	if err != nil {
//		return nil, err
//	}
//
//
//
//	return r, nil
//}
