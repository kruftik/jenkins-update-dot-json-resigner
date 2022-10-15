package hudson_tools

import (
	"jenkins-resigner-service/internal/types/common"
)

type UpdatesReleaseFile struct {
	Name  string `json:"name"`
	Title string `json:"title"`

	FilePath string `json:"filepath"`

	SHA256 string `json:"SHA256,omitempty"`
	MD5    string `json:"MD5,omitempty"`
}

type UpdatesRelease struct {
	Title string `json:"title"`

	LicPath  string `json:"licpath"`
	LicTitle string `json:"lictitle"`

	Path string `json:"path,omitempty"`

	Files []UpdatesReleaseFile `json:"files"`
}

type UpdatesData struct {
	Name string `json:"name"`

	Releases []UpdatesRelease `json:"releases"`
}

type Updates struct {
	Data      []UpdatesData      `json:"data"`
	Signature common.SignatureV1 `json:"signature"`
	Version   int                `json:"version"`
}
