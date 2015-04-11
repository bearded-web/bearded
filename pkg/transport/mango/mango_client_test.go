package mango

import (
	"encoding/json"
	"testing"
	"time"

	"golang.org/x/net/context"
	"github.com/bearded-web/bearded/pkg/transport"
	"github.com/stretchr/testify/require"
)

func TestMangoClientLoop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	sock, err := newSock()
	require.NoError(t, err)

	serv, err := newSock()
	require.NoError(t, err)
	err = serv.Listen(testAddr)
	require.NoError(t, err)
	defer serv.Close()

	s := &client{
		sock: sock,
		addr: testAddr,
	}
	in := make(chan *transport.Message)
	out := make(chan *transport.Message)

	ch := s.Loop(ctx, in, out)

	msg, err := transport.NewMessage(transport.CmdRequest, "data")
	require.NoError(t, err)
	data, err := json.Marshal(msg)
	require.NoError(t, err)
	serv.Send(data)

	recvdMsg := <-in
	require.Equal(t, msg, recvdMsg)

	out <- msg
	data, err = serv.Recv()
	require.NoError(t, err)
	recvdMsg = &transport.Message{}
	err = json.Unmarshal(data, recvdMsg)
	require.NoError(t, err)
	require.Equal(t, msg, recvdMsg)

	cancel()
	err = <-ch
	require.NoError(t, err)

}
