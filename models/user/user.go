package user

import (
	"fmt"
	"github.com/bearded-web/bearded/pkg/pagination"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type User struct {
	Id       bson.ObjectId `json:"id" bson:"_id"`
	Email    string        `json:"email"`
	Password string        `json:"-"` // password hash in passlib format: $hashAlgo[$values]$hexdigest_hash$

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

func (u *User) String() string {
	return fmt.Sprintf("%x %s", u.Id, u.Email)
}

type UserList struct {
	pagination.Meta `json:",inline"`
	Results         []*User `json:"results"`
}
