package email

import (
	"fmt"

	"gopkg.in/gomail.v1"
)

type MemoryBackend struct {
	messages chan *gomail.Message
}

// Just for testing purposes. Is not thread safe.
func NewMemoryBackend(max int) *MemoryBackend {
	return &MemoryBackend{
		messages: make(chan *gomail.Message, max),
	}
}

func (b *MemoryBackend) Send(m *gomail.Message) error {
	if b.messages == nil {
		return fmt.Errorf("MemoryBackend is closed")
	}
	select {
	case b.messages <- m:
		return nil
	default:
		return fmt.Errorf("MemoryBackend has maximum messages")
	}
}

func (b *MemoryBackend) Close() {
	b.messages = nil
}

func (b *MemoryBackend) Messages() <-chan *gomail.Message {
	return b.messages
}
