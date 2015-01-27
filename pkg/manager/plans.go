package manager

import (
	"time"

	"github.com/Sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/plan"
)

type PlanManager struct {
	manager *Manager
	col     *mgo.Collection
}

type PlanFltr struct {
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
	return nil
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

func (m *PlanManager) FilterBy(fltr *PlanFltr) ([]*plan.Plan, int, error) {
	query := bson.M{}

	if fltr != nil {

	}

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
