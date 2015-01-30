package manager

import (
	"github.com/Sirupsen/logrus"
	"github.com/bearded-web/bearded/models/target"
	"github.com/bearded-web/bearded/pkg/fltr"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type TargetFltr struct {
	Project bson.ObjectId     `fltr:"project,in"`
	Type    target.TargetType `fltr:"type,in"`
	Updated time.Time         `fltr:"updated,gte,lte"`
	Created time.Time         `fltr:"created,gte,lte"`
}

type TargetManager struct {
	manager *Manager
	col     *mgo.Collection // default collection
}

func (m *TargetManager) Init() error {
	logrus.Infof("Initialize target indexes")
	err := m.col.EnsureIndex(mgo.Index{
		Key:        []string{"project"},
		Background: false,
	})
	return err
}

func (m *TargetManager) All() ([]*target.Target, int, error) {
	results := []*target.Target{}

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

func (m *TargetManager) GetById(id bson.ObjectId) (*target.Target, error) {
	u := &target.Target{}
	return u, m.manager.GetById(m.col, id, &u)
}

func (m *TargetManager) FilterBy(f *FeedItemFltr) ([]*target.Target, int, error) {
	query := fltr.GetQuery(f)
	return m.FilterByQuery(query)
}

func (m *TargetManager) FilterByQuery(query bson.M) ([]*target.Target, int, error) {
	results := []*target.Target{}
	count, err := m.manager.FilterBy(m.col, &query, &results)
	return results, count, err
}

func (m *TargetManager) Create(raw *target.Target) (*target.Target, error) {
	// TODO (m0sth8): add validation
	raw.Id = bson.NewObjectId()
	raw.Created = time.Now().UTC()
	raw.Updated = raw.Created
	if err := m.col.Insert(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (m *TargetManager) Update(obj *target.Target) error {
	obj.Updated = time.Now().UTC()
	return m.col.UpdateId(obj.Id, obj)
}

func (m *TargetManager) Remove(obj *target.Target) error {
	return m.col.RemoveId(obj.Id)
}
