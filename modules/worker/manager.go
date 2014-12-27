package worker

import (
	"code.google.com/p/go.net/context"
	"github.com/bearded-web/bearded/models/task"
	"github.com/bearded-web/bearded/modules/docker"
)

type WorkerC struct {
	Worker *Worker
	Cancel context.CancelFunc
	Result <- chan error
}

type ManagerOpts struct {
	Size int
}

type Manager struct {
	opts    *ManagerOpts
	workers []*WorkerC
	queue   chan *task.Task
	docker  *docker.Docker
}

func NewManager(d *docker.Docker, opts *ManagerOpts) *Manager {
	return &Manager{
		workers: []*WorkerC{},
		opts:    opts,
		queue:   make(chan *task.Task, opts.Size),
		docker: d,
	}
}

func (m *Manager) Queue() chan<- *task.Task {
	return m.queue
}

func (m *Manager) startWorkers(ctx context.Context) {
	m.workers = []*WorkerC{}
	for i := 0; i < m.opts.Size; i++ {
		println("start worker", i)
		w := NewWorker(m.docker)
		wCtx, cancel := context.WithCancel(ctx)
		res := make(chan error)
		go func(res chan<- error){
			res <- w.Work(wCtx, m.queue)
		}(res) // wait for finishing
		m.workers = append(m.workers, &WorkerC{w, cancel, res})
	}
}

func (m *Manager) stopWorkers() {
	for _, wC := range m.workers {
		wC.Cancel()
		err := <- wC.Result
		if err != nil {
			println(err)
		}
	}
}

func (m *Manager) Serve(ctx context.Context) error {
	m.startWorkers(ctx)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	m.stopWorkers()
	return nil
}
