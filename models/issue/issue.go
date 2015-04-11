package issue

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/bearded-web/bearded/pkg/pagination"
	"gopkg.in/mgo.v2/bson"
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

func (i *Issue) GenerateUniqId() string {
	fields := []string{}
	fields = append(fields, i.Summary)
	fields = append(fields, fmt.Sprintf("%d", i.VulnType))
	fields = append(fields, i.Desc)

	if i.Vector != nil {
		fields = append(fields, i.Vector.Url)
		for _, transaction := range i.Vector.HttpTransactions {
			fields = append(
				fields,
				transaction.Method,
				transaction.Url,
				fmt.Sprintf("%#v", transaction.Params),
				fmt.Sprintf("%#v", transaction.Request),
			)
		}
	}
	hash := md5.New()
	hash.Write([]byte(strings.Join(fields, ":")))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

type Status struct {
	Confirmed bool `json:"confirmed" description:"the issue was confirmed by someone"`
	False     bool `json:"false"`
	Muted     bool `json:"muted"`
	Resolved  bool `json:"resolved"`
}

type Report struct {
	Report      bson.ObjectId `json:"report" description:"report id"`
	Scan        bson.ObjectId `json:"scan,omitempty" description:"scan id"`
	ScanSession bson.ObjectId `json:"scanSession,omitempty" bson:"scanSession" description:"scan session id"`
}

type Activity struct {
	Type    ActivityType `json:"type"`
	Created time.Time    `json:"created"`

	User   bson.ObjectId `json:"user,omitempty" bson:",omitempty" description:"who did the activity"`
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
	Issue  `json:",inline" bson:",inline"`
	Status `json:",inline" bson:",inline"`
}

func (i *TargetIssue) AddReportActivity(reportId, scanId, sessionId bson.ObjectId) {
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

type TargetIssueList struct {
	pagination.Meta `json:",inline"`
	Results         []*TargetIssue `json:"results"`
}
