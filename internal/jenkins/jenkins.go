package jenkins

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	"go.uber.org/zap"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/config"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/sourcefileproviders"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/types"
)

type PatchedFileRefresher interface {
	RefreshContent(ctx context.Context) error
}

var (
	_ PatchedFileRefresher = (*Service)(nil)
)

type Service struct {
	log                *zap.SugaredLogger
	cfg                config.AppConfig
	sourceFileProvider sourcefileproviders.Provider
	signer             types.Signer
	patchers           []types.Patcher

	mu sync.Mutex

	metadata sourcefileproviders.FileMetadata
}

func NewJenkinsUpdateCenter(
	ctx context.Context,
	log *zap.SugaredLogger,
	cfg config.AppConfig,
	sourceFileProvider sourcefileproviders.Provider,
	signer types.Signer,
	patchers []types.Patcher,
) (*Service, error) {
	s := &Service{
		log:                log,
		cfg:                cfg,
		sourceFileProvider: sourceFileProvider,
		signer:             signer,
		patchers:           patchers,
	}

	if err := s.RefreshContent(ctx); err != nil {
		return nil, fmt.Errorf("cannot refresh content: %w", err)
	}

	return s, nil
}

func (s *Service) CleanUp(_ context.Context) error {
	return nil
}

func (s *Service) writeDataToFileWithTrailers(where string, data, prefix, suffix []byte) error {
	f, err := os.Create(where)
	if err != nil {
		return fmt.Errorf("cannot create file: %w", err)
	}
	defer f.Close()

	r := io.MultiReader(
		bytes.NewReader(prefix),
		bytes.NewReader(data),
		bytes.NewReader(suffix),
	)

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("cannot write to file: %w", err)
	}

	return nil
}

func (s *Service) RefreshContent(ctx context.Context) error {
	newMetadata, err := s.sourceFileProvider.GetMetadata(ctx)
	if err != nil {
		return fmt.Errorf("failed to get JSONP metadata: %w", err)
	}

	_, err1 := os.Stat(path.Join(s.cfg.DataDirPath, UpdateCenterDotJSON))
	_, err2 := os.Stat(path.Join(s.cfg.DataDirPath, UpdateCenterDotHTML))

	if err1 != nil || err2 != nil {
		s.log.Info("temp file(s) do not exist, force update")
	}

	if newMetadata == s.metadata && err1 == nil && err2 == nil {
		s.log.Debugf("original file didn't change: %d bytes, last-modified: %s", newMetadata.Size, newMetadata.LastModified)
		return nil
	}

	s.log.Infof("original file changed: %d bytes, last-modified: %s", newMetadata.Size, newMetadata.LastModified)

	s.mu.Lock()
	defer s.mu.Unlock()

	newMetadata, signedJSON, err := s.getOriginal(ctx)
	if err != nil {
		return err
	}

	if err := s.patchAndSign(signedJSON); err != nil {
		return fmt.Errorf("cannot patch and sign file: %w", err)
	}

	bytez, err := signedJSON.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to write patched content to buffer: %w", err)
	}

	if err := s.writeDataToFileWithTrailers(path.Join(s.cfg.DataDirPath, UpdateCenterDotJSON), bytez, sourcefileproviders.WrappedJSONPPrefix, sourcefileproviders.WrappedJSONPSuffix); err != nil {
		return fmt.Errorf("cannot write update-center.json: %w", err)
	}

	if err := s.writeDataToFileWithTrailers(path.Join(s.cfg.DataDirPath, UpdateCenterDotHTML), bytez, sourcefileproviders.WrappedHTMLPrefix, sourcefileproviders.WrappedHTMLSuffix); err != nil {
		return fmt.Errorf("cannot write update-center.json.html: %w", err)
	}

	s.metadata = newMetadata

	return nil
}

func (s *Service) patchAndSign(signedJSON *types.SignedUpdateJSON) error {
	for _, patcher := range s.patchers {
		if err := patcher.Patch(signedJSON.GetUnsigned()); err != nil {
			return fmt.Errorf("cannot patch original file: %w", err)
		}
	}

	if err := signedJSON.Sign(s.signer); err != nil {
		return fmt.Errorf("cannot attach new signature: %w", err)
	}

	return nil
}
