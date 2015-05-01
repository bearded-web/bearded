package target

import (
	"github.com/bearded-web/bearded/models/file"
	"github.com/bearded-web/bearded/models/target"
)

type CommentEntity struct {
	Text string `json:"text" description:"raw markdown text"`
}

type WebTargetEntity struct {
	Domain string `json:"domain" cweb:"nonzero"`
}

type AndroidTargetEntity struct {
	Name string     `json:"name,omitempty" description:"target name, 80 symbols max" cmobile:"nonzero" validate:"max=80"`
	File *file.Meta `json:"file" description:"apk file metadata"`
}

type TargetEntity struct {
	Type    target.TargetType    `json:"type,omitempty" description:"one of [web|android]" creating:"nonzero"`
	Web     *WebTargetEntity     `json:"web,omitempty" description:"information about web target" cweb:"nonzero"`
	Android *AndroidTargetEntity `json:"android,omitempty" description:"information about android target" cmobile:"nonzero"`
	Project string               `json:"project,omitempty" creating:"nonzero,bsonId"`
}
