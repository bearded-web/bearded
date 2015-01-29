package feed

import (
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/scan"
	"github.com/bearded-web/bearded/pkg/pagination"
)

type ItemType string

const (
	TypeComment ItemType = "comment"
	TypeScan    ItemType = "scan"
)

// It's a hack to show custom type as string in swagger
func (t ItemType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t ItemType) Enum() []interface{} {
	return []interface{}{TypeScan, TypeComment}
}

func (t ItemType) Convert(text string) (interface{}, error) {
	return ItemType(text), nil
}

type FeedItem struct {
	Id      bson.ObjectId `json:"id" bson:"_id"`
	Type    ItemType      `json:"type"`
	Created time.Time     `json:"created,omitempty" description:"when feed item is created"`
	Updated time.Time     `json:"updated,omitempty" description:"when feed item is updated"`

	Owner   bson.ObjectId `json:"owner" bson:"owner" description:""`
	Target  bson.ObjectId `json:"target" bson:"target" description:"target for this feed item"`
	Project bson.ObjectId `json:"project" bson:"project" description:"project for this feed item"`

	// data for scan types
	ScanId bson.ObjectId `json:"-" bson:"scanid,omitempty"`
	// scan field enriched from scanId
	Scan *scan.Scan `json:"scan,omitempty" bson:"-" description:"scan shows only for type: scan"`
}

type Feed struct {
	pagination.Meta `json:",inline"`
	Results         []*FeedItem `json:"results"`
}

func (p *FeedItem) String() string {
	var str string
	if p.Id != "" {
		str = fmt.Sprintf("%x - %s", string(p.Id), p.Type)
	} else {
		str = fmt.Sprintf("%s", p.Type)
	}
	return str
}
