package comment

import (
	"time"

	"github.com/bearded-web/bearded/pkg/pagination"
	"gopkg.in/mgo.v2/bson"
)

type Comment struct {
	Id      bson.ObjectId `json:"id" bson:"_id"`
	Created time.Time     `json:"created,omitempty" description:"when item is created"`
	Updated time.Time     `json:"updated,omitempty" description:"when item is updated"`
	Owner   bson.ObjectId `json:"owner" bson:"owner" description:"user who created a comment"`
	Text    string        `json:"text" description:"raw markdown text"`

	Type Type          `json:"-"`
	Link bson.ObjectId `json:"-"`
}

type CommentList struct {
	pagination.Meta `json:",inline"`
	Results         []*Comment `json:"results"`
}
