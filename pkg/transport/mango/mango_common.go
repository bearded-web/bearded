package mango

import (
	"encoding/json"

	"golang.org/x/net/context"
	"github.com/bearded-web/bearded/pkg/transport"
	"github.com/facebookgo/stackerr"
	"github.com/gdamore/mangos"
	"github.com/gdamore/mangos/protocol/pair"
	"github.com/gdamore/mangos/transport/tcp"
	"github.com/gdamore/mangos/transport/tlstcp"
)

func send(s mangos.Socket, obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return stackerr.Wrap(err)
	}
	return stackerr.Wrap(s.Send(data))
}

func recv(s mangos.Socket, obj interface{}) error {
	data, err := s.Recv()
	if err != nil {
		return stackerr.Wrap(err)
	}
	return stackerr.Wrap(json.Unmarshal(data, obj))
}

func newSock() (mangos.Socket, error) {
	sock, err := pair.NewSocket()
	if err != nil {
		return nil, stackerr.Newf("can't get new pair socket: %s", err.Error())
	}
	sock.AddTransport(tcp.NewTransport())
	sock.AddTransport(tlstcp.NewTransport())
	if err != nil {
		return nil, err
	}
	return sock, nil
}

func handleConnection(ctx context.Context, sock mangos.Socket,
	in chan<- *transport.Message, out <-chan *transport.Message) error {

	ch := make(chan error, 2)

	// start read loop
	go func(ch chan<- error) {
		var chErr error
		defer func() {
			if chErr != nil {
				ch <- chErr
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			msg := &transport.Message{}

			if err := recv(sock, msg); err != nil {
				chErr = stackerr.Wrap(err)
				return
			}
			select {
			case <-ctx.Done():
				return
			case in <- msg:
			}
		}

	}(ch)
	// start write loop
	go func(ch chan<- error) {
		var chErr error
		defer func() {
			if chErr != nil {
				ch <- chErr
			}
		}()
		for {
			var msg *transport.Message
			select {
			case <-ctx.Done():
				return
			case msg = <-out:
			}
			if err := send(sock, msg); err != nil {
				chErr = stackerr.Wrap(err)
				return
			}
		}

	}(ch)
	select {
	case <-ctx.Done():
		if err := ctx.Err(); err != context.Canceled {
			return err
		}
	case err := <-ch:
		return stackerr.Wrap(err)
	}
	return nil
}
