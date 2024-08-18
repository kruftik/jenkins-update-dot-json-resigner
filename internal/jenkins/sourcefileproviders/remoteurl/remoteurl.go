package remoteurl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/sourcefileproviders"
)

var (
	_ sourcefileproviders.Provider = (*Provider)(nil)
)

type Provider struct {
	log *zap.SugaredLogger

	url string

	hc *http.Client
}

func NewRemoteURLProvider(log *zap.SugaredLogger, sURL string) (*Provider, error) {
	p := &Provider{
		log: log,
		url: sURL,
	}

	if _, err := url.ParseRequestURI(sURL); err != nil {
		return nil, fmt.Errorf("failed to parse source URL %q: %w", sURL, err)
	}

	p.hc = &http.Client{
		// TODO: something goes wrong with this settings
		//Transport: &http.Transport{
		//	MaxIdleConns:    MaxIdleConns,
		//	IdleConnTimeout: IdleConnTimeout * time.Second,
		//},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			p.log.Debugf("%s %s: redirected to %s", req.Method, via[0].URL.String(), req.URL.String())
			return nil
		},
	}

	if err := p.validate(sURL); err != nil {
		return nil, fmt.Errorf("failed to validate source URL %q: %w", sURL, err)
	}

	return p, nil
}

func (p *Provider) validate(src string) error {
	resp, err := http.Head(src)
	if err != nil {
		return err
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			p.log.Warn(err)
		}
	}()

	return nil
}

func (p *Provider) getRemoteURLMetadata(r *http.Response) (sourcefileproviders.FileMetadata, error) {
	dt, err := http.ParseTime(r.Header.Get("Last-Modified"))
	if err != nil {
		return sourcefileproviders.FileMetadata{}, errors.Wrapf(err, "%s is not valid datetime string", r.Header.Get("Last-Modified"))
	}

	meta := sourcefileproviders.FileMetadata{
		LastModified: dt,
		Size:         r.ContentLength,
		Etag:         r.Header.Get("ETag"),
	}

	return meta, nil
}

func (p *Provider) GetMetadata(ctx context.Context) (sourcefileproviders.FileMetadata, error) {
	p.log.Debugf("HEAD %s...", p.url)

	req, err := http.NewRequest(http.MethodHead, p.url, nil)
	if err != nil {
		return sourcefileproviders.FileMetadata{}, fmt.Errorf("cannot create request: %w", err)
	}

	resp, err := p.hc.Do(req.WithContext(ctx))
	if err != nil {
		return sourcefileproviders.FileMetadata{}, fmt.Errorf("cannot HEAD %s: %w", p.url, err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			p.log.Warn(err)
		}
	}()

	return p.getRemoteURLMetadata(resp)
}

func (p *Provider) GetBody(ctx context.Context) (sourcefileproviders.FileMetadata, io.ReadCloser, error) {
	p.log.Debugf("GET %s...", p.url)

	req, err := http.NewRequest(http.MethodGet, p.url, nil)
	if err != nil {
		return sourcefileproviders.FileMetadata{}, nil, fmt.Errorf("cannot create request: %w", err)
	}

	// We need content-length header in response
	req.Header.Set("Accept-Encoding", "identity")

	resp, err := p.hc.Do(req.WithContext(ctx))
	if err != nil {
		return sourcefileproviders.FileMetadata{}, nil, fmt.Errorf("cannot GET %s: %w", p.url, err)
	}
	defer resp.Body.Close()

	metadata, err := p.getRemoteURLMetadata(resp)
	if err != nil {
		return sourcefileproviders.FileMetadata{}, nil, fmt.Errorf("failed to get remote URL metadata: %w", err)
	}

	f, err := os.CreateTemp("", "update-center-remote-url.*.jsonp")
	if err != nil {
		return sourcefileproviders.FileMetadata{}, nil, fmt.Errorf("cannot create temporary file: %w", err)
	}

	p.log.Debugf("%s temporary file created", f.Name())

	length, err := io.Copy(f, resp.Body)
	if err != nil {
		return sourcefileproviders.FileMetadata{}, nil, fmt.Errorf("cannot write body to temporary file: %w", err)
	}

	r, err := sourcefileproviders.NewJSONPTrailersStrippingReader(TempFileReadCloser{f}, length)
	if err != nil {
		return sourcefileproviders.FileMetadata{}, nil, fmt.Errorf("failed to create JSONP trailer reader: %w", err)
	}

	return metadata, r, nil
}
