package docker

type Executor struct {
	docker *Docker
}

func NewExecutor(docker *Docker) *Executor {
	return &Executor{
		docker: docker,
	}
}

//func (e *Executor)
