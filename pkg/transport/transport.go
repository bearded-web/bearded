package transport

import (
	"fmt"

	"code.google.com/p/go.net/context"
	"github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
)

type Looper interface {
	Loop(ctx context.Context, in chan<- *Message, out <-chan *Message) <-chan error
}

type LoopTransport struct {
	session *Session
	out     chan *Message
	loop    Looper
}

// check interface
var _ Transport = (*LoopTransport)(nil)

func NewLoopTransport(loop Looper) Transport {
	s := LoopTransport{
		session: NewSession(),
		out:     make(chan *Message),
		loop:    loop,
	}
	return &s
}

func (s *LoopTransport) Serve(ctx context.Context, h Handler) error {
	in := make(chan *Message)
	ch := s.loop.Loop(ctx, in, s.out)
loop:
	for {
		select {
		// error from outside
		case <-ctx.Done():
			if ctx.Err() == context.Canceled {
				break loop
			}
			return ctx.Err()
		// error from looper
		case err := <-ch:
			return err
		// somebody send data to us
		case msg := <-in:
			if msg.Cmd == CmdResponse || msg.Cmd == CmdError {
				s.response(ctx, msg)
				continue loop
			}
			if msg.Cmd != CmdRequest {
				continue loop
			}
			if h != nil {
				go func(msg *Message) {
					resp := &Message{
						Id:  msg.Id,
						Cmd: CmdResponse,
					}
					respObj, err := h.Handle(ctx, msg)
					if err == nil {
						err = resp.SetData(respObj)
					}
					if err != nil {
						logrus.Error(err)
						resp.SetData(err.Error())
						resp.Cmd = CmdError
					}
					if err != s.send(ctx, resp) {
						logrus.Error(err)
					}

				}(msg)
				continue loop
			}
		}

	}
	return nil
}

func (s *LoopTransport) Request(ctx context.Context, send, recv interface{}) error {
	msg, err := NewMessage(CmdRequest, send)
	if err != nil {
		return err
	}
	ch := s.session.Add(msg.Id)
	logrus.Debug("send request", spew.Sdump(msg))
	if err := s.send(ctx, msg); err != nil {
		return err
	}
	logrus.Debug("wait for response")
	// wait for response
	select {
	case <-ctx.Done():
		return ctx.Err()
	case msg := <-ch:
		if msg.Cmd == CmdResponse {
			return msg.GetData(recv)
		}
		if msg.Cmd == CmdError {
			retErr := ""
			err = msg.GetData(&retErr)
			if err != nil {
				return fmt.Errorf("Can't get error response: %v", err)
			}
			return fmt.Errorf(retErr)
		}
		return fmt.Errorf("Unspopported cmd %v", msg.Cmd)
	}
}

func (s *LoopTransport) send(ctx context.Context, msg *Message) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.out <- msg:
	}
	return nil
}

func (s *LoopTransport) response(ctx context.Context, msg *Message) {
	ch := s.session.Pick(msg.Id)
	if ch == nil {
		return
	}
	select {
	case <-ctx.Done():
	case ch <- msg:
	}
	return
}
