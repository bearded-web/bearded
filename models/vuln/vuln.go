package vuln

import (
	"github.com/bearded-web/bearded/models/issue"
	"github.com/bearded-web/bearded/pkg/pagination"
)

type Reference struct {
	Url   string `json:"url,omitempty"`
	Title string `json:"title"`
}

type VulnFix struct {
	Guidance string `json:"guidance"`
	Effort   int    `json:"effort"`
}

type Vuln struct {
	Id          int              `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Severity    issue.Severity   `json:"severity"`
	Tags        []string         `json:"tags,omitempty"`
	References  []Reference      `json:"references"`
	Wasc        []string         `json:"wasc,omitempty"`
	Cwe         []string         `json:"cwe,omitempty"`
	OwaspTop10  map[string][]int `json:"owasp_top_10,omitempty"`
	Fix         VulnFix          `json:"fix"`
}

func (v *Vuln) Compact() *CompactVuln {
	return &CompactVuln{
		Id:       v.Id,
		Title:    v.Title,
		Severity: v.Severity,
	}
}

type VulnList struct {
	pagination.Meta `json:",inline"`
	Results         []*Vuln `json:"results"`
}

type CompactVuln struct {
	Id       int            `json:"id"`
	Title    string         `json:"title"`
	Severity issue.Severity `json:"severity"`
}

type CompactVulnList struct {
	pagination.Meta `json:",inline"`
	Results         []*CompactVuln `json:"results"`
}
