package manager

// Tokens manager

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/token"
	"github.com/bearded-web/bearded/pkg/fltr"
	"github.com/bearded-web/bearded/pkg/utils"
)

type TokenManager struct {
	manager *Manager
	col     *mgo.Collection
}

type TokenFltr struct {
	Updated time.Time     `fltr:"updated,gte,gt,lte,lt"`
	Created time.Time     `fltr:"created,gte,gt,lte,lt"`
	User    bson.ObjectId `fltr:"user"`
	Removed bool          `fltr:"removed"`
}

func (s *TokenManager) Init() error {
	logrus.Infof("Initialize token indexes")

	// TODO (m0sth8): check what indexes are really used
	for _, index := range []string{"created", "updated", "user"} {
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

func (m *TokenManager) Fltr() *TokenFltr {
	return &TokenFltr{}
}

func (m *TokenManager) GetById(id bson.ObjectId) (*token.Token, error) {
	u := &token.Token{}
	return u, m.manager.GetById(m.col, id, &u)
}

func (m *TokenManager) FilterBy(f *TokenFltr, opts ...Opts) ([]*token.Token, int, error) {
	query := fltr.GetQuery(f)
	return m.FilterByQuery(query, opts...)
}

func (m *TokenManager) FilterByQuery(query bson.M, opts ...Opts) ([]*token.Token, int, error) {
	results := []*token.Token{}
	count, err := m.manager.FilterBy(m.col, &query, &results, opts...)
	return results, count, err
}

func (m *TokenManager) Create(raw *token.Token) (*token.Token, error) {
	// TODO (m0sth8): add validation
	raw.Id = bson.NewObjectId()
	raw.Created = time.Now().UTC()
	raw.Updated = raw.Created
	if len(raw.Name) == 0 {
		raw.Name = fmt.Sprintf("%s", raw.Created.Format(time.UnixDate))
	}
	if len(raw.Hash) == 0 {
		raw.Hash = utils.RandomString(token.TokenLength)
	}
	if len(raw.Scopes) == 0 {
		raw.Scopes = append(raw.Scopes, token.DefaultScope)
	}
	if err := m.col.Insert(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (m *TokenManager) Update(obj *token.Token) error {
	obj.Updated = time.Now().UTC()
	return m.col.UpdateId(obj.Id, obj)
}

func (m *TokenManager) Remove(obj *token.Token) error {
	obj.Removed = true
	return m.Update(obj)
}

func (m *TokenManager) RemoveAll(query bson.M) (int, error) {
	info, err := m.col.RemoveAll(query)
	if info != nil {
		return info.Removed, err
	}
	return 0, err
}
