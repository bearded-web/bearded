package manager

import (
	"time"

	"github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/plugin"
)

type PluginManager struct {
	manager *Manager
	col     *mgo.Collection
}

func (s *PluginManager) Init() error {
	logrus.Infof("Initialize plugin indexes")
	err := s.col.EnsureIndex(mgo.Index{
		Key:        []string{"name", "version"},
		Unique:     true,
		Background: false,
	})
	return err
}

func (m *PluginManager) GetById(id string) (*plugin.Plugin, error) {
	u := &plugin.Plugin{}
	if err := m.col.FindId(bson.ObjectIdHex(id)).One(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (m *PluginManager) All() ([]*plugin.Plugin, int, error) {
	results := []*plugin.Plugin{}

	query := &bson.M{}
	q := m.col.Find(query)
	if err := q.All(&results); err != nil {
		return nil, 0, err
	}
	count, err := q.Count()
	if err != nil {
		return nil, 0, err
	}
	return results, count, nil
}

func (m *PluginManager) Create(raw *plugin.Plugin) (*plugin.Plugin, error) {
	// TODO (m0sth8): add validattion
	raw.Id = bson.NewObjectId()
	raw.Created = time.Now()
	raw.Updated = raw.Created
	if err := m.col.Insert(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (m *PluginManager) Update(obj *plugin.Plugin) error {
	obj.Updated = time.Now()
	return m.col.UpdateId(obj.Id, obj)
}
