package email

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsoleBackend(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	c := NewConsoleBackend()
	c.SetOutput(buf)
	msg := NewMessage()
	msg.SetBody("text/html", "Hello <b>Bob</b> and <i>Cora</i>!")
	err := c.Send(msg)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Hello <b>Bob</b> and <i>Cora</i>!")
}
