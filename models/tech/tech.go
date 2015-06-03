package tech

import (
	"time"

	"github.com/bearded-web/bearded/pkg/pagination"
	"gopkg.in/mgo.v2/bson"
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
	Report      bson.ObjectId `json:"report"`
	Scan        bson.ObjectId `json:"scan,omitempty" description:"scan id"`
	ScanSession bson.ObjectId `json:"session,omitempty" bson:"session" description:"scan session id"`
}

type Activity struct {
	Type    ActivityType `json:"type"`
	Created time.Time    `json:"created"`

	User   bson.ObjectId `json:"user,omitempty" bson:",omitempty" description:"who did the activity"`
	Report *Report       `json:"report,omitempty" description:"link to report for reported activity"`
}

func (i *TargetTech) AddUserReportActivity(userId bson.ObjectId) {
	i.Activities = append(i.Activities, &Activity{
		Created: time.Now().UTC(),
		Type:    ActivityReported,
		User:    userId,
	})
}

func (i *TargetTech) AddReportActivity(reportId, scanId, sessionId bson.ObjectId) {
	i.Activities = append(i.Activities, &Activity{
		Created: time.Now().UTC(),
		Type:    ActivityReported,
		Report: &Report{
			Report:      reportId,
			Scan:        scanId,
			ScanSession: sessionId,
		},
	})
}

type TargetTechList struct {
	pagination.Meta `json:",inline"`
	Results         []*TargetTech `json:"results"`
}
