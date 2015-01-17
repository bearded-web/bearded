package plan

import "gopkg.in/mgo.v2/bson"

type WorkflowStep struct {
	Plugin        string `json:"plugin"`
	Description   string `json:"description"`
	Configuration string `json:"configuration"` // plugin configuration in json format
}

type Plan struct {
	Id            bson.ObjectId   `json:"id" bson:"_id,omitempty"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`   // human readable description
	Configuration string          `json:"configuration"` // global plan configuration in json format
	Workflow      []*WorkflowStep `json:"workflow"`
}
