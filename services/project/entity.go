package project

import "github.com/bearded-web/bearded/models/project"

type ProjectEntity struct {
	Name string `json:"name"`
	//	Owner   bson.ObjectId `json:"owner,omitempty"`
	Members []*project.Member `json:"members"`
}
