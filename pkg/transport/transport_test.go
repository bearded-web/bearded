package transport

import (
	"fmt"
	"testing"
	"time"

	"code.google.com/p/go.net/context"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

type MockLoop struct {
	In  chan<- *Message
	Out <-chan *Message
	Err <-chan error
}

func (l *MockLoop) Loop(ctx context.Context,
	in chan<- *Message, out <-chan *Message) <-chan error {
	l.In = in
	l.Out = out
	ch := make(chan error)
	l.Err = ch
	return ch
}

func TestLoopTransport(t *testing.T) {
	bg := context.Background()

	ctx, cancel := context.WithTimeout(bg, 100*time.Millisecond)
	defer cancel()

	loop := &MockLoop{}
	tr := NewLoopTransport(loop)

	pluginName := ""

	h := Handle(func(ctx context.Context, msg Extractor) (interface{}, error) {
		println("handle")
		if err := msg.Extract(&pluginName); err != nil {
			return nil, err
		}
		return []string{"0.0.2"}, nil
	})

	go func() {
		err := tr.Serve(ctx, h)
		fmt.Printf("err: %v", err)
	}()
	// wait for serve execution
	time.Sleep(1 * time.Millisecond)

	// test transport on request
	msg, err := NewMessage(CmdRequest, "barbudo/wappalyzer")
	require.NoError(t, err)

	loop.In <- msg
	resp := <-loop.Out
	spew.Dump(resp)
	require.NotNil(t, resp)
	require.Equal(t, CmdResponse, resp.Cmd)

	recvd := []string{}
	resp.GetData(&recvd)
	require.Equal(t, recvd, []string{"0.0.2"})

	// test transport request
	var outMsg *Message
	go func() {
		outMsg = <-loop.Out
		m, _ := NewMessage(CmdResponse, "recv")
		m.Id = outMsg.Id
		loop.In <- m

	}()
	recvStr := ""
	err = tr.Request(ctx, "send", &recvStr)
	require.NoError(t, err)
	require.Equal(t, "recv", recvStr)

	require.NotNil(t, outMsg)
	require.Equal(t, CmdRequest, outMsg.Cmd)
	outData := ""
	outMsg.GetData(&outData)
	require.Equal(t, "send", outData)

}
