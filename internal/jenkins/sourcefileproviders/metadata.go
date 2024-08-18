package sourcefileproviders

import (
	"time"
)

type JSONFileMetadata struct {
	LastModified time.Time
	Size         int64
	Etag         string
}
