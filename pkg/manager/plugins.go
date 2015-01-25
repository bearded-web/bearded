package manager

import (
	"time"

	"github.com/Sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/plugin"
)

type PluginManager struct {
	manager *Manager
	col     *mgo.Collection
}

type PluginFltr struct {
	Name    string
	Version string
	Type    plugin.PluginType
}

func (s *PluginManager) Init() error {
	logrus.Infof("Initialize plugin indexes")
	err := s.col.EnsureIndex(mgo.Index{
		Key:        []string{"name", "version"},
		Unique:     true,
		Background: false,
	})
	if err != nil {
		return err
	}
	for _, index := range []string{"name", "version"} {
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

func (m *PluginManager) Fltr() *PluginFltr {
	return &PluginFltr{}
}

func (m *PluginManager) GetById(id string) (*plugin.Plugin, error) {
	u := &plugin.Plugin{}
	if err := m.col.FindId(bson.ObjectIdHex(id)).One(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (m *PluginManager) GetByNameVersion(name, version string) (*plugin.Plugin, error) {
	p := &plugin.Plugin{}
	query := bson.M{}
	query["name"] = name
	query["version"] = version
	return p, m.manager.GetBy(m.col, &query, p)
}

func (m *PluginManager) All() ([]*plugin.Plugin, int, error) {
	results := []*plugin.Plugin{}

	count, err := m.manager.All(m.col, &results)
	return results, count, err
}

func (m *PluginManager) FilterBy(fltr *PluginFltr) ([]*plugin.Plugin, int, error) {
	query := bson.M{}

	if fltr != nil {
		if fltr.Name != "" {
			query["name"] = fltr.Name
		}
		if fltr.Version != "" {
			query["version"] = fltr.Version
		}
		if fltr.Type != "" {
			query["type"] = fltr.Type
		}
	}

	results := []*plugin.Plugin{}
	count, err := m.manager.FilterBy(m.col, &query, &results)
	return results, count, err

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
