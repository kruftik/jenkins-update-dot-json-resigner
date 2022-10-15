package server

import (
	"fmt"
	"net/http/httputil"
	"net/url"

	"jenkins-resigner-service/internal/config"
)

func initProxy() (*httputil.ReverseProxy, error) {
	originURL, err := url.ParseRequestURI(config.Opts.NewDownloadURL)
	if err != nil {
		return nil, fmt.Errorf("origin URL is incorrect: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(originURL)

	return proxy, nil
}