package report

type Url struct {
	Url string `json:"url"`
}

type Extra struct {
	Url   string `json:"url" description:""`
	Title string `json:"title"`
}

type Issue struct {
	Severity Severity `json:"severity"`
	Summary  string   `json:"summary"`
	Desc     string   `json:"desc"`
	Urls     []*Url   `json:"urls,omitempty" description:"where this issue is happened"`
	Extras   []*Extra `json:"extras,omitempty" bson:"extras" description:"information about vulnerability"`
}
