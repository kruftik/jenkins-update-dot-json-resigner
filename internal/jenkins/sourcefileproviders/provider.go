package sourcefileproviders

import (
	"context"
	"io"
)

type SourceFileProvider interface {
	GetJSONPBody(ctx context.Context) (JSONFileMetadata, io.ReadCloser, error)
	GetJSONPMetadata(ctx context.Context) (JSONFileMetadata, error)
}

//type JSONProvider interface {
//	GetFreshContent() (*UpdateJSON, *JSONMetadataT, error)
//	GetFreshMetadata() (*JSONMetadataT, error)
//
//	RefreshMetadata(*JSONMetadataT) (*JSONMetadataT, error)
//
//	GetContent() (*UpdateJSON, *JSONMetadataT, error)
//
//	IsContentUpdated() (bool, error)
//}
