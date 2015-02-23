package manager

import (
	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/comment"
	"github.com/bearded-web/bearded/pkg/fltr"
)

type CommentManager struct {
	manager *Manager
	col     *mgo.Collection
}

type CommentFltr struct {
	Type    comment.Type  `fltr:"type"`
	Link    bson.ObjectId `fltr:"link"`
	Updated time.Time     `fltr:"updated,gte,gt,lte,lt"`
	Created time.Time     `fltr:"created,gte,gt,lte,lt"`
}

func (s *CommentManager) Init() error {
	logrus.Infof("Initialize comment indexes")
	err := s.col.EnsureIndex(mgo.Index{
		Key:        []string{"type", "link"},
		Background: true,
	})
	if err != nil {
		return err
	}
	for _, index := range []string{"created", "updated", "owner"} {
		err := s.col.EnsureIndex(mgo.Index{
			Key:        []string{index},
			Background: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *CommentManager) Fltr() *CommentFltr {
	return &CommentFltr{}
}

func (m *CommentManager) GetById(id bson.ObjectId) (*comment.Comment, error) {
	u := &comment.Comment{}
	return u, m.manager.GetById(m.col, id, &u)
}

func (m *CommentManager) FilterBy(f *CommentFltr, opts ...Opts) ([]*comment.Comment, int, error) {
	query := fltr.GetQuery(f)
	return m.FilterByQuery(query, opts...)
}

func (m *CommentManager) FilterByQuery(query bson.M, opts ...Opts) ([]*comment.Comment, int, error) {
	results := []*comment.Comment{}
	count, err := m.manager.FilterBy(m.col, &query, &results, opts...)
	return results, count, err
}

func (m *CommentManager) Create(raw *comment.Comment) (*comment.Comment, error) {
	// TODO (m0sth8): add validation
	raw.Id = bson.NewObjectId()
	raw.Created = time.Now().UTC()
	raw.Updated = raw.Created
	if err := m.col.Insert(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (m *CommentManager) Update(obj *comment.Comment) error {
	obj.Updated = time.Now().UTC()
	return m.col.UpdateId(obj.Id, obj)
}

func (m *CommentManager) Remove(obj *comment.Comment) error {
	return m.col.RemoveId(obj.Id)
}
