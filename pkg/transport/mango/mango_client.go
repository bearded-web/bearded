package mango

import (
	"golang.org/x/net/context"
	"github.com/bearded-web/bearded/pkg/transport"
	"github.com/gdamore/mangos"
)

type client struct {
	addr string
	sock mangos.Socket
}

func NewClient(addr string) (transport.Transport, error) {
	sock, err := newSock()
	if err != nil {
		return nil, err
	}
	s := client{
		addr: addr,
		sock: sock,
	}
	return transport.NewLoopTransport(&s), nil
}

func (s *client) Loop(ctx context.Context,
	in chan<- *transport.Message, out <-chan *transport.Message) <-chan error {
	ch := make(chan error, 1)

	go func(ch chan<- error) {
		defer close(ch)
		err := s.sock.Dial(s.addr)
		if err != nil {
			ch <- err
			return
		}
		defer s.sock.Close()
		err = handleConnection(ctx, s.sock, in, out)
		if err != nil {
			ch <- err
			return
		}
	}(ch)
	return ch
}
