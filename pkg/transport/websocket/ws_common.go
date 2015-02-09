package websocket

import (
	"code.google.com/p/go.net/context"
	"github.com/Sirupsen/logrus"
	"github.com/bearded-web/bearded/pkg/transport"
	"github.com/gorilla/websocket"
)

func handleConnection(ctx context.Context, ws *websocket.Conn,
	in chan<- *transport.Message, out <-chan *transport.Message) {

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
			if err := ws.ReadJSON(msg); err != nil {
				chErr = err
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
			if err := ws.WriteJSON(msg); err != nil {
				chErr = err
				return
			}
		}

	}(ch)
	select {
	case <-ctx.Done():
		if ctx.Err() != context.Canceled {
			logrus.Error(ctx.Err())
		}
	case err := <-ch:
		logrus.Error(err)
	}
}
