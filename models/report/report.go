package report

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/file"
	"github.com/bearded-web/bearded/models/issue"
	"github.com/bearded-web/bearded/models/tech"
	"github.com/bearded-web/bearded/pkg/pagination"
)

type Raw struct {
	Raw   string       `json:"raw"`
	Files []*file.Meta `json:"files,omitempty" bson:"files,omitempty"`
}

type Report struct {
	Id          bson.ObjectId `json:"id,omitempty" bson:"_id"`
	Type        ReportType    `json:"type" description:"one of [raw,issues,techs,multi,empty]"`
	Created     time.Time     `json:"created,omitempty" description:"when report is created"`
	Updated     time.Time     `json:"updated,omitempty" description:"when report is updated"`
	Scan        bson.ObjectId `json:"scan,omitempty" description:"scan id"`
	ScanSession bson.ObjectId `json:"scanSession,omitempty" bson:"scanSession" description:"scan session id"`

	Raw `json:",inline,omitempty" bson:"raw,inline"`

	Multi  []*Report      `json:"multi,omitempty" bson:"multi,omitempty"`
	Issues []*issue.Issue `json:"issues,omitempty" bson:"issues,omitempty"`
	Techs  []*tech.Tech   `json:"techs,omitempty"`
}

type ReportList struct {
	pagination.Meta `json:",inline"`
	Results         []*Report `json:"results"`
}

// set scan to report and all underlying multi reports if they are existed
func (r *Report) SetScan(scanId bson.ObjectId) {
	r.Scan = scanId
	if r.Type == TypeMulti {
		for _, rep := range r.Multi {
			rep.SetScan(scanId)
		}
	}
}

func (r *Report) SetScanSession(sessionId bson.ObjectId) {
	r.ScanSession = sessionId
	if r.Type == TypeMulti {
		for _, rep := range r.Multi {
			rep.SetScanSession(sessionId)
		}
	}
}
