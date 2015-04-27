package manager

// TargetIssues manager

import (
	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/issue"
	"github.com/bearded-web/bearded/pkg/fltr"
)

type IssueManager struct {
	manager *Manager
	col     *mgo.Collection
}

type IssueFltr struct {
	Updated   time.Time      `fltr:"updated,gte,gt,lte,lt"`
	Created   time.Time      `fltr:"created,gte,gt,lte,lt"`
	Target    bson.ObjectId  `fltr:"target,in"`
	Project   bson.ObjectId  `fltr:"project"`
	Confirmed bool           `fltr:"confirmed"`
	Muted     bool           `fltr:"muted"`
	Resolved  bool           `fltr:"resolved"`
	False     bool           `fltr:"false"`
	Severity  issue.Severity `fltr:"severity,in"`
}

func (s *IssueManager) Init() error {
	logrus.Infof("Initialize issue indexes")
	err := s.col.EnsureIndex(mgo.Index{
		Key:        []string{"target", "uniqId"},
		Unique:     true,
		Background: false,
	})
	if err != nil {
		return err
	}

	// TODO (m0sth8): check what indexes are really used
	for _, index := range []string{"created", "updated", "target", "project"} {
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

func (m *IssueManager) Fltr() *IssueFltr {
	return &IssueFltr{}
}

func (m *IssueManager) GetById(id bson.ObjectId) (*issue.TargetIssue, error) {
	u := &issue.TargetIssue{}
	return u, m.manager.GetById(m.col, id, &u)
}

func (m *IssueManager) FilterBy(f *IssueFltr, opts ...Opts) ([]*issue.TargetIssue, int, error) {
	query := fltr.GetQuery(f)
	return m.FilterByQuery(query, opts...)
}

func (m *IssueManager) FilterByQuery(query bson.M, opts ...Opts) ([]*issue.TargetIssue, int, error) {
	results := []*issue.TargetIssue{}
	count, err := m.manager.FilterBy(m.col, &query, &results, opts...)
	return results, count, err
}

func (m *IssueManager) Create(raw *issue.TargetIssue) (*issue.TargetIssue, error) {
	// TODO (m0sth8): add validation
	raw.Id = bson.NewObjectId()
	raw.Created = time.Now().UTC()
	raw.Updated = raw.Created
	if len(raw.Severity) == 0 {
		raw.Severity = issue.SeverityInfo
	}
	if len(raw.UniqId) == 0 {
		raw.UniqId = raw.Id.Hex()
	}
	if err := m.col.Insert(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (m *IssueManager) Update(obj *issue.TargetIssue) error {
	obj.Updated = time.Now().UTC()
	return m.col.UpdateId(obj.Id, obj)
}

func (m *IssueManager) Remove(obj *issue.TargetIssue) error {
	return m.col.RemoveId(obj.Id)
}

func (m *IssueManager) RemoveAll(query bson.M) (int, error) {
	info, err := m.col.RemoveAll(query)
	if info != nil {
		return info.Removed, err
	}
	return 0, err
}
