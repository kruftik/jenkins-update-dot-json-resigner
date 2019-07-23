package jenkins_update_center

import (
	"jenkins-resigner-service/jenkins_update_center/json_schema"
	"time"
)

type JSONMetadataT struct {
	LastModified time.Time
	Size         int64
}

type JSONProvider interface {
	init(string) error

	GetContent() (*json_schema.UpdateJSON, error)
	GetMetadata() (*JSONMetadataT, error)

	IsContentUpdated() (bool, error)
}
