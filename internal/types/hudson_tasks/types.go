package hudson_tasks

import (
	"jenkins-resigner-service/internal/types/common"
)

type ListElement struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Updates struct {
	List      []ListElement      `json:"list"`
	Signature common.SignatureV1 `json:"signature"`
}
