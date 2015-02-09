package scan

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/plan"
	"github.com/bearded-web/bearded/pkg/pagination"
)

// useful time points
type Dates struct {
	Created  *time.Time `json:"created,omitempty"`
	Updated  *time.Time `json:"updated,omitempty"`
	Queued   *time.Time `json:"queued,omitempty"`
	Started  *time.Time `json:"started,omitempty"`
	Finished *time.Time `json:"finished,omitempty"`
}

type Session struct {
	Id     bson.ObjectId      `json:"id,omitempty"`
	Status ScanStatus         `json:"status" description:"one of [created|queued|working|paused|finished|failed]"`
	Step   *plan.WorkflowStep `json:"step"`
	Plugin bson.ObjectId      `json:"plugin,omitempty" description:"plugin id"`
	Scan   bson.ObjectId      `json:"scan" description:"scan id"`
	// dates
	Dates `json:",inline"`

	// Children can be created by plugins
	Children []*Session    `json:"children,omitempty" description:"children can be created by scripts" bson:"children,omitempty"`
	Parent   bson.ObjectId `json:"parent,omitempty" description:"parent session for this one" bson:"parent,omitempty"`
}

func (p *Session) GetChild(id bson.ObjectId) *Session {
	for _, sess := range p.Children {
		if sess.Id == id {
			return sess
		}
		if sess.Children != nil && len(sess.Children) > 0 {
			child := sess.GetChild(id)
			if child != nil {
				return child
			}
		}
	}
	return nil
}

func (p *Session) HasParent() bool {
	return p.Parent != ""
}

type ScanConf struct {
	Target string                 `json:"target"`
	Params map[string]interface{} `json:"params"`
}

type Scan struct {
	Id     bson.ObjectId `json:"id,omitempty" bson:"_id"`
	Status ScanStatus    `json:"status,omitempty" description:"one of [created|queued|working|pause|finished|failed]"`
	Conf   ScanConf      `json:"conf,omitempty"`

	//	Report    *report.Report `json:"report,omitempty" form:"-"`
	Sessions []*Session `json:"sessions,omitempty"`

	Plan    bson.ObjectId `json:"plan"`
	Owner   bson.ObjectId `json:"owner,omitempty"`
	Target  bson.ObjectId `json:"target"`
	Project bson.ObjectId `json:"project"`

	// dates
	Dates `json:",inline"`
}

type ScanList struct {
	pagination.Meta `json:",inline"`
	Results         []*Scan `json:"results"`
}

func (p *Scan) String() string {
	return fmt.Sprintf("%x [%s]", string(p.Id), p.Status)
}

func (p *Scan) GetSession(id bson.ObjectId) *Session {
	for _, sess := range p.Sessions {
		if sess.Id == id {
			return sess
		}
		if sess.Children != nil && len(sess.Children) > 0 {
			child := sess.GetChild(id)
			if child != nil {
				return child
			}
		}
	}
	return nil
}
