package client

import (
	"bytes"
	"fmt"
	"io"

	"code.google.com/p/go.net/context"
	"github.com/bearded-web/bearded/models/file"
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

func (s *FilesService) Download(ctx context.Context, id string) (io.Reader, error) {
	buf := new(bytes.Buffer)
	err := s.client.Get(ctx, fmt.Sprintf("%s/%s", filesUrl, id), "download", buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
