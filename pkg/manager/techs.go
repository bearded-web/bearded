package manager

// TargetTechs manager

import (
	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/tech"
	"github.com/bearded-web/bearded/pkg/fltr"
)

type TechManager struct {
	manager *Manager
	col     *mgo.Collection
}

type TechFltr struct {
	Updated time.Time       `fltr:"updated,gte,gt,lte,lt"`
	Created time.Time       `fltr:"created,gte,gt,lte,lt"`
	Target  bson.ObjectId   `fltr:"target,in"`
	Project bson.ObjectId   `fltr:"project"`
	Status  tech.StatusType `fltr:"status,in,nin"`
}

func (s *TechManager) Init() error {
	logrus.Infof("Initialize tech indexes")
	err := s.col.EnsureIndex(mgo.Index{
		Key:        []string{"target", "name", "version"},
		Unique:     true,
		Background: false,
	})
	if err != nil {
		return err
	}

	// TODO (m0sth8): check what indexes are really used
	for _, index := range []string{"created", "updated", "target", "project", "status"} {
		err := s.col.EnsureIndex(mgo.Index{
			Key:        []string{index},
			Background: true,
		})
		if err != nil {
			return err
		}
	}
	if s.manager.Cfg.TextSearchEnable {
		logrus.Infof("Create text indexes for tech")
		err := s.col.EnsureIndex(mgo.Index{
			Key:             []string{"$text:name"},
			Background:      true,
			DefaultLanguage: "english",
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *TechManager) Fltr() *TechFltr {
	return &TechFltr{}
}

func (m *TechManager) GetById(id bson.ObjectId) (*tech.TargetTech, error) {
	u := &tech.TargetTech{}
	return u, m.manager.GetById(m.col, id, &u)
}

func (m *TechManager) GetByUniqId(target bson.ObjectId, uniqId string) (*tech.TargetTech, error) {
	u := &tech.TargetTech{}
	return u, m.manager.GetBy(m.col, &bson.M{"target": target, "uniqId": uniqId}, &u)
}

func (m *TechManager) FilterBy(f *TechFltr, opts ...Opts) ([]*tech.TargetTech, int, error) {
	query := fltr.GetQuery(f)
	return m.FilterByQuery(query, opts...)
}

func (m *TechManager) FilterByQuery(query bson.M, opts ...Opts) ([]*tech.TargetTech, int, error) {
	results := []*tech.TargetTech{}
	count, err := m.manager.FilterBy(m.col, &query, &results, opts...)
	return results, count, err
}

func (m *TechManager) Create(raw *tech.TargetTech) (*tech.TargetTech, error) {
	// TODO (m0sth8): add validation
	raw.Id = bson.NewObjectId()
	raw.Created = time.Now().UTC()
	raw.Updated = raw.Created
	if raw.Status == "" {
		raw.Status = tech.StatusUnknown
	}
	if err := m.col.Insert(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (m *TechManager) Update(obj *tech.TargetTech) error {
	obj.Updated = time.Now().UTC()
	return m.col.UpdateId(obj.Id, obj)
}

func (m *TechManager) Remove(obj *tech.TargetTech) error {
	return m.col.RemoveId(obj.Id)
}

func (m *TechManager) RemoveAll(query bson.M) (int, error) {
	info, err := m.col.RemoveAll(query)
	if info != nil {
		return info.Removed, err
	}
	return 0, err
}
