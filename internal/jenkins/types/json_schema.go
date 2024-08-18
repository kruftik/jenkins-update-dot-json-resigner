package types

import (
	"encoding/json"
	"fmt"
	"io"

	cjson "github.com/gibson042/canonicaljson-go"
)

type Core struct {
	BuildDate string `json:"buildDate"`
	Name      string `json:"name"`
	Sha1      string `json:"sha1"`
	Sha256    string `json:"sha256"`
	Size      int64  `json:"size,omitempty"`
	URL       string `json:"url"`
	Version   string `json:"version"`
}

type Dependencies struct {
	Name     string `json:"name"`
	Optional bool   `json:"optional"`
	Version  string `json:"version"`
}

type Developers struct {
	DeveloperID string `json:"developerId,omitempty"`
	Email       string `json:"email,omitempty"`
	Name        string `json:"name,omitempty"`
}

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

type Plugins map[string]Plugin

type Signature struct {
	Certificates []string `json:"certificates"`

	CorrectDigest    string `json:"correct_digest"`
	CorrectDigest512 string `json:"correct_digest512"`

	CorrectSignature    string `json:"correct_signature"`
	CorrectSignature512 string `json:"correct_signature512"`
}

type SignedUpdateJSON struct {
	*InsecureUpdateJSON
	Signature Signature `json:"signature"`
}

func (o *SignedUpdateJSON) Sign(signer Signer) error {
	signature, err := signer.GetSignature(o.GetUnsigned())
	if err != nil {
		return fmt.Errorf("cannot calculate signature: %w", err)
	}

	if err := signer.VerifySignature(o.InsecureUpdateJSON, signature); err != nil {
		return fmt.Errorf("cannot verify signature: %w", err)
	}

	o.Signature = signature

	return nil
}

func (o *SignedUpdateJSON) MarshalJSON() ([]byte, error) {
	bytez, err := cjson.Marshal(*o)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal SignedUpdateJSON: %w", err)
	}

	return replaceSymbolsByTrickyMap(bytez), nil
}

func (o *SignedUpdateJSON) MarshalJSONTo(w io.Writer) error {
	return cjson.NewEncoder(w).Encode(o)
}

func (o *SignedUpdateJSON) GetUnsigned() *InsecureUpdateJSON {
	return o.InsecureUpdateJSON
}

type InsecureUpdateJSON struct {
	ConnectionCheckURL  string                 `json:"connectionCheckUrl"`
	Core                Core                   `json:"core"`
	Deprecations        map[string]interface{} `json:"deprecations"`
	GenerationTimestamp string                 `json:"generationTimestamp"`
	ID                  string                 `json:"id"`
	Plugins             Plugins                `json:"plugins"`
	UpdateCenterVersion string                 `json:"updateCenterVersion"`
	Warnings            []interface{}          `json:"warnings"`
}

func (o *InsecureUpdateJSON) MarshalJSONTo(w io.Writer) error {
	return cjson.NewEncoder(w).Encode(o)
}

func (o *InsecureUpdateJSON) MarshalJSON() ([]byte, error) {
	bytez, err := cjson.Marshal(*o)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal InsecureUpdateJSON: %w", err)
	}

	return replaceSymbolsByTrickyMap(bytez), nil
}
