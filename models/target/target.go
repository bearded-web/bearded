package target

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/file"
	"github.com/bearded-web/bearded/models/issue"
	"github.com/bearded-web/bearded/pkg/pagination"
)

type SummaryReport struct {
	Issues map[issue.Severity]int `json:"issues"`
}

type Target struct {
	Id      bson.ObjectId  `json:"id,omitempty" bson:"_id"`
	Type    TargetType     `json:"type" description:"one of [web|android]"`
	Web     *WebTarget     `json:"web,omitempty" description:"information about web target"`
	Android *AndroidTarget `json:"android,omitempty" description:"information about android target"`
	Project bson.ObjectId  `json:"project"`
	Created time.Time      `json:"created,omitempty"`
	Updated time.Time      `json:"updated,omitempty"`

	SummaryReport *SummaryReport `json:"summaryReport,omitempty" bson:"summaryReport"`
}

type WebTarget struct {
	Domain string `json:"domain"`
}

type MobileTarget struct {
	Name string     `json:"name" description:"target name, 80 symbols max" mobile:"nonzero,max=80"`
	File *file.Meta `json:"file" description:"apk file metadata"`
}

type AndroidTarget struct {
	MobileTarget `json:",inline" bson:",inline"`
}

type TargetList struct {
	pagination.Meta `json:",inline"`
	Results         []*Target `json:"results"`
}

func (t *Target) Addr() string {
	if t.Type == "web" {
		return t.Web.Domain
	}
	return ""
}
