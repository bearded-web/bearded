package manager

import (
	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"strings"

	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/pkg/fltr"
	"github.com/bearded-web/bearded/pkg/utils"
)

type UserManager struct {
	manager *Manager
	col     *mgo.Collection // default collection
}

type UserFltr struct {
	Id      bson.ObjectId `fltr:"id,in" bson:"_id"`
	Email   string        `fltr:"email"`
	Created time.Time     `fltr:"created,gte,lte"`
}

func (m *UserManager) Init() error {
	logrus.Infof("Initialize user indexes")
	err := m.col.EnsureIndex(mgo.Index{
		Key:        []string{"email"},
		Unique:     true,
		Background: false,
	})
	if err != nil {
		return err
	}
	for _, index := range []string{"email", "created"} {
		err := m.col.EnsureIndex(mgo.Index{
			Key:        []string{index},
			Background: true,
		})
		if err != nil {
			return err
		}
	}
	// TODO (m0sth8): extract system users creation to project initialization
	agent := &user.User{
		Email:    "agent@barbudo.net",
		Password: "",
	}
	if _, err := m.Create(agent); err != nil {
		if !m.manager.IsDup(err) {
			return err
		}
	}

	// TODO (m0sth8): remove after the next release
	users := []*user.User{}
	if _, err := m.manager.FilterBy(m.col, &bson.M{"nickname": bson.M{"$exists": false}}, &users); err != nil {
		return err
	}
	for _, user := range users {
		m.Update(user)
	}

	return err
}

func (m *UserManager) GetById(id bson.ObjectId) (*user.User, error) {
	obj := &user.User{}
	if err := m.col.FindId(id).One(obj); err != nil {
		return nil, err
	}
	if obj.Avatar == "" {
		obj.Avatar = utils.GetGravatar(obj.Email, 38, utils.AvatarRetro)
	}
	return obj, nil
}

func (m *UserManager) GetByEmail(email string) (*user.User, error) {
	u := &user.User{}
	if err := m.col.Find(bson.D{{Name: "email", Value: email}}).One(u); err != nil {
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

func (m *UserManager) FilterBy(f *UserFltr) ([]*user.User, int, error) {
	query := fltr.GetQuery(f)
	results := []*user.User{}
	count, err := m.manager.FilterBy(m.col, &query, &results)
	return results, count, err
}

func (m *UserManager) FilterByQuery(query bson.M) ([]*user.User, int, error) {
	results := []*user.User{}
	count, err := m.manager.FilterBy(m.col, &query, &results)
	return results, count, err
}

func (m *UserManager) Create(raw *user.User) (*user.User, error) {
	// TODO (m0sth8): add validation
	raw.Id = bson.NewObjectId()
	raw.Created = time.Now().UTC()
	raw.Updated = raw.Created
	raw.Avatar = utils.GetGravatar(raw.Email, 38, utils.AvatarRetro)
	if raw.Nickname == "" {
		raw.Nickname = strings.Split(raw.Email, "@")[0]
	}

	if err := m.col.Insert(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (m *UserManager) Update(obj *user.User) error {
	obj.Updated = time.Now().UTC()
	if obj.Avatar == "" {
		obj.Avatar = utils.GetGravatar(obj.Email, 38, utils.AvatarRetro)
	}
	if obj.Nickname == "" {
		obj.Nickname = strings.Split(obj.Email, "@")[0]
	}
	return m.col.UpdateId(obj.Id, obj)
}
