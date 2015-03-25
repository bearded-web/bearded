package manager

import (
	"io"

	"github.com/facebookgo/stackerr"
	"gopkg.in/mgo.v2"

	"github.com/bearded-web/bearded/models/file"
)

type FileManager struct {
	manager *Manager
	grid    *mgo.GridFS
}

func (m *FileManager) Init() error {
	return nil
}

// Get file by id, don't forget to close file after
func (m *FileManager) GetById(id string) (*file.File, error) {
	f, err := m.grid.OpenId(id)
	if err != nil {
		return nil, stackerr.Wrap(err)
	}
	meta := &file.Meta{}
	if err = f.GetMeta(meta); err != nil {
		return nil, stackerr.Wrap(err)
	}
	return &file.File{Meta: meta, ReadCloser: f}, nil
}

// create file with data
func (m *FileManager) Create(r io.Reader, metaInfo *file.Meta) (*file.Meta, error) {
	f, err := m.grid.Create("")
	// according to gridfs code, the error here is impossible
	if err != nil {
		return nil, stackerr.Wrap(err)
	}
	size, err := io.Copy(f, r)
	if err != nil {
		return nil, stackerr.Wrap(err)
	}
	meta := &file.Meta{
		Id:          file.UniqueFileId(),
		Size:        int(size),
		ContentType: metaInfo.ContentType,
		Name:        metaInfo.Name,
	}
	f.SetId(meta.Id)
	f.SetMeta(meta)
	if meta.ContentType != "" {
		f.SetContentType(meta.ContentType)
	}
	if err = f.Close(); err != nil {
		return nil, stackerr.Wrap(err)
	}
	return meta, nil
}
