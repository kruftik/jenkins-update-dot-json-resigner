package jenkins_update_center

import (
	"jenkins-resigner-service/jenkins_update_center/json_schema"
	"os"
)

type localFileJSONProvider struct {
	src      string
	metadata *JSONMetadataT
}

func (p *localFileJSONProvider) Init(src string) error {
	p.src = src
	p.metadata = &JSONMetadataT{}

	return nil
}

func (p *localFileJSONProvider) GetContent() (*json_schema.UpdateJSON, error) {
	return nil, nil
}

func (p localFileJSONProvider) getFreshMetadata() (*JSONMetadataT, error) {
	fi, err := os.Stat(p.src)
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
