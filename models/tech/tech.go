package tech

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Tech struct {
	Categories []Category `json:"categories,omitempty"`
	Name       string     `json:"name"`
	Version    string     `json:"version"`
	Confidence int        `json:"confidence"`
	Icon       string     `json:"icon,omitempty" description:"base64 image"`
	Url        string     `json:"url" description:"url to technology"`
}

type Status struct {
	Confirmed bool `json:"confirmed"`
	False     bool `json:"false"`
}

type TargetTech struct {
	Id         bson.ObjectId `json:"id,omitempty" bson:"_id"`
	Target     bson.ObjectId `json:"target"`
	Project    bson.ObjectId `json:"project"`
	Created    time.Time     `json:"created,omitempty" description:"when issue is created"`
	Updated    time.Time     `json:"updated,omitempty" description:"when issue is updated"`
	Activities []*Activity   `json:"activities,omitempty"`

	Tech   `json:",inline"`
	Status `json:",inline"`
}

type Report struct {
	Report  bson.ObjectId `json:"report"`
	Scan    bson.ObjectId `json:"scan,omitempty" description:"scan id"`
	Session bson.ObjectId `json:"session,omitempty" bson:"session" description:"scan session id"`
}

type Activity struct {
	Type    ActivityType `json:"type"`
	Created time.Time    `json:"created"`

	User   bson.ObjectId `json:"user,omitempty" description:"who did the activity"`
	Report *Report       `json:"report,omitempty" description:"link to report for reported activity"`
}
