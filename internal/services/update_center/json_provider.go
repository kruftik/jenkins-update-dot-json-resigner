package update_center

import (
	"time"
)

var (
	wrappedJSONPrefix  = []byte("updateCenter.post(\n")
	wrappedJSONPostfix = []byte("\n);")
)

type JSONMetadataT struct {
	LastModified time.Time
	Size         int64
	etag         string
}

type JSONProvider interface {
	GetFreshContent() (*SignedUpdatedJSON, *JSONMetadataT, error)
	GetFreshMetadata() (*JSONMetadataT, error)

	RefreshMetadata(*JSONMetadataT) (*JSONMetadataT, error)

	GetContent() (*SignedUpdatedJSON, *JSONMetadataT, error)

	IsContentUpdated() (bool, error)
}
