package mango

import (
	"encoding/json"
	"testing"
	"time"

	"golang.org/x/net/context"
	"github.com/bearded-web/bearded/pkg/transport"
	"github.com/stretchr/testify/require"
)

func TestHandleConnection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	sock, err := newSock()
	require.NoError(t, err)
	err = sock.Listen(testAddr)
	require.NoError(t, err)
	defer sock.Close()

	in := make(chan *transport.Message)
	out := make(chan *transport.Message)

	ch := make(chan error)
	go func() {
		defer close(ch)
		ch <- handleConnection(ctx, sock, in, out)
	}()

	client, err := newSock()
	require.NoError(t, err)
	err = client.Dial(testAddr)
	require.NoError(t, err)

	msg, err := transport.NewMessage(transport.CmdRequest, "data")
	require.NoError(t, err)
	data, err := json.Marshal(msg)
	require.NoError(t, err)
	client.Send(data)

	recvdMsg := <-in
	require.Equal(t, msg, recvdMsg)

	out <- msg
	data, err = client.Recv()
	require.NoError(t, err)
	recvdMsg = &transport.Message{}
	err = json.Unmarshal(data, recvdMsg)
	require.NoError(t, err)
	require.Equal(t, msg, recvdMsg)

	cancel()
	err = <-ch
	require.NoError(t, err)

}
