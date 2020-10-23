package jenkins_update_center

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
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
		if err = f.Close(); err != nil {
			log.Warn(err)
		}
	}()

	return nil
}

func NewLocalFileJSONProvider(path string) (*localFileJSONProvider, error) {
	p := &localFileJSONProvider{}

	if err := p.init(path); err != nil {
		return nil, errors.Wrap(err, "cannot call init function of LocalFileJSONProvider")
	}

	// Warm up cache
	if _, _, err := p.GetContent(); err != nil {
		return nil, errors.Wrap(err, "cannot warm up LocalFileJSONProvider cache")
	}

	return p, nil
}

func (p *localFileJSONProvider) init(src string) error {
	p.path = src

	return nil
}

func (p *localFileJSONProvider) GetFreshContent() (*UpdateJSON, *JSONMetadataT, error) {
	log.Infof("Reading %s...", p.path)

	sbytes, err := ioutil.ReadFile(p.path)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "cannot read update.json content: %s", err)
	}

	uj, err := prepareUpdateJSONObject(sbytes)
	if err != nil {
		return nil, nil, err
	}

	if err = uj.VerifySignature(); err != nil {
		return nil, nil, err
	}

	meta, err := p.GetFreshMetadata()

	return uj, meta, nil
}

//func (p localFileJSONProvider) getMetadata() (*JSONMetadataT, error) {
//	return p.metadata, nil
//}

func (p localFileJSONProvider) GetFreshMetadata() (*JSONMetadataT, error) {
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

func (p *localFileJSONProvider) RefreshMetadata(meta *JSONMetadataT) (*JSONMetadataT, error) {
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

func (p *localFileJSONProvider) IsContentUpdated() (bool, error) {
	meta, err := p.GetFreshMetadata()
	if err != nil {
		return false, errors.Wrap(err, "cannot if metadata cache is still valid")
	}

	isUpdated := (p.metadata.LastModified == meta.LastModified) && (p.metadata.Size == meta.Size)

	return isUpdated, nil
}

func (p *localFileJSONProvider) GetContent() (*UpdateJSON, *JSONMetadataT, error) {
	return p.GetFreshContent()
}
