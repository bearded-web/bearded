package manager

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/scan"
	"github.com/bearded-web/bearded/pkg/fltr"
)

type ScanManager struct {
	manager *Manager
	col     *mgo.Collection
}

type ScanFltr struct {
	Status  scan.ScanStatus `fltr:"status,in,nin"`
	Target  bson.ObjectId   `fltr:"target,in"`
	Project bson.ObjectId   `fltr:"project"`
	Plan    bson.ObjectId   `fltr:"plan,in"`
}

func (s *ScanManager) Init() error {
	logrus.Infof("Initialize scan indexes")
	for _, index := range []string{"owner", "status"} {
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

func (m *ScanManager) GetById(id bson.ObjectId) (*scan.Scan, error) {
	u := &scan.Scan{}
	if err := m.col.FindId(id).One(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (m *ScanManager) GetByMulti(ids []bson.ObjectId) (map[bson.ObjectId]*scan.Scan, error) {
	scans := []*scan.Scan{}
	query := bson.M{
		"id": bson.M{"$in": ids},
	}
	if err := m.col.Find(query).All(&scans); err != nil {
		return nil, err
	}
	results := map[bson.ObjectId]*scan.Scan{}
	for _, sc := range scans {
		results[sc.Id] = sc
	}
	return results, nil
}

func (m *ScanManager) All() ([]*scan.Scan, int, error) {
	results := []*scan.Scan{}

	count, err := m.manager.All(m.col, &results)
	return results, count, err
}

func (m *ScanManager) FilterBy(f *ScanFltr) ([]*scan.Scan, int, error) {
	query := fltr.GetQuery(f)
	results := []*scan.Scan{}
	count, err := m.manager.FilterBy(m.col, &query, &results)
	return results, count, err

}

func (m *ScanManager) FilterByQuery(query bson.M) ([]*scan.Scan, int, error) {
	results := []*scan.Scan{}
	count, err := m.manager.FilterBy(m.col, &query, &results)
	return results, count, err
}

func (m *ScanManager) Create(raw *scan.Scan) (*scan.Scan, error) {
	// TODO (m0sth8): add validation
	raw.Id = bson.NewObjectId()
	raw.Dates.Created = TimeP(time.Now().UTC())
	raw.Dates.Updated = raw.Dates.Created
	for _, sess := range raw.Sessions {
		sess.Scan = raw.Id
	}
	if err := m.col.Insert(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (m *ScanManager) Update(obj *scan.Scan) error {
	now := time.Now().UTC()
	obj.Dates.Updated = &now
	switch st := obj.Status; true {
	case st == scan.StatusQueued:
		if obj.Queued == nil {
			obj.Queued = &now
		}
	case st == scan.StatusFinished || st == scan.StatusFailed:
		if obj.Finished == nil {
			obj.Finished = &now
		}
	case st == scan.StatusWorking:
		if obj.Started == nil {
			obj.Started = &now
		}
	}
	return m.col.UpdateId(obj.Id, obj)
}

func (m *ScanManager) Remove(obj *scan.Scan) error {
	return m.col.RemoveId(obj.Id)
}

// sessions

func (m *ScanManager) UpdateSession(sc *scan.Scan, obj *scan.Session) error {
	var index *int
	isRootSession := !obj.HasParent()
	if isRootSession {
		// find session indx for updating, but only for root session
		for i, s := range sc.Sessions {
			if s.Id == obj.Id {
				index = &i
				break
			}
		}
	}
	if isRootSession && index == nil {
		return mgo.ErrNotFound
	}
	now := time.Now().UTC()
	switch st := obj.Status; true {
	case st == scan.StatusQueued:
		if obj.Queued == nil {
			obj.Queued = &now
		}
	case st == scan.StatusFinished || st == scan.StatusFailed:
		if obj.Finished == nil {
			obj.Finished = &now
		}
	case st == scan.StatusWorking:
		if obj.Started == nil {
			obj.Started = &now
		}
	}

	scanModified := false

	// only root sessions have influence on scan
	if isRootSession {
		// if session is queued then scan should also be queued
		if obj.Status == scan.StatusQueued && (sc.Status == scan.StatusCreated || sc.Status == scan.StatusWorking) {
			sc.Status = scan.StatusQueued
			scanModified = true
		}

		// if session is working then scan should also be working
		if obj.Status == scan.StatusWorking && (sc.Status != scan.StatusWorking) {
			sc.Status = scan.StatusWorking
			scanModified = true
		}

		// if session failed then scan should be failed too
		if obj.Status == scan.StatusFailed {
			sc.Status = scan.StatusFailed
			scanModified = true
		}
		// if session was the last one
		if obj.Status == scan.StatusFinished && (*index+1) == len(sc.Sessions) {
			sc.Status = scan.StatusFinished
			scanModified = true
		}
	}

	// if scan modified then update the whole scan object
	if scanModified || !isRootSession {
		return m.Update(sc)
	}

	// TODO (m0sth8): update non root session with set
	key := fmt.Sprintf("sessions.%d", *index)
	update := bson.M{"$set": bson.M{key: obj, "updated": now}}
	m.col.UpdateId(sc.Id, update)
	return m.Update(sc)
}
