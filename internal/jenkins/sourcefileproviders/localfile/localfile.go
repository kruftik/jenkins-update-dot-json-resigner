package localfile

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/sourcefileproviders"
)

var (
	_ sourcefileproviders.Provider = Provider{}
)

type Provider struct {
	path string
}

func NewLocalFileProvider(path string) (*Provider, error) {
	p := &Provider{
		path: path,
	}

	if err := p.validate(path); err != nil {
		return nil, fmt.Errorf("invalid local file %s: %w", path, err)
	}

	return p, nil
}

func (p Provider) validate(src string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return fmt.Errorf("cannot close file: %w", err)
	}

	return nil
}

func (p Provider) getMetadata(fi os.FileInfo) sourcefileproviders.FileMetadata {
	return sourcefileproviders.FileMetadata{
		LastModified: fi.ModTime(),
		Size:         fi.Size(),
	}
}

func (p Provider) GetMetadata(_ context.Context) (sourcefileproviders.FileMetadata, error) {
	fi, err := os.Stat(p.path)
	if err != nil {
		return sourcefileproviders.FileMetadata{}, err
	}

	return p.getMetadata(fi), nil
}

func (p Provider) GetBody(_ context.Context) (sourcefileproviders.FileMetadata, io.ReadCloser, error) {
	f, err := os.Open(p.path)
	if err != nil {
		return sourcefileproviders.FileMetadata{}, nil, fmt.Errorf("cannot open file: %w", err)
	}

	fi, err := f.Stat()
	if err != nil {
		return sourcefileproviders.FileMetadata{}, nil, fmt.Errorf("cannot stat file: %w", err)
	}

	r, err := sourcefileproviders.NewJSONPTrailersStrippingReader(f, fi.Size())
	if err != nil {
		return sourcefileproviders.FileMetadata{}, nil, err
	}

	return p.getMetadata(fi), r, nil
}

//func (p Provider) GetFreshContent() (*UpdateJSON, *JSONMetadataT, error) {
//	log.Infof("Reading %s...", p.path)
//
//	sbytes, err := ioutil.ReadFile(p.path)
//	if err != nil {
//		return nil, nil, errors.Wrapf(err, "cannot read update.json content: %s", err)
//	}
//
//	uj, err := prepareUpdateJSONObject(sbytes)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	if err = uj.VerifySignature(); err != nil {
//		return nil, nil, err
//	}
//
//	meta, err := p.GetFreshMetadata()
//
//	return uj, meta, nil
//}
//
////func (p Provider) getMetadata() (*JSONMetadataT, error) {
////	return p.metadata, nil
////}
//
//
//func (p *Provider) RefreshMetadata(meta *JSONMetadataT) (*JSONMetadataT, error) {
//	var err error
//
//	if meta == nil {
//		meta, err = p.GetFreshMetadata()
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	p.metadata = meta
//
//	return meta, nil
//}
//
//func (p *Provider) IsContentUpdated() (bool, error) {
//	meta, err := p.GetFreshMetadata()
//	if err != nil {
//		return false, errors.Wrap(err, "cannot if metadata cache is still valid")
//	}
//
//	isUpdated := (p.metadata.LastModified == meta.LastModified) && (p.metadata.Size == meta.Size)
//
//	return isUpdated, nil
//}
//
//func (p *Provider) GetContent() (*UpdateJSON, *JSONMetadataT, error) {
//	return p.GetFreshContent()
//}