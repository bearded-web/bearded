package manager

import (
	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/plan"
	"github.com/bearded-web/bearded/models/target"
	"github.com/bearded-web/bearded/pkg/fltr"
)

type PlanManager struct {
	manager *Manager
	col     *mgo.Collection
}

type PlanFltr struct {
	Name       string            `fltr:"name"`
	TargetType target.TargetType `fltr:"targetType,in"`
}

func (s *PlanManager) Init() error {
	logrus.Infof("Initialize plan indexes")
	err := s.col.EnsureIndex(mgo.Index{
		Key:        []string{"name"},
		Unique:     true,
		Background: false,
	})
	if err != nil {
		return err
	}
	for _, index := range []string{} {
		err := s.col.EnsureIndex(mgo.Index{
			Key:        []string{index},
			Background: true,
		})
		if err != nil {
			return err
		}
	}
	// temporary fix
	_, err = s.col.UpdateAll(bson.M{"targetType": bson.M{"$exists": false}}, bson.M{"$set": bson.M{"targetType": target.TypeWeb}})
	return err
}

func (m *PlanManager) Fltr() *PlanFltr {
	return &PlanFltr{}
}

func (m *PlanManager) GetById(id bson.ObjectId) (*plan.Plan, error) {
	u := &plan.Plan{}
	return u, m.manager.GetById(m.col, id, &u)
}

func (m *PlanManager) All() ([]*plan.Plan, int, error) {
	results := []*plan.Plan{}

	count, err := m.manager.All(m.col, &results)
	return results, count, err
}

func (m *PlanManager) FilterBy(f *PlanFltr) ([]*plan.Plan, int, error) {
	query := fltr.GetQuery(f)
	results := []*plan.Plan{}
	count, err := m.manager.FilterBy(m.col, &query, &results)
	return results, count, err
}

func (m *PlanManager) FilterByQuery(query bson.M) ([]*plan.Plan, int, error) {
	results := []*plan.Plan{}
	count, err := m.manager.FilterBy(m.col, &query, &results)
	return results, count, err
}

func (m *PlanManager) Create(raw *plan.Plan) (*plan.Plan, error) {
	// TODO (m0sth8): add validattion
	raw.Id = bson.NewObjectId()
	raw.Created = time.Now().UTC()
	raw.Updated = raw.Created
	if err := m.col.Insert(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (m *PlanManager) Update(obj *plan.Plan) error {
	obj.Updated = time.Now().UTC()
	return m.col.UpdateId(obj.Id, obj)
}

func (m *PlanManager) Remove(obj *plan.Plan) error {
	return m.col.RemoveId(obj.Id)
}
