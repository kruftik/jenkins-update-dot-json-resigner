package jenkins

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/sourcefileproviders"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/types"
)

func (s *Service) getOriginal(ctx context.Context) (sourcefileproviders.FileMetadata, *types.SignedUpdateJSON, error) {
	ctx, cancel := context.WithTimeout(ctx, s.cfg.GetUpdateJSONBodyTimeout)
	defer cancel()

	metadata, r, err := s.sourceFileProvider.GetBody(ctx)
	if err != nil {
		return sourcefileproviders.FileMetadata{}, nil, fmt.Errorf("cannot get source file: %w", err)
	}
	defer r.Close()

	signedJSON := &types.SignedUpdateJSON{}

	if err := json.NewDecoder(r).Decode(signedJSON); err != nil {
		return sourcefileproviders.FileMetadata{}, nil, fmt.Errorf("cannot unmarshal json: %w", err)
	}

	return metadata, signedJSON, nil
}
