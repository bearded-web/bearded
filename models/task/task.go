package task

import (
	"github.com/bearded-web/bearded/models/report"
	"time"
)


type TaskType string

const (
	TypeDocker TaskType = "docker"
)

type StatusType string
type Callback func(*Task)

const (
	StatusCreated StatusType = "created"
	StatusQueued  StatusType = "queued"
	StatusWorking            = "working"
	StatusPaused             = "paused"
	StatusDone               = "done"
	StatusError              = "error"
)

type Docker struct {
	Image string   `json:"image"`
	Cmd   []string `json:"cmd"`
}

type Task struct {
	Id        string         `json:"id"`
	Type      TaskType       `json:"type"`
	Docker    *Docker        `json:"docker,omitempty"`
	State     *State         `json:"state,omitempty" form:"-"`
	Report    *report.Report `json:"report,omitempty" form:"-"`
	callbacks []Callback     `json:"-" form:"-"`
}

type State struct {
	Status StatusType `json:"status"`
	Created time.Time

}

func New() *Task {
	return &Task{
		State: &State{
			Status: StatusCreated,
		},
		callbacks: []Callback{},
	}
}

func (t *Task) SetStatus(status StatusType) {
	t.State.Status = status
	if t.callbacks != nil {
		for _, cb := range t.callbacks {
			go cb(t)
		}
	}
}

func (t *Task) OnStateChange(fn func(t *Task)) {
	t.callbacks = append(t.callbacks, fn)
}
