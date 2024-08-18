package sourcefileproviders

import (
	"context"
	"io"
)

type Provider interface {
	GetBody(ctx context.Context) (FileMetadata, io.ReadCloser, error)
	GetMetadata(ctx context.Context) (FileMetadata, error)
}
