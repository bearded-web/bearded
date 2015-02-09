package scheduler

import "github.com/bearded-web/bearded/models/scan"

type Fake struct {
}

func NewFake() Scheduler {
	return &Fake{}
}

func (f *Fake) AddScan(*scan.Scan) error {
	return nil
}
func (f *Fake) GetSession() (*scan.Session, error) {
	return nil, nil
}
func (f *Fake) UpdateScan(*scan.Scan) error {
	return nil
}
