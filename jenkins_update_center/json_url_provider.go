package jenkins_update_center

import (
	"jenkins-resigner-service/jenkins_update_center/json_schema"
	"net/http"
	"net/url"
	"time"
)

type urlJSONProvider struct {
	src      *url.URL
	metadata *JSONMetadataT

	hc *http.Client
}

func (p *urlJSONProvider) Init(src string) error {
	sURL, err := url.ParseRequestURI(src)
	if err != nil {
		return err
	}

	p.src = sURL
	p.metadata = &JSONMetadataT{}

	p.hc = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 30 * time.Second,
		},
		Timeout: 30 * time.Second,
	}

	return nil
}

func (p urlJSONProvider) GetContent() (*json_schema.UpdateJSON, error) {
	return nil, nil
}

func (p urlJSONProvider) getFreshMetadata() (*JSONMetadataT, error) {
	resp, err := p.hc.Head(p.src.String())
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
