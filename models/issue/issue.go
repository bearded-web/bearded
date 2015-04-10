package issue

import (
	"github.com/bearded-web/bearded/pkg/pagination"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Extra struct {
	Url   string `json:"url" description:""`
	Title string `json:"title"`
}

type Reference struct {
	Url   string `json:"url"`
	Title string `json:"title"`
}

type Issue struct {
	UniqId     string       `json:"uniqId,omitempty" bson:"uniqId" description:"id for merging similar issues"`
	Summary    string       `json:"summary"`
	VulnType   int          `json:"vulnType,omitempty" bson:"vulnType" description:"vulnerability type from vulndb"`
	Severity   Severity     `json:"severity"`
	References []*Reference `json:"references,omitempty" bson:"references" description:"information about vulnerability"`
	Extras     []*Extra     `json:"extras,omitempty" bson:"extras" description:"information about vulnerability, deprecated"`
	Desc       string       `json:"desc,omitempty"`
	Vector     *Vector      `json:"vector,omitempty"`
	//	Affect   Affect   `json:"affect,omitempty" description:"who is affected by the issue?"`
}

type Status struct {
	Confirmed bool `json:"confirmed" description:"the issue was confirmed by someone"`
	False     bool `json:"false"`
	Muted     bool `json:"muted"`
	Resolved  bool `json:"resolved"`
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

type TargetIssue struct {
	Id         bson.ObjectId `json:"id,omitempty" bson:"_id"`
	Target     bson.ObjectId `json:"target"`
	Project    bson.ObjectId `json:"project"`
	Created    time.Time     `json:"created,omitempty" description:"when issue is created"`
	Updated    time.Time     `json:"updated,omitempty" description:"when issue is updated"`
	Activities []*Activity   `json:"activities,omitempty"`

	// usually this field is taken from the last report
	Issue  `json:",inline"`
	Status `json:",inline"`
}

type TargetIssueList struct {
	pagination.Meta `json:",inline"`
	Results         []*TargetIssue `json:"results"`
}
