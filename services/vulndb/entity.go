package vulndb

import (
	"github.com/bearded-web/bearded/models/issue"
	"github.com/bearded-web/bearded/pkg/pagination"
)

type CompactVuln struct {
	Id       int            `json:"id"`
	Title    string         `json:"title"`
	Severity issue.Severity `json:"severity"`
}

type CompactVulnList struct {
	pagination.Meta `json:",inline"`
	Results         []*CompactVuln `json:"results"`
}
