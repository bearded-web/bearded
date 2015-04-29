package email

import (
	"github.com/Sirupsen/logrus"
	"gopkg.in/gomail.v1"
)

type ConsoleBackend struct {
}

func NewConsoleBackend() *ConsoleBackend {
	return &ConsoleBackend{}
}

func (b *ConsoleBackend) Send(m *gomail.Message) error {
	msg := m.Export()
	logrus.Infof("%v\n", msg)
	return nil
}

func (b *ConsoleBackend) Close() {

}
