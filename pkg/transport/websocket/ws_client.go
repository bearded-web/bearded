package websocket

import (
	"crypto/tls"

	"code.google.com/p/go.net/context"
	"github.com/bearded-web/bearded/pkg/transport"
	"github.com/gorilla/websocket"
)

type client struct {
	addr   string
	dialer *websocket.Dialer
}

func NewClient(addr string) transport.Transport {
	s := client{
		addr:   addr,
		dialer: &websocket.Dialer{},
	}
	return transport.NewLoopTransport(&s)
}

func NewTlsClient(addr string, certFile, keyFile string) (transport.Transport, error) {
	config := &tls.Config{
		NextProtos: []string{"http/1.1"},
	}
	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	s := client{
		addr: addr,
		dialer: &websocket.Dialer{
			TLSClientConfig: config,
		},
	}
	return transport.NewLoopTransport(&s), nil
}

func (s *client) Loop(ctx context.Context,
	in chan<- *transport.Message, out <-chan *transport.Message) <-chan error {
	ch := make(chan error, 1)

	go func(ch chan<- error) {
		defer close(ch)
		ws, _, err := s.dialer.Dial(s.addr, nil)
		if err != nil {
			ch <- err
			return
		}
		defer ws.Close()
		handleConnection(ctx, ws, in, out)
	}(ch)
	return ch
}
