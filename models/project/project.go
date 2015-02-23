package project

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/pkg/pagination"
)

type Member struct {
	User bson.ObjectId `json:"user"`
}

type Project struct {
	Id      bson.ObjectId `json:"id,omitempty" bson:"_id"`
	Name    string        `json:"name"`
	Owner   bson.ObjectId `json:"owner,omitempty"`
	Created time.Time     `json:"created,omitempty"`
	Updated time.Time     `json:"updated,omitempty"`

	Members []*Member `json:"members" bson:"members"`
}

type ProjectList struct {
	pagination.Meta `json:",inline"`
	Results         []*Project `json:"results"`
}

func (p *Project) GetMember(userId bson.ObjectId) *Member {
	for _, m := range p.Members {
		if m.User == userId {
			return m
		}
	}
	return nil
}

type MemberList struct {
	pagination.Meta `json:",inline"`
	Results         []*Member `json:"results"`
}
