package manager

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bearded-web/bearded/models/project"
	"github.com/bearded-web/bearded/pkg/fltr"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const defaultProject = "Default"

type ProjectFltr struct {
	Owner  bson.ObjectId `fltr:"owner"`
	Member bson.ObjectId `fltr:"member" bson:"members.user"`
}

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
	for _, index := range []string{"owner", "members.user"} {
		err := m.col.EnsureIndex(mgo.Index{
			Key:        []string{index},
			Background: true,
		})
		if err != nil {
			return err
		}
	}
	// TODO (m0sth8): exclude to migration
	_, err = m.col.UpdateAll(bson.M{"members": bson.M{"$exists": false}}, bson.M{"$set": bson.M{"members": []*project.Member{}}})
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

func (m *ProjectManager) FilterBy(f *ProjectFltr) ([]*project.Project, int, error) {
	query := fltr.GetQuery(f)
	return m.FilterByQuery(query)
}

func (m *ProjectManager) FilterByQuery(query bson.M, opts ...Opts) ([]*project.Project, int, error) {
	results := []*project.Project{}
	count, err := m.manager.FilterBy(m.col, &query, &results, opts...)
	return results, count, err
}

func (m *ProjectManager) Create(raw *project.Project) (*project.Project, error) {
	// TODO (m0sth8): add validation
	raw.Id = bson.NewObjectId()
	raw.Created = time.Now().UTC()
	raw.Updated = raw.Created
	if raw.Members == nil {
		raw.Members = []*project.Member{}
	}
	if err := m.col.Insert(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (m *ProjectManager) CreateDefault(owner bson.ObjectId) (*project.Project, error) {
	p := &project.Project{
		Owner: owner,
		Name:  defaultProject,
	}
	return m.Create(p)
}

func (m *ProjectManager) Update(obj *project.Project) error {
	obj.Updated = time.Now().UTC()
	return m.col.UpdateId(obj.Id, obj)
}
