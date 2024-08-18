package types

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
