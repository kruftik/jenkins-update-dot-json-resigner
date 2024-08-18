package jenkins

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/sourcefileproviders"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/types"
)

func (s *Service) getOriginal(ctx context.Context) (sourcefileproviders.JSONFileMetadata, *types.SignedUpdateJSON, error) {
	ctx, cancel := context.WithTimeout(ctx, s.cfg.UpdateJSONDownloadTimeout)
	defer cancel()

	metadata, r, err := s.sourceFileProvider.GetJSONPBody(ctx)
	if err != nil {
		return sourcefileproviders.JSONFileMetadata{}, nil, fmt.Errorf("cannot get source file: %w", err)
	}
	defer r.Close()

	signedJSON := &types.SignedUpdateJSON{}

	if err := json.NewDecoder(r).Decode(signedJSON); err != nil {
		return sourcefileproviders.JSONFileMetadata{}, nil, fmt.Errorf("cannot unmarshal json: %w", err)
	}

	if err := s.signer.VerifySignature(signedJSON.GetUnsigned(), signedJSON.Signature); err != nil {
		return sourcefileproviders.JSONFileMetadata{}, nil, fmt.Errorf("cannot verify original file signature: %w", err)
	}

	s.log.Debug("original file signature verified")

	return metadata, signedJSON, nil
}
