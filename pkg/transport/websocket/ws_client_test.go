package websocket

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"net/http/httptest"

	"code.google.com/p/go.net/context"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	"github.com/bearded-web/bearded/pkg/transport"
)

func TestClientLoop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	sIn := make(chan *transport.Message)
	sOut := make(chan *transport.Message)

	serv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		upgrader := websocket.Upgrader{}
		ws, err := upgrader.Upgrade(resp, req, http.Header{})
		if err != nil {
			http.Error(resp, "Application error", http.StatusInternalServerError)
			return
		}
		defer ws.Close()
		handleConnection(ctx, ws, sIn, sOut)
	}))

	addr := fmt.Sprintf("ws://%s", serv.URL[len("http://"):])
	s := &client{
		addr:   addr,
		dialer: &websocket.Dialer{},
	}
	in := make(chan *transport.Message)
	out := make(chan *transport.Message)

	ch := s.Loop(ctx, in, out)
	//	s.Loop(ctx, in, out)
	time.Sleep(1 * time.Millisecond)

	msg, err := transport.NewMessage(transport.CmdRequest, "data")
	require.NoError(t, err)
	sOut <- msg
	require.NoError(t, err)

	recvdMsg := <-in
	require.Equal(t, msg, recvdMsg)

	out <- msg
	recvdMsg = <-sIn
	require.NoError(t, err)
	require.Equal(t, msg, recvdMsg)

	cancel()
	err = <-ch
	require.NoError(t, err)

}
