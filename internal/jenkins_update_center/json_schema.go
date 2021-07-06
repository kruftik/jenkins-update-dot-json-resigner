package jenkins_update_center

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

type Plugin struct {
	BuildDate              string               `json:"buildDate"`
	DefaultBranch          string               `json:"defaultBranch,omitempty"`
	CompatibleSinceVersion string               `json:"compatibleSinceVersion,omitempty"`
	Dependencies           []Dependencies       `json:"dependencies"`
	Developers             []Developers         `json:"developers"`
	Excerpt                string               `json:"excerpt"`
	Gav                    string               `json:"gav"`
	IssueTrackers          []PluginIssueTracker `json:"issueTrackers,omitempty"`
	Labels                 []string             `json:"labels"`
	MinimumJavaVersion     string               `json:"minimumJavaVersion,omitempty"`
	Name                   string               `json:"name"`
	Popularity             int                  `json:"popularity"`
	PreviousTimestamp      string               `json:"previousTimestamp,omitempty"`
	PreviousVersion        string               `json:"previousVersion,omitempty"`
	ReleaseTimestamp       string               `json:"releaseTimestamp,omitempty"`
	RequiredCore           string               `json:"requiredCore"`
	Scm                    string               `json:"scm,omitempty"`
	SHA1                   string               `json:"sha1"`
	SHA256                 string               `json:"sha256"`
	Size                   int64                `json:"size"`
	Title                  string               `json:"title"`
	URL                    string               `json:"url"`
	Version                string               `json:"version"`
	Wiki                   string               `json:"wiki"`
}

type Plugins map[string]Plugin

type Signature struct {
	Certificates []string `json:"certificates"`

	CorrectDigest    string `json:"correct_digest"`
	CorrectDigest512 string `json:"correct_digest512"`

	CorrectSignature    string `json:"correct_signature"`
	CorrectSignature512 string `json:"correct_signature512"`

	// incorrect digest and signatures are not included anymore
	//Digest              string   `json:"digest"`
	//Digest512           string   `json:"digest512"`
	//Signature           string   `json:"signature"`
	//Signature512        string   `json:"signature512"`
}

//type Versions struct {
//	LastVersion string `json:"lastVersion"`
//	Pattern     string `json:"pattern"`
//}

//type Warnings struct {
//	ID       string     `json:"id"`
//	Message  string     `json:"message"`
//	Name     string     `json:"name"`
//	Type     string     `json:"type"`
//	URL      string     `json:"url"`
//	Versions []Versions `json:"versions"`
//}

type UpdateJSON struct {
	ConnectionCheckURL  string                 `json:"connectionCheckUrl"`
	Core                Core                   `json:"core"`
	Deprecations        map[string]interface{} `json:"deprecations"`
	GenerationTimestamp string                 `json:"generationTimestamp"`
	ID                  string                 `json:"id"`
	Plugins             Plugins                `json:"plugins"`
	Signature           Signature              `json:"signature"`
	UpdateCenterVersion string                 `json:"updateCenterVersion"`
	Warnings            []interface{}          `json:"warnings"`
}

type InsecureUpdateJSON struct {
	ConnectionCheckURL  string                 `json:"connectionCheckUrl"`
	Core                Core                   `json:"core"`
	Deprecations        map[string]interface{} `json:"deprecations"`
	GenerationTimestamp string                 `json:"generationTimestamp"`
	ID                  string                 `json:"id"`
	Plugins             Plugins                `json:"plugins"`
	Signature           Signature              `json:"-"`
	UpdateCenterVersion string                 `json:"updateCenterVersion"`
	Warnings            []interface{}          `json:"warnings"`
}
