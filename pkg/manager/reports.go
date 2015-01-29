package manager

import (
	"time"

	"github.com/Sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/report"
	"github.com/bearded-web/bearded/pkg/fltr"
)

type ReportManager struct {
	manager *Manager
	col     *mgo.Collection
}

type ReportFltr struct {
	Type report.ReportType `fltr:"type,in,nin"`
}

func (s *ReportManager) Init() error {
	logrus.Infof("Initialize report indexes")
	err := s.col.EnsureIndex(mgo.Index{
		Key:        []string{"scanSession"},
		Unique:     true,
		Background: false,
	})
	if err != nil {
		return err
	}
	for _, index := range []string{"type", "scan"} {
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

func (m *ReportManager) Fltr() *ReportFltr {
	return &ReportFltr{}
}

func (m *ReportManager) GetById(id bson.ObjectId) (*report.Report, error) {
	u := &report.Report{}
	return u, m.manager.GetById(m.col, id, &u)
}

func (m *ReportManager) GetBySession(sessId bson.ObjectId) (*report.Report, error) {
	query := bson.M{"scanSession": sessId}
	u := &report.Report{}
	return u, m.manager.GetBy(m.col, &query, &u)
}

func (m *ReportManager) All() ([]*report.Report, int, error) {
	results := []*report.Report{}
	count, err := m.manager.All(m.col, &results)
	return results, count, err
}

func (m *ReportManager) FilterBy(f *ReportFltr) ([]*report.Report, int, error) {
	query := fltr.GetQuery(f)
	results := []*report.Report{}
	count, err := m.manager.FilterBy(m.col, &query, &results)
	return results, count, err
}

func (m *ReportManager) FilterByQuery(query bson.M) ([]*report.Report, int, error) {
	results := []*report.Report{}
	count, err := m.manager.FilterBy(m.col, &query, &results)
	return results, count, err
}

func (m *ReportManager) Create(raw *report.Report) (*report.Report, error) {
	// TODO (m0sth8): add validattion
	raw.Id = bson.NewObjectId()
	raw.Created = time.Now().UTC()
	raw.Updated = raw.Created
	if err := m.col.Insert(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (m *ReportManager) Update(obj *report.Report) error {
	obj.Updated = time.Now().UTC()
	return m.col.UpdateId(obj.Id, obj)
}

func (m *ReportManager) Remove(obj *report.Report) error {
	return m.col.RemoveId(obj.Id)
}
