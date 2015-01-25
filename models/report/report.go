package report

import (
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"time"

	"github.com/bearded-web/bearded/pkg/pagination"
)

type ReportType string

const (
	TypeRaw ReportType = "raw"
)

// It's a hack to show custom type as string in swagger
func (t ReportType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

type Report struct {
	Id          bson.ObjectId `json:"id,omitempty" bson:"_id"`
	Type        ReportType    `json:"type" description:"one of [raw]"`
	Created     time.Time     `json:"created,omitempty" description:"when report is created"`
	Updated     time.Time     `json:"updated,omitempty" description:"when report is updated"`
	Scan        bson.ObjectId `json:"scan,omitempty" description:"scan id"`
	ScanSession bson.ObjectId `json:"scanSession,omitempty" bson:"scanSession" description:"scan session id"`

	Raw string `json:"raw"`
}

type ReportList struct {
	pagination.Meta `json:",inline"`
	Results         []*Report `json:"results"`
}
