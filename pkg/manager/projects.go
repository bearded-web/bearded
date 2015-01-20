package manager

import (
	"github.com/bearded-web/bearded/models/project"
	"github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

const defaultProject = "Default"

type ProjectManager struct {
	manager *Manager
	col     *mgo.Collection // default collection
}

func (m *ProjectManager) Init() error {
	logrus.Infof("Initialize project indexes")
	err := m.col.EnsureIndex(mgo.Index{
		Key:        []string{"owner", "name"},
		Unique:     true,
		Background: false,
	})
	if err != nil {
		return err
	}
	err = m.col.EnsureIndex(mgo.Index{
		Key:        []string{"owner"},
		Background: false,
	})
	return err
}

func (m *ProjectManager) All() ([]*project.Project, int, error) {
	results := []*project.Project{}
	count, err := m.manager.All(m.col, &results)
	return results, count, err
}

func (m *ProjectManager) GetById(id bson.ObjectId) (*project.Project, error) {
	u := &project.Project{}
	return u, m.manager.GetById(m.col, id, u)
}

func (m *ProjectManager) GetByOwner(owner string) ([]*project.Project, int, error) {
	results := []*project.Project{}
	q := m.col.Find(bson.D{{"owner", m.manager.ToId(owner)}})
	if err := q.All(&results); err != nil {
		return results, 0, err
	}
	count, err := q.Count()
	if err != nil {
		return results, 0, err
	}
	return results, count, nil
}

func (m *ProjectManager) Create(raw *project.Project) (*project.Project, error) {
	// TODO (m0sth8): add validation
	raw.Id = bson.NewObjectId()
	raw.Created = time.Now()
	raw.Updated = raw.Created
	if err := m.col.Insert(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (m *ProjectManager) CreateDefault(owner string) (*project.Project, error) {
	p := &project.Project{
		Owner: m.manager.ToId(owner),
		Name:  defaultProject,
	}
	return m.Create(p)
}

func (m *ProjectManager) Update(obj *project.Project) error {
	obj.Updated = time.Now()
	return m.col.UpdateId(obj.Id, obj)
}
