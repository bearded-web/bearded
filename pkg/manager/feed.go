package manager

import (
	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/feed"
	"github.com/bearded-web/bearded/models/scan"
	"github.com/bearded-web/bearded/pkg/fltr"
)

type FeedManager struct {
	manager *Manager
	col     *mgo.Collection
}

type FeedItemFltr struct {
	Target  bson.ObjectId `fltr:"target,in"`
	Project bson.ObjectId `fltr:"project,in"`
	Type    feed.ItemType `fltr:"type,in"`
	Updated time.Time     `fltr:"updated,gte,gt,lte,lt"`
}

func (s *FeedManager) Init() error {
	logrus.Infof("Initialize feed indexes")
	for _, index := range []string{"target", "project", "updated", "type"} {
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

func (m *FeedManager) Fltr() *FeedItemFltr {
	return &FeedItemFltr{}
}

func (m *FeedManager) Enrich(item *feed.FeedItem) error {
	switch item.Type {
	case feed.TypeScan:
		if item.ScanId != "" && item.Scan == nil {
			obj, err := m.manager.Scans.GetById(item.ScanId)
			if err != nil {
				return err
			}
			item.Scan = obj
		}
	}
	return nil
}

func (m *FeedManager) EnrichMulti(items []*feed.FeedItem) ([]*feed.FeedItem, error) {
	// TODO (m0sth8): To optimize speed 1) request similar items in bulk. 2) in different go routines

	results := make([]*feed.FeedItem, 0, len(items))
	for _, item := range items {
		if err := m.Enrich(item); err != nil {
			logrus.Error(err)
			//			continue
			// TODO (m0sth8): skip non enriched items?
		}
		results = append(results, item)
	}
	return results, nil
}

func (m *FeedManager) GetById(id bson.ObjectId) (*feed.FeedItem, error) {
	u := &feed.FeedItem{}
	return u, m.manager.GetById(m.col, id, &u)
}

func (m *FeedManager) FilterBy(f *FeedItemFltr, opts ...Opts) ([]*feed.FeedItem, int, error) {
	query := fltr.GetQuery(f)
	return m.FilterByQuery(query, opts...)
}

func (m *FeedManager) FilterByQuery(query bson.M, opts ...Opts) ([]*feed.FeedItem, int, error) {
	results := []*feed.FeedItem{}
	count, err := m.manager.FilterBy(m.col, &query, &results, opts...)
	if err == nil && count > 0 {
		results, err = m.EnrichMulti(results)
	}
	return results, count, err
}

func (m *FeedManager) Create(raw *feed.FeedItem) (*feed.FeedItem, error) {
	// TODO (m0sth8): add validation
	raw.Id = bson.NewObjectId()
	raw.Created = time.Now().UTC()
	raw.Updated = raw.Created
	if err := m.col.Insert(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (m *FeedManager) Update(obj *feed.FeedItem) error {
	obj.Updated = time.Now().UTC()
	return m.col.UpdateId(obj.Id, obj)
}

func (m *FeedManager) Remove(obj *feed.FeedItem) error {
	return m.col.RemoveId(obj.Id)
}

func (m *FeedManager) AddScan(sc *scan.Scan) (*feed.FeedItem, error) {
	feedItem := feed.FeedItem{
		Type:    feed.TypeScan,
		Project: sc.Project,
		Target:  sc.Target,
		ScanId:  sc.Id,
		Owner:   sc.Owner,
	}
	return m.Create(&feedItem)
}

func (m *FeedManager) UpdateScan(sc *scan.Scan) error {
	query := bson.M{
		"type":    feed.TypeScan,
		"project": sc.Project,
		"target":  sc.Target,
		"scanid":  sc.Id,
	}
	now := time.Now().UTC()
	update := bson.M{"$set": bson.M{"updated": now}}
	return m.col.Update(query, update)
}
