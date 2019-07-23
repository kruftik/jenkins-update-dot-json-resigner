package jenkins_update_center

import (
	"jenkins-resigner-service/jenkins_update_center/json_schema"
	"os"
)

type localFileJSONProvider struct {
	path     string
	metadata *JSONMetadataT
}

func ValidateLocalFileJSONProviderSource(src string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		err = f.Close()
		log.Warn(err)
	}()

	return nil
}

func NewLocalFileJSONProvider(path string) (*localFileJSONProvider, error) {
	p := &localFileJSONProvider{}

	if err := p.init(path); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *localFileJSONProvider) init(src string) error {
	p.path = src
	p.metadata = &JSONMetadataT{}

	return nil
}

func (p *localFileJSONProvider) GetContent() (*json_schema.UpdateJSON, error) {
	return nil, nil
}

func (p localFileJSONProvider) getFreshMetadata() (*JSONMetadataT, error) {
	fi, err := os.Stat(p.path)
	if err != nil {
		return nil, err
	}

	meta := &JSONMetadataT{
		LastModified: fi.ModTime(),
		Size:         fi.Size(),
	}

	return meta, nil
}

func (p *localFileJSONProvider) GetMetadata() (*JSONMetadataT, error) {
	meta, err := p.getFreshMetadata()
	if err != nil {
		return nil, err
	}

	p.metadata = meta

	return meta, nil
}

func (p *localFileJSONProvider) IsContentUpdated() (bool, error) {
	meta, err := p.getFreshMetadata()
	if err != nil {
		return false, err
	}

	isUpdated := (p.metadata.LastModified == meta.LastModified) && (p.metadata.Size == meta.Size)

	return isUpdated, nil
}
