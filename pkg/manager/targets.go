package manager

import (
	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/issue"
	"github.com/bearded-web/bearded/models/target"
	"github.com/bearded-web/bearded/pkg/fltr"
	"github.com/bearded-web/bearded/pkg/utils"
)

type TargetFltr struct {
	Project bson.ObjectId     `fltr:"project,in"`
	Type    target.TargetType `fltr:"type,in"`
	Updated time.Time         `fltr:"updated,gte,lte"`
	Created time.Time         `fltr:"created,gte,lte"`
}

type TargetManager struct {
	manager *Manager
	col     *mgo.Collection // default collection
}

func (m *TargetManager) Init() error {
	logrus.Infof("Initialize target indexes")
	err := m.col.EnsureIndex(mgo.Index{
		Key:        []string{"project"},
		Background: false,
	})
	return err
}

func (m *TargetManager) All() ([]*target.Target, int, error) {
	results := []*target.Target{}

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

func (m *TargetManager) GetById(id bson.ObjectId) (*target.Target, error) {
	u := &target.Target{}
	return u, m.manager.GetById(m.col, id, &u)
}

func (m *TargetManager) FilterBy(f *FeedItemFltr) ([]*target.Target, int, error) {
	query := fltr.GetQuery(f)
	return m.FilterByQuery(query)
}

func (m *TargetManager) FilterByQuery(query bson.M) ([]*target.Target, int, error) {
	results := []*target.Target{}
	count, err := m.manager.FilterBy(m.col, &query, &results)
	return results, count, err
}

func (m *TargetManager) Create(raw *target.Target) (*target.Target, error) {
	// TODO (m0sth8): add validation
	raw.Id = bson.NewObjectId()
	raw.Created = time.Now().UTC()
	raw.Updated = raw.Created
	raw.SummaryReport = &target.SummaryReport{
		Issues: map[issue.Severity]int{},
	}
	if err := m.col.Insert(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (m *TargetManager) Update(obj *target.Target) error {
	obj.Updated = time.Now().UTC()
	return m.col.UpdateId(obj.Id, obj)
}

func (m *TargetManager) Remove(obj *target.Target) error {
	return m.col.RemoveId(obj.Id)
}

func (m *TargetManager) UpdateSummary(obj *target.Target) error {
	fltr := &IssueFltr{
		Target:   obj.Id,
		False:    utils.BoolP(false),
		Resolved: utils.BoolP(false),
		Muted:    utils.BoolP(false),
	}
	// TODO(m0sth8): get only severity/count through aggregation
	issues, _, err := m.manager.Issues.FilterBy(fltr)
	if err != nil {
		return err
	}
	summary := map[issue.Severity]int{}
	for _, issue := range issues {
		summary[issue.Severity] = summary[issue.Severity] + 1
	}
	if obj.SummaryReport == nil {
		obj.SummaryReport = &target.SummaryReport{}
	}
	obj.SummaryReport.Issues = summary
	// TODO(m0sth8): update only summary field
	return m.Update(obj)
}

func (m *TargetManager) UpdateSummaryById(id bson.ObjectId) error {
	obj, err := m.GetById(id)
	if err != nil {
		return err
	}
	return m.UpdateSummary(obj)
}