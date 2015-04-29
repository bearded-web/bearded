package file

import (
	"io"

	"github.com/bearded-web/bearded/pkg/utils"
)

type Meta struct {
	Id          string `json:"id,omitempty" description:"unique file id"`
	Name        string `json:"name"`
	Size        int    `json:"size,omitempty"`
	ContentType string `json:"contentType"`
}

type File struct {
	Meta          *Meta `json:"meta"`
	io.ReadCloser `json:"-"`
}

func UniqueFileId() string {
	return utils.UuidV4String()
}
