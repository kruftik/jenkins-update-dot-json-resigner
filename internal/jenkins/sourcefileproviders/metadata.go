package sourcefileproviders

import (
	"time"
)

type FileMetadata struct {
	LastModified time.Time
	Size         int64
	Etag         string
}
