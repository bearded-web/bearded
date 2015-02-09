package transport

import (
	"sync"
)

type Session struct {
	sessions map[int]chan<- *Message
	sRw      sync.Mutex
}

func NewSession() *Session {
	return &Session{
		sessions: map[int]chan<- *Message{},
	}
}

func (s *Session) Pick(id int) chan<- *Message {
	s.sRw.Lock()
	defer s.sRw.Unlock()
	ch, ok := s.sessions[id]
	if !ok {
		return nil
	}
	delete(s.sessions, id)
	return ch
}

func (s *Session) Add(id int) <-chan *Message {
	ch := make(chan *Message, 1)
	s.sRw.Lock()
	defer s.sRw.Unlock()
	s.sessions[id] = ch
	return ch
}
