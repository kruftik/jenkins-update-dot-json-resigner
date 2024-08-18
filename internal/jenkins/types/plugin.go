package types

import (
	"encoding/json"
)

type Plugins map[string]Plugin

type PluginIssueTracker struct {
	ReportUrl string `json:"reportUrl,omitempty"`
	Type      string `json:"type,omitempty"`
	ViewUrl   string `json:"viewUrl,omitempty"`
}

type PluginIssueTrackersList struct {
	list []PluginIssueTracker
}

func (l *PluginIssueTrackersList) UnmarshalJSON(data []byte) error {
	l.list = make([]PluginIssueTracker, 0)
	return json.Unmarshal(data, &l.list)
}

func (l *PluginIssueTrackersList) MarshalJSON() ([]byte, error) {
	if len(l.list) == 0 {
		return []byte("[]"), nil
	}

	return json.Marshal(l.list)
}

type Plugin struct {
	BuildDate              string                   `json:"buildDate"`
	DefaultBranch          string                   `json:"defaultBranch,omitempty"`
	CompatibleSinceVersion string                   `json:"compatibleSinceVersion,omitempty"`
	Dependencies           []Dependencies           `json:"dependencies"`
	Developers             []Developers             `json:"developers"`
	Excerpt                string                   `json:"excerpt"`
	Gav                    string                   `json:"gav"`
	IssueTrackers          *PluginIssueTrackersList `json:"issueTrackers,omitempty"`
	Labels                 []string                 `json:"labels"`
	//MinimumJavaVersion     string               `json:"minimumJavaVersion,omitempty"`
	Name              string `json:"name"`
	Popularity        int    `json:"popularity"`
	PreviousTimestamp string `json:"previousTimestamp,omitempty"`
	PreviousVersion   string `json:"previousVersion,omitempty"`
	ReleaseTimestamp  string `json:"releaseTimestamp,omitempty"`
	RequiredCore      string `json:"requiredCore"`
	Scm               string `json:"scm,omitempty"`
	SHA1              string `json:"sha1"`
	SHA256            string `json:"sha256"`
	Size              int64  `json:"size"`
	Title             string `json:"title"`
	URL               string `json:"url"`
	Version           string `json:"version"`
	Wiki              string `json:"wiki"`
}
