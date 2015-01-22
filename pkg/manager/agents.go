package manager

import (
	"time"

	"github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/agent"
)

type AgentManager struct {
	manager *Manager
	col     *mgo.Collection
}

type AgentFltr struct {
	Name string
	Type agent.Type
}

func (s *AgentManager) Init() error {
	logrus.Infof("Initialize agent indexes")
	err := s.col.EnsureIndex(mgo.Index{
		Key:        []string{"name", "type"},
		Unique:     true,
		Background: false,
	})
	if err != nil {
		return err
	}
	for _, index := range []string{"name", "status"} {
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

func (m *AgentManager) Fltr() *AgentFltr {
	return &AgentFltr{}
}

func (m *AgentManager) GetById(id bson.ObjectId) (*agent.Agent, error) {
	u := &agent.Agent{}
	return u, m.manager.GetById(m.col, id, &u)
}

func (m *AgentManager) All() ([]*agent.Agent, int, error) {
	results := []*agent.Agent{}

	count, err := m.manager.All(m.col, &results)
	return results, count, err
}

func (m *AgentManager) FilterBy(fltr *AgentFltr) ([]*agent.Agent, int, error) {
	query := bson.M{}

	if fltr != nil {
		if fltr.Name != "" {
			query["name"] = fltr.Name
		}
		if fltr.Type != "" {
			query["type"] = fltr.Type
		}
	}

	results := []*agent.Agent{}
	count, err := m.manager.FilterBy(m.col, &query, &results)
	return results, count, err

}

func (m *AgentManager) Create(raw *agent.Agent) (*agent.Agent, error) {
	// TODO (m0sth8): add validattion
	raw.Id = bson.NewObjectId()
	raw.Created = time.Now()
	raw.Updated = raw.Created
	if err := m.col.Insert(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (m *AgentManager) Update(obj *agent.Agent) error {
	obj.Updated = time.Now()
	return m.col.UpdateId(obj.Id, obj)
}

func (m *AgentManager) Remove(obj *agent.Agent) error {
	return m.col.RemoveId(obj.Id)
}
