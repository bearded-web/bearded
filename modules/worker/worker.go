package worker

import (
	"github.com/bearded-web/bearded/models/task"
	"code.google.com/p/go.net/context"
	"github.com/bearded-web/bearded/modules/docker"
	"github.com/bearded-web/bearded/models/report"
)

type WorkerOpts struct {
}

type Worker struct {
	docker *docker.Docker
}

func NewWorker(d *docker.Docker) *Worker {
	return &Worker{
		docker: d,
	}
}

func (w *Worker) Work(ctx context.Context, queue <- chan *task.Task) error{
	for {
		select {
		case <- ctx.Done():
			return ctx.Err()
		case task := <- queue:
			w.HandleTask(ctx, task)
		}
	}
}

func (w *Worker) HandleTask(ctx context.Context, t *task.Task) {
	t.SetStatus(task.StatusWorking)
	if t.Type == task.TypeDocker {
		cfg := docker.ContainerConfig{
			Image: t.Docker.Image,
			Cmd: t.Docker.Cmd,
			Tty: true,

		}
		ch := w.docker.RunImage(ctx, &cfg)
		select {
		case <- ctx.Done():
		case res := <- ch:
			if res.Err != nil {
				t.SetStatus(task.StatusError)
				return
			}
			t.Report = &report.Report{
				Type: report.TypeRaw,
				Raw: string(res.Log),
			}
			t.SetStatus(task.StatusDone)
			return
		}
	}
	t.SetStatus(task.StatusError) // no handlers

}
