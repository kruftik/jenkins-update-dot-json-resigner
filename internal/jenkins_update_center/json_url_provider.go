package jenkins_update_center

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

var (
	MaxIdleConns    = 10
	IdleConnTimeout = 30 * time.Second
	Timeout         time.Duration
)

type urlJSONProvider struct {
	url      *url.URL
	metadata *JSONMetadataT

	content *UpdateJSON

	hc *http.Client
}

func ValidateURLJSONProviderSource(src string) error {
	resp, err := http.Head(src)
	if err != nil {
		return err
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			log.Warn(err)
		}
	}()

	return nil
}

func NewURLJSONProvider(sURL string) (*urlJSONProvider, error) {
	p := &urlJSONProvider{}

	if err := p.init(sURL); err != nil {
		return nil, errors.Wrap(err, "cannot call init function of URLJSONProvider")
	}

	// Warm up cache
	if _, _, err := p.GetContent(); err != nil {
		return nil, errors.Wrap(err, "cannot warm up  URLJSONProvider cache")
	}

	return p, nil
}

func (p *urlJSONProvider) init(src string) error {
	sURL, err := url.ParseRequestURI(src)
	if err != nil {
		return err
	}

	p.url = sURL
	p.metadata = &JSONMetadataT{}

	p.hc = &http.Client{
		// TODO: something goes wrong with this settings
		//Transport: &http.Transport{
		//	MaxIdleConns:    MaxIdleConns,
		//	IdleConnTimeout: IdleConnTimeout * time.Second,
		//},
		Timeout: Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			//l := len(via)
			log.Debugf("Got redirect to %s", req.URL.String())
			return nil
		},
	}

	log.Debugf("http client initialized with timeout %v", Timeout)

	return nil
}

func (p urlJSONProvider) getRemoteURLMetadata(r *http.Response) (*JSONMetadataT, error) {
	dt, err := http.ParseTime(r.Header.Get("Last-Modified"))
	if err != nil {
		return nil, errors.Wrapf(err, "%s is not valid datetime string", r.Header.Get("Last-Modified"))
	}

	meta := &JSONMetadataT{
		LastModified: dt,
		Size:         r.ContentLength,
		etag:         r.Header.Get("ETag"),
	}

	return meta, nil
}

func (p urlJSONProvider) GetFreshContent() (*UpdateJSON, *JSONMetadataT, error) {
	log.Infof("Downloading %s...", p.url)

	resp, err := p.hc.Get(p.url.String())
	if err != nil {
		return nil, nil, fmt.Errorf("cannot GET %s: %s", p.url.String(), err)
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			log.Warn(err)
		}
	}()

	jsonFileData := &bytes.Buffer{}

	n, err := jsonFileData.ReadFrom(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot save update.json content to buffer: %s", err)
	}

	log.Debugf("Successfully written %d bytes to buffer", n)

	uj, err := prepareUpdateJSONObject(jsonFileData.Bytes())
	if err != nil {
		return nil, nil, err
	}

	if err = uj.VerifySignature(); err != nil {
		return nil, nil, fmt.Errorf("signature of original update-center.json is not valid: %w", err)
	}

	meta, err := p.getRemoteURLMetadata(resp)
	if err != nil {
		return nil, nil, err
	}

	return uj, meta, nil
}

//func (p urlJSONProvider) getMetadata() (*JSONMetadataT, error) {
//	return p.metadata, nil
//}

func (p urlJSONProvider) GetFreshMetadata() (*JSONMetadataT, error) {
	resp, err := p.hc.Head(p.url.String())
	if err != nil {
		return nil, err
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Warn(err)
		}
	}()

	meta, err := p.getRemoteURLMetadata(resp)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

func (p *urlJSONProvider) RefreshMetadata(meta *JSONMetadataT) (*JSONMetadataT, error) {
	var err error

	if meta == nil {
		meta, err = p.GetFreshMetadata()
		if err != nil {
			return nil, err
		}
	}

	p.metadata = meta

	return meta, nil
}

func (p *urlJSONProvider) IsContentUpdated() (bool, error) {
	meta, err := p.GetFreshMetadata()
	if err != nil {
		return false, errors.Wrap(err, "cannot if metadata cache is still valid")
	}

	//isUpdated := meta.LastModified.After(p.metadata.LastModified) || meta.Size != p.metadata.Size
	isUpdated := meta.Size != p.metadata.Size || meta.etag != p.metadata.etag

	log.Debugf("Remote content isUpdated=%t", isUpdated)

	return isUpdated, nil
}

func (p *urlJSONProvider) GetContent() (*UpdateJSON, *JSONMetadataT, error) {
	isUpdated, err := p.IsContentUpdated()
	if err != nil {
		return nil, nil, err
	}

	if isUpdated {
		p.content, p.metadata, err = p.GetFreshContent()
		if err != nil {
			return nil, nil, err
		}
	}

	return p.content, p.metadata, nil
}
