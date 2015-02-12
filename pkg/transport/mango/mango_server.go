package mango

import (
	"crypto/tls"

	"code.google.com/p/go.net/context"
	"github.com/bearded-web/bearded/pkg/transport"
	"github.com/gdamore/mangos"
)

type server struct {
	addr string
	sock mangos.Socket
}

func NewServer(addr string) (transport.Transport, error) {
	sock, err := newSock()
	if err != nil {
		return nil, err
	}
	s := server{
		addr: addr,
		sock: sock,
	}
	return transport.NewLoopTransport(&s), nil
}

func NewTlsServer(addr string, certFile, keyFile string) (transport.Transport, error) {
	sock, err := newSock()
	if err != nil {
		return nil, err
	}
	config := &tls.Config{}
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	sock.SetOption(mangos.OptionTlsConfig, config)
	s := server{
		addr: addr,
		sock: sock,
	}
	return transport.NewLoopTransport(&s), nil
}

func (s *server) Loop(ctx context.Context,
	in chan<- *transport.Message, out <-chan *transport.Message) <-chan error {
	ch := make(chan error, 1)

	go func(ch chan<- error) {
		defer close(ch)
		err := s.sock.Listen(s.addr)
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
