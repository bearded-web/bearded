package manager

import (
	"github.com/bearded-web/bearded/models/user"
	"github.com/Sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type UserManager struct {
	manager *Manager
	col     *mgo.Collection // default collection
}

func (m *UserManager) Init() error {
	logrus.Infof("Initialize user indexes")
	err := m.col.EnsureIndex(mgo.Index{
		Key:        []string{"email"},
		Unique:     true,
		Background: false,
	})
	return err
}

func (m *UserManager) GetById(id string) (*user.User, error) {
	u := &user.User{}
	if err := m.col.FindId(bson.ObjectIdHex(id)).One(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (m *UserManager) GetByEmail(email string) (*user.User, error) {
	u := &user.User{}
	if err := m.col.Find(bson.D{{"email", email}}).One(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (m *UserManager) All() ([]*user.User, int, error) {
	results := []*user.User{}

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

func (m *UserManager) Create(raw *user.User) (*user.User, error) {
	// TODO (m0sth8): add validation
	raw.Id = bson.NewObjectId()
	raw.Created = time.Now()
	raw.Updated = raw.Created
	if err := m.col.Insert(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (m *UserManager) Update(obj *user.User) error {
	obj.Updated = time.Now()
	return m.col.UpdateId(obj.Id, obj)
}
