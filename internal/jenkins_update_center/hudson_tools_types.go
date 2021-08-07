package jenkins_update_center

type HudsonToolUpdatesReleaseFile struct {
	Name  string `json:"name"`
	Title string `json:"title"`

	FilePath string `json:"filepath"`

	SHA256 string `json:"SHA256,omitempty"`
	MD5    string `json:"MD5,omitempty"`
}

type HudsonToolUpdatesRelease struct {
	Title string `json:"title"`

	LicPath  string `json:"licpath"`
	LicTitle string `json:"lictitle"`

	Path string `json:"path,omitempty"`

	Files []HudsonToolUpdatesReleaseFile `json:"files"`
}

type HudsonToolUpdatesData struct {
	Name string `json:"name"`

	Releases []HudsonToolUpdatesRelease `json:"releases"`
}

type HudsonToolUpdates struct {
	Data      []HudsonToolUpdatesData `json:"data"`
	Signature Signature               `json:"signature"`
	Version   int                     `json:"version"`
}
