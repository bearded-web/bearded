package scheduler

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/bearded-web/bearded/models/scan"
	"github.com/bearded-web/bearded/pkg/manager"
)

type Scheduler interface {
	// this method blocks until context is done or returns pack of jobs
	//	GetJobs(context.Context, *agent.Agent) ([]*agent.Job, error)

	AddScan(*scan.Scan) error
	GetSession() (*scan.Session, error)
	UpdateScan(*scan.Scan) error
}

type MemoryScheduler struct {
	mgr   *manager.Manager
	scans map[string]*scan.Scan
	rw    sync.RWMutex
}

var _ Scheduler = &MemoryScheduler{} // check interface compatibility

// Memory scheduler is just a prototype of scheduler, it mustn't be used in production environment
func NewMemoryScheduler(mgr *manager.Manager) *MemoryScheduler {
	return &MemoryScheduler{
		scans: map[string]*scan.Scan{},
		mgr:   mgr,
	}
}

//func (s *MemoryScheduler) GetJobs(ctx context.Context, agnt *agent.Agent) ([]*agent.Job, error) {
//
//}

func (s *MemoryScheduler) AddScan(sc *scan.Scan) error {
	return s.UpdateScan(sc)
}

func (s *MemoryScheduler) UpdateScan(sc *scan.Scan) error {
	s.rw.Lock()
	s.scans[s.mgr.FromId(sc.Id)] = sc
	s.rw.Unlock()
	return nil
}

func (s *MemoryScheduler) GetSession() (*scan.Session, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

scans:
	for id, sc := range s.scans {
	sessions:
		for _, sess := range sc.Sessions {
			switch sess.Status {

			case scan.StatusCreated:
				sess.Status = scan.StatusQueued
				err := s.mgr.Scans.UpdateSession(sc, sess)
				if err != nil {
					if s.mgr.IsNotFound(err) {
						delete(s.scans, id)
						continue scans
					}
					logrus.Error(err)
					continue scans
				}
				return sess, nil
			case scan.StatusQueued:
				// all scans session run in sequence order
				continue scans
			case scan.StatusFinished:
				// go to the next session
				continue sessions
			case scan.StatusWorking:
				// this scan is still working go to the next one
				continue scans
			case scan.StatusPaused:
				continue scans
			case scan.StatusFailed:
				delete(s.scans, id)
				continue scans
			}
		}
		// it looks like all session is finished, delete scan from queue
		delete(s.scans, id)
	}
	return nil, nil
}
