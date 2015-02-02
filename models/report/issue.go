package report

type Url struct {
	Url string `json:"url"`
}

type Extra struct {
	Url   string `json:"url"`
	Title string `json:"title"`
}

type Issue struct {
	Severity Severity `json:"severity"`
	Summary  string   `json:"summary"`
	Desc     string   `json:"desc"`
	Urls     []*Url   `json:"urls,omitempty"`                 // where is issue is happens
	Extras   []*Extra `json:"extras,omitempty" bson:"extras"` // information about vulnerability
}
