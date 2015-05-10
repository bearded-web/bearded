package token

import (
	"time"

	"github.com/bearded-web/bearded/pkg/pagination"
	"gopkg.in/mgo.v2/bson"
)

const TokenLength = 32
const DefaultScope = "all"

type Token struct {
	Id        bson.ObjectId `json:"id,omitempty" bson:"_id"`
	User      bson.ObjectId `json:"user"`
	Name      string        `json:"name,omitempty" description:"token name"`
	Hash      string        `json:"-"`
	HashValue string        `json:"value" bson:"-"`
	Scopes    []string      `json:"scopes,omitempty"`
	Created   time.Time     `json:"created,omitempty"`
	Updated   time.Time     `json:"updated,omitempty"`
	Removed   bool          `json:"-"`
}

type TokenList struct {
	pagination.Meta `json:",inline"`
	Results         []*Token `json:"results"`
}
