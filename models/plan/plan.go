package plan

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/target"
	"github.com/bearded-web/bearded/pkg/pagination"
)

type File struct {
	Path string `json:"path" description:"absolute path to file in container"`
	Name string `json:"name" description:"rename file to name or use path if empty"`
}

type SharedFile struct {
	Path string `json:"path" description:"path is relative to /share/"`
	Text string `json:"text" description:"text is a content of text file"`
}

type Conf struct {
	CommandArgs string  `json:"commandArgs,omitempty" description:"passed to command line for plugins with type:util"`
	Target      string  `json:"target" description:"used in script, taken from scan conf directly"`
	TakeFiles   []*File `json:"takeFiles" description:"copy this files from container when it done"`

	SharedFiles []*SharedFile `json:"sharedFiles" description:"share file to container`
}

type WorkflowStep struct {
	Plugin string `json:"plugin" description:"plugin name"`
	Name   string `json:"name" description:"step name"`
	Desc   string `json:"desc,omitempty" description:"step description"`
	Conf   *Conf  `json:"conf,omitempty"`
}

type Plan struct {
	Id         bson.ObjectId     `json:"id,omitempty" bson:"_id,omitempty"`
	Name       string            `json:"name"`
	Desc       string            `json:"desc"` // human readable description
	Workflow   []*WorkflowStep   `json:"workflow"`
	Created    time.Time         `json:"created,omitempty" description:"when plan is created"`
	Updated    time.Time         `json:"updated,omitempty" description:"when plan is updated"`
	TargetType target.TargetType `json:"targetType" bson:"targetType" description:"what target type is supported"`
}

type PlanList struct {
	pagination.Meta `json:",inline"`
	Results         []*Plan `json:"results"`
}

func (p *Plan) String() string {
	var str string
	if p.Id != "" {
		str = fmt.Sprintf("%x - %s", string(p.Id), p.Name)
	} else {
		str = fmt.Sprintf("%s", p.Name)
	}
	return str
}
