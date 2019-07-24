package jenkins_update_center

import (
	"bytes"
	"fmt"
	"jenkins-resigner-service/jenkins_update_center/json_schema"
	"net/http"
	"net/url"
	"time"
)

var (
	MaxIdleConns    = 10
	IdleConnTimeout = 30 * time.Second
	Timeout         = 30 * time.Second
)

type urlJSONProvider struct {
	url      *url.URL
	metadata *JSONMetadataT

	hc *http.Client
}

func ValidateURLJSONProviderSource(src string) error {
	resp, err := http.Head(src)
	if err != nil {
		return err
	}
	defer func() {
		err = resp.Body.Close()
		log.Warn(err)
	}()

	return nil
}

func NewURLJSONProvider(sURL string) (*urlJSONProvider, error) {
	p := &urlJSONProvider{}

	if err := p.init(sURL); err != nil {
		return nil, err
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
		Transport: &http.Transport{
			MaxIdleConns:    MaxIdleConns,
			IdleConnTimeout: IdleConnTimeout * time.Second,
		},
		Timeout: Timeout,
	}

	return nil
}

func (p urlJSONProvider) GetContent() (*json_schema.UpdateJSON, error) {
	log.Infof("Downloading %s...", p.url)

	resp, err := http.Get(p.url.String())
	if err != nil {
		return nil, fmt.Errorf("cannot GET %s: %s", p.url.String(), err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	jsonFileData := &bytes.Buffer{}

	n, err := jsonFileData.ReadFrom(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot save update.json content to buffer: %s", err)
	}

	log.Debugf("Successfully written %d bytes to buffer", n)

	return prepareUpdateJSONObject(jsonFileData.Bytes())
}

func (p urlJSONProvider) getFreshMetadata() (*JSONMetadataT, error) {
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

	dt, err := http.ParseTime(resp.Header.Get("Last-Modified"))
	if err != nil {
		return nil, err
	}

	meta := &JSONMetadataT{
		LastModified: dt,
		Size:         resp.ContentLength,
	}

	return meta, nil
}

func (p urlJSONProvider) GetMetadata() (*JSONMetadataT, error) {
	meta, err := p.getFreshMetadata()
	if err != nil {
		return nil, err
	}

	p.metadata = meta

	return p.metadata, nil
}

func (p *urlJSONProvider) IsContentUpdated() (bool, error) {
	meta, err := p.getFreshMetadata()
	if err != nil {
		return false, err
	}

	isUpdated := meta.LastModified.After(p.metadata.LastModified) || meta.Size != p.metadata.Size

	return isUpdated, nil
}
