package websocket

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"golang.org/x/net/context"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	"github.com/bearded-web/bearded/pkg/transport"
)

func TestServerLoop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	addr := "127.0.0.1:62322"

	serv := &http.Server{Addr: addr}

	s := &server{
		server: serv,
	}
	in := make(chan *transport.Message)
	out := make(chan *transport.Message)

	//	ch := s.Loop(ctx, in, out)
	s.Loop(ctx, in, out)
	time.Sleep(1 * time.Millisecond)

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(fmt.Sprintf("ws://%s", addr), nil)
	require.NoError(t, err)
	defer conn.Close()

	msg, err := transport.NewMessage(transport.CmdRequest, "data")
	require.NoError(t, err)
	err = conn.WriteJSON(msg)
	require.NoError(t, err)

	recvdMsg := <-in
	require.Equal(t, msg, recvdMsg)

	out <- msg
	recvdMsg = &transport.Message{}
	err = conn.ReadJSON(recvdMsg)
	require.NoError(t, err)
	require.Equal(t, msg, recvdMsg)

	//	cancel()
	//	err = <-ch
	//	require.NoError(t, err)

}
