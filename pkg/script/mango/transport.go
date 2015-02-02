package mango

import (
	"encoding/json"
	"sync"

	"code.google.com/p/go.net/context"
	"github.com/Sirupsen/logrus"
	"github.com/facebookgo/stackerr"
	"github.com/gdamore/mangos"
	"github.com/gdamore/mangos/protocol/pair"
	"github.com/gdamore/mangos/transport/ipc"

	"fmt"
	"github.com/bearded-web/bearded/pkg/agent/message"
)

type OnRequest func(context.Context, *message.Message) (interface{}, error)

type Transport struct {
	sock      mangos.Socket
	sessions  map[int]chan<- *message.Message
	sRw       sync.Mutex
	onRequest OnRequest
}

func newListener(addr string) (*Transport, error) {
	sock, err := newPairSock()
	if err != nil {
		return nil, err
	}
	err = sock.Listen(addr)
	if err != nil {
		return nil, err
	}
	return newTransport(sock), nil
}

func newDialer(addr string) (*Transport, error) {
	sock, err := newPairSock()
	if err != nil {
		return nil, err
	}
	err = sock.Dial(addr)
	if err != nil {
		return nil, err
	}
	return newTransport(sock), nil
}

func newTransport(sock mangos.Socket) *Transport {
	return &Transport{
		sock:     sock,
		sessions: map[int]chan<- *message.Message{},
	}
}

func newPairSock() (mangos.Socket, error) {
	sock, err := pair.NewSocket()
	if err != nil {
		return nil, stackerr.Newf("can't get new pair socket: %s", err.Error())
	}
	sock.AddTransport(ipc.NewTransport())
	if err != nil {
		return nil, err
	}
	return sock, nil
}

// don't forget to call stop
func (c *Transport) Stop() error {
	return c.sock.Close()
}

func (c *Transport) Serve(ctx context.Context) error {
loop:
	for {
		msg := &message.Message{}
		select {
		case <-ctx.Done():
			break loop
		case err := <-c.asyncRecv(msg):
			if err != nil {
				logrus.Error(err)
				continue loop
			}
			if msg.Cmd == message.CmdResponse || msg.Cmd == message.CmdError {
				c.response(ctx, msg)
				continue loop
			}
			if c.onRequest != nil {
				go func(msg *message.Message) {
					resp := &message.Message{
						Id:  msg.Id,
						Cmd: message.CmdResponse,
					}
					respObj, err := c.onRequest(ctx, msg)
					if err == nil {
						err = resp.SetData(respObj)
					}
					if err != nil {
						logrus.Error(err)
						resp.SetData(err.Error())
						resp.Cmd = message.CmdError
					}
					err = c.send(resp)
					if err != nil {
						logrus.Error(stackerr.Wrap(err))
					}
				}(msg)
				continue loop
			}
		}
	}
	return nil
}

func (c *Transport) Request(ctx context.Context, cmd message.Cmd, send, recv interface{}) error {
	msg, err := message.New(cmd, send)
	if err != nil {
		return err
	}
	ch := c.addSession(msg.Id)
	//	defer c.pickSession(msg.Id)
	err = c.send(msg)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case msg := <-ch:
		if msg.Cmd == message.CmdResponse {
			return msg.GetData(recv)
		}
		if msg.Cmd == message.CmdError {
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

func (c *Transport) OnRequest(f OnRequest) {
	c.onRequest = f
}

func (c *Transport) response(ctx context.Context, msg *message.Message) {
	ch := c.pickSession(msg.Id)
	if ch == nil {
		return
	}
	select {
	case <-ctx.Done():
	case ch <- msg:
	}
	return
}

func (c *Transport) pickSession(id int) chan<- *message.Message {
	c.sRw.Lock()
	defer c.sRw.Unlock()
	ch, ok := c.sessions[id]
	if !ok {
		return nil
	}
	delete(c.sessions, id)
	return ch
}

func (c *Transport) addSession(id int) <-chan *message.Message {
	ch := make(chan *message.Message, 1)
	c.sRw.Lock()
	defer c.sRw.Unlock()
	c.sessions[id] = ch
	return ch
}

func (c *Transport) send(obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return stackerr.Wrap(err)
	}
	return stackerr.Wrap(c.sock.Send(data))
}

func (c *Transport) recv(obj interface{}) error {
	data, err := c.sock.Recv()
	println(string(data))
	if err != nil {
		return stackerr.Wrap(err)
	}
	return stackerr.Wrap(json.Unmarshal(data, obj))
}

func (c *Transport) asyncRecv(obj interface{}) <-chan error {
	ch := make(chan error, 1)
	go func(ch chan<- error) {
		ch <- c.recv(obj)
	}(ch)
	return ch
}
