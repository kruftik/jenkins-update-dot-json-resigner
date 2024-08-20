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
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/json"
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
	log *zap.SugaredLogger,
	cfg config.AppConfig,
	sourceFileProvider sourcefileproviders.Provider,
	signer types.Signer,
	patchers []types.Patcher,
) *Service {
	s := &Service{
		log:                log,
		cfg:                cfg,
		sourceFileProvider: sourceFileProvider,
		signer:             signer,
		patchers:           patchers,
	}

	return s
}

func (s *Service) CleanUp(_ context.Context) error {
	if err := os.Remove(path.Join(s.cfg.DataDirPath, UpdateCenterDotJSON)); err != nil {
		return fmt.Errorf("cannot remove jsonp file")
	}
	if err := os.Remove(path.Join(s.cfg.DataDirPath, UpdateCenterDotHTML)); err != nil {
		return fmt.Errorf("cannot remove html file")
	}

	s.log.Debugf("patched files removed")

	return nil
}

func (s *Service) writeDataWithTrailers(w io.Writer, data io.Reader, prefix, suffix []byte) error {
	r := io.MultiReader(
		bytes.NewReader(prefix),
		data,
		bytes.NewReader(suffix),
	)

	if _, err := io.Copy(w, r); err != nil {
		return fmt.Errorf("cannot write to file: %w", err)
	}

	return nil
}

func (s *Service) RefreshContent(ctx context.Context) error {
	newMetadata, err := s.sourceFileProvider.GetMetadata(ctx)
	if err != nil {
		return fmt.Errorf("failed to get JSONP metadata: %w", err)
	}

	jsonpFile := path.Join(s.cfg.DataDirPath, UpdateCenterDotJSON)
	htmlFile := path.Join(s.cfg.DataDirPath, UpdateCenterDotHTML)

	_, err1 := os.Stat(jsonpFile)
	_, err2 := os.Stat(htmlFile)

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

	if err := s.signer.VerifySignature(signedJSON.GetUnsigned(), signedJSON.Signature); err != nil {
		return fmt.Errorf("cannot verify original file signature: %w", err)
	}

	if err := s.patchAndSign(signedJSON); err != nil {
		return fmt.Errorf("cannot patch and sign file: %w", err)
	}

	bytez, err := json.MarshalJSON(signedJSON)
	if err != nil {
		return fmt.Errorf("failed to write patched content to buffer: %w", err)
	}

	f, err := os.Create(jsonpFile)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", jsonpFile, err)
	}
	defer f.Close()

	if err := s.writeDataWithTrailers(f, bytes.NewReader(bytez), sourcefileproviders.WrappedJSONPPrefix, sourcefileproviders.WrappedJSONPSuffix); err != nil {
		return fmt.Errorf("cannot write %s: %w", jsonpFile, err)
	}

	s.log.Debugf("%s file saved", jsonpFile)

	f, err = os.Create(htmlFile)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", htmlFile, err)
	}
	defer f.Close()

	if err := s.writeDataWithTrailers(f, bytes.NewReader(bytez), sourcefileproviders.WrappedHTMLPrefix, sourcefileproviders.WrappedHTMLSuffix); err != nil {
		return fmt.Errorf("cannot write %s: %w", htmlFile, err)
	}

	s.log.Debugf("%s file saved", htmlFile)

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
