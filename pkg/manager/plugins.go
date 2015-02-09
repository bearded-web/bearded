package manager

import (
	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"strings"

	"github.com/bearded-web/bearded/models/plugin"
	"github.com/bearded-web/bearded/pkg/fltr"
)

type PluginManager struct {
	manager *Manager
	col     *mgo.Collection
}

type PluginFltr struct {
	Name    string            `fltr:"name,in"`
	Version string            `fltr:"version,in,nin,gte,gt,lte,lte"`
	Type    plugin.PluginType `fltr:"type,in,nin"`
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
	for _, index := range []string{"name", "version", "type"} {
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
	query["enabled"] = true
	opts := Opts{}
	// if version is not specified, then look for the biggest version
	if version != "" {
		opts.Sort = []string{"-version"}
	} else {
		query["version"] = version
	}
	return p, m.manager.GetBy(m.col, &query, p, opts)
}

// plugin name must be in format "name:version"
func (m *PluginManager) GetByName(name string) (*plugin.Plugin, error) {
	plNameVersion := strings.Split(name, ":")
	version := ""
	if len(plNameVersion) > 1 {
		version = plNameVersion[1]
	}
	return m.GetByNameVersion(plNameVersion[0], version)
}

func (m *PluginManager) All() ([]*plugin.Plugin, int, error) {
	results := []*plugin.Plugin{}

	count, err := m.manager.All(m.col, &results)
	return results, count, err
}

func (m *PluginManager) FilterBy(f *PluginFltr) ([]*plugin.Plugin, int, error) {
	query := fltr.GetQuery(f)
	results := []*plugin.Plugin{}
	count, err := m.manager.FilterBy(m.col, &query, &results)
	return results, count, err
}

func (m *PluginManager) FilterByQuery(query bson.M) ([]*plugin.Plugin, int, error) {
	results := []*plugin.Plugin{}
	count, err := m.manager.FilterBy(m.col, &query, &results)
	return results, count, err
}

func (m *PluginManager) Create(raw *plugin.Plugin) (*plugin.Plugin, error) {
	// TODO (m0sth8): add validattion
	raw.Id = bson.NewObjectId()
	raw.Created = time.Now().UTC()
	raw.Updated = raw.Created
	if err := m.col.Insert(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (m *PluginManager) Update(obj *plugin.Plugin) error {
	obj.Updated = time.Now().UTC()
	return m.col.UpdateId(obj.Id, obj)
}
