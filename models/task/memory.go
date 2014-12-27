package task

import (
	"github.com/bearded-web/bearded/modules/utils"
	"sync"
)

type MemoryManager struct {
	tasks map[string]*Task
	rw    sync.RWMutex
}

func NewMemoryManager() *MemoryManager {
	return &MemoryManager{
		tasks: map[string]*Task{},
	}
}

func (a *MemoryManager) Add(t *Task) error {
	t.Id = utils.UuidV4String()
	a.rw.Lock()
	a.tasks[t.Id] = t
	a.rw.Unlock()
	return nil
}

func (a *MemoryManager) Get(id string) (*Task, error) {
	a.rw.RLock()
	defer a.rw.RUnlock()
	if t, ok := a.tasks[id]; !ok {
		return nil, nil
	} else {
		return t, nil
	}
}

func (a *MemoryManager) Delete(id string) error {
	a.rw.Lock()
	defer a.rw.Unlock()
	delete(a.tasks, id)
	return nil
}

func (a *MemoryManager) All() ([]*Task, error) {
	tasks := []*Task{}
	a.rw.RLock()
	defer a.rw.RUnlock()
	for _, t := range a.tasks {
		tasks = append(tasks, t)
	}
	return tasks, nil
}
