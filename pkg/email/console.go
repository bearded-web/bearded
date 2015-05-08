package email

import (
	"fmt"
	"io"

	"github.com/Sirupsen/logrus"
	"gopkg.in/gomail.v1"
)

type logWriter struct{}

func (w *logWriter) Write(data []byte) (int, error) {
	logrus.Debug(string(data))
	return len(data), nil
}

type ConsoleBackend struct {
	output io.Writer
}

func NewConsoleBackend() *ConsoleBackend {
	return &ConsoleBackend{
		output: &logWriter{},
	}
}

func (b *ConsoleBackend) Send(m *gomail.Message) error {
	msg := m.Export()
	b.output.Write([]byte(fmt.Sprintf("%v\n", msg)))
	return nil
}

func (b *ConsoleBackend) Close() {

}

func (b *ConsoleBackend) SetOutput(w io.Writer) {
	b.output = w
}
