package jenkins_update_center

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
	GetFreshContent() (*UpdateJSON, *JSONMetadataT, error)
	GetFreshMetadata() (*JSONMetadataT, error)

	RefreshMetadata(*JSONMetadataT) (*JSONMetadataT, error)

	GetContent() (*UpdateJSON, *JSONMetadataT, error)

	IsContentUpdated() (bool, error)
}
