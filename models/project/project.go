package project

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/pkg/pagination"
)

type Project struct {
	Id      bson.ObjectId `json:"id,omitempty" bson:"_id"`
	Name    string        `json:"name"`
	Owner   bson.ObjectId `json:"owner,omitempty"`
	Created time.Time     `json:"created,omitempty"`
	Updated time.Time     `json:"updated,omitempty"`
}

type ProjectList struct {
	pagination.Meta `json:",inline"`
	Results         []*Project `json:"results"`
}
