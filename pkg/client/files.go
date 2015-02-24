package client

import (
	"code.google.com/p/go.net/context"

	"github.com/bearded-web/bearded/models/file"
	"io"
)

const (
	filesUrl       = "files"
	filesFieldName = "file"
)

type FilesService struct {
	client *Client
}

func (s *FilesService) String() string {
	return Stringify(s)
}

type FilesListOpts struct {
	Name string `url:"name"`
}

func (s *FilesService) Create(ctx context.Context, filename string, data io.Reader) (*file.Meta, error) {
	meta := &file.Meta{}
	files := []*UploadedFile{
		&UploadedFile{
			Fieldname: filesFieldName,
			Filename:  filename,
			Data:      data,
		},
	}
	return meta, s.client.Upload(ctx, filesUrl, files, meta)
}
