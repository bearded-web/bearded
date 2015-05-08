package email

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryBackend(t *testing.T) {
	b := NewMemoryBackend(1)

	msg := NewMessage()
	msg.SetHeader("From", "alex@example.com")
	msg.SetHeader("To", "bob@example.com", "cora@example.com")
	msg.SetAddressHeader("Cc", "dan@example.com", "Dan")
	msg.SetHeader("Subject", "Hello!")
	msg.SetBody("text/html", "Hello <b>Bob</b> and <i>Cora</i>!")
	err := b.Send(msg)
	assert.NoError(t, err)
	err = b.Send(msg)
	assert.Error(t, err)
	assert.EqualError(t, err, "MemoryBackend has maximum messages")

	select {
	case msg2 := <-b.Messages():
		assert.Equal(t, msg2, msg)
	case <-time.After(time.Millisecond * 10):
		t.Fatal("Timeout exceeded")
	}

	b.Close()
	err = b.Send(msg)
	assert.Error(t, err)
	assert.EqualError(t, err, "MemoryBackend is closed")
}
