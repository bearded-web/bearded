package manager

import (
	"time"

	"github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/scan"
)

type ScanManager struct {
	manager *Manager
	col     *mgo.Collection
}

type ScanFltr struct {
}

func (s *ScanManager) Init() error {
	logrus.Infof("Initialize scan indexes")
	for _, index := range []string{"owner"} {
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

func (m *ScanManager) Fltr() *ScanFltr {
	return &ScanFltr{}
}

func (m *ScanManager) GetById(id string) (*scan.Scan, error) {
	u := &scan.Scan{}
	if err := m.col.FindId(bson.ObjectIdHex(id)).One(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (m *ScanManager) All() ([]*scan.Scan, int, error) {
	results := []*scan.Scan{}

	count, err := m.manager.All(m.col, &results)
	return results, count, err
}

func (m *ScanManager) FilterBy(fltr *ScanFltr) ([]*scan.Scan, int, error) {
	query := bson.M{}

	if fltr != nil {

	}

	results := []*scan.Scan{}
	count, err := m.manager.FilterBy(m.col, &query, &results)
	return results, count, err

}

func (m *ScanManager) Create(raw *scan.Scan) (*scan.Scan, error) {
	// TODO (m0sth8): add validattion
	raw.Id = bson.NewObjectId()
	raw.Dates.Created = TimeP(time.Now())
	raw.Dates.Updated = raw.Dates.Created
	if err := m.col.Insert(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (m *ScanManager) Update(obj *scan.Scan) error {
	obj.Dates.Updated = TimeP(time.Now())
	return m.col.UpdateId(obj.Id, obj)
}

func (m *ScanManager) Remove(obj *scan.Scan) error {
	return m.col.RemoveId(obj.Id)
}
