package task

type Manager interface {
	Add(*Task) error
	Get(id string) (*Task, error)
	Delete(id string) error
	All() ([]*Task, error)
}
