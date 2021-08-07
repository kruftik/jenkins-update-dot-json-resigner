package jenkins_update_center

type HudsonTaskListElement struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type HudsonTaskUpdates struct {
	List      []HudsonTaskListElement `json:"list"`
	Signature Signature               `json:"signature"`
}
