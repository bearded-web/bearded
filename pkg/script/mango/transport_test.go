package mango

import (
	"code.google.com/p/go.net/context"
	"encoding/json"
	"github.com/bearded-web/bearded/pkg/agent/message"
	"github.com/gdamore/mangos"
	"github.com/gdamore/mangos/protocol/pair"
	"github.com/gdamore/mangos/transport/ipc"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

const addr string = "ipc:///tmp/mango_test.socket"

func TestNewTransport(t *testing.T) {
	var (
		transport *Transport
		err       error
	)
	transport, err = newListener("127.0.0.1:233")
	require.Error(t, err)
	require.Nil(t, transport)

	transport, err = newDialer("127.0.0.1:233")
	require.Error(t, err)
	require.Nil(t, transport)

	transport, err = newListener(addr)
	require.NoError(t, err)
	require.NotNil(t, transport)
	transport.Stop()

	transport, err = newDialer(addr)
	require.NoError(t, err)
	require.NotNil(t, transport)
	transport.Stop()
}

func TestTransportSend(t *testing.T) {
	sock, transport := getSocketAndTransport(t, addr)
	defer sock.Close()
	defer transport.Stop()

	err := transport.send("message")
	require.NoError(t, err)
	data, err := sock.Recv()
	require.NoError(t, err)
	require.NotNil(t, data)
	require.Equal(t, "\"message\"", string(data))

	err = transport.send(make(chan int))
	require.Error(t, err) // json can't marshal channel

}

func TestTransportRecv(t *testing.T) {
	sock, transport := getSocketAndTransport(t, addr)
	defer sock.Close()
	defer transport.Stop()

	err := sock.Send([]byte("\"message\""))
	require.NoError(t, err)

	msg := ""
	err = transport.recv(&msg)
	require.NoError(t, err)
	require.Equal(t, "message", msg)

	err = sock.Send([]byte("bad json"))
	require.NoError(t, err)
	err = transport.recv(&msg)
	require.Error(t, err)
}

func TestTransportAsyncRecv(t *testing.T) {
	sock, transport := getSocketAndTransport(t, addr)
	defer sock.Close()
	defer transport.Stop()
	msg := ""
	ch := transport.asyncRecv(&msg)

	err := sock.Send([]byte("\"message\""))
	require.NoError(t, err)

	require.NoError(t, <-ch)
	require.Equal(t, "message", msg)

	err = sock.Send([]byte("bad json"))
	require.NoError(t, err)
	ch = transport.asyncRecv(&msg)
	require.Error(t, <-ch)
}

func TestTransportSessions(t *testing.T) {
	transport := getTransport(t, addr)
	defer transport.Stop()

	require.Len(t, transport.sessions, 0)

	chOut := transport.addSession(10)
	require.NotNil(t, chOut)
	require.Len(t, transport.sessions, 1)

	chIn := transport.pickSession(10)
	require.NotNil(t, chIn)
	require.Len(t, transport.sessions, 0)
	require.NotEqual(t, chIn, chOut)

	msg := &message.Message{}

	chIn <- msg
	require.Equal(t, msg, <-chOut)

	chIn = transport.pickSession(10)
	require.Nil(t, chIn)
	require.Len(t, transport.sessions, 0)
}

func TestTransportRequestResponse(t *testing.T) {
	sock, transport := getSocketAndTransport(t, addr)
	defer sock.Close()
	defer transport.Stop()
	bg := context.Background()

	ctx, cancel := context.WithTimeout(bg, 10*time.Millisecond)
	defer cancel()
	msgData := ""
	ch := make(chan error)
	go func() {
		for {
			select {
			case <-ctx.Done():
				break
			default:
				ch <- transport.Request(ctx, message.CmdRunPlugin, "message", &msgData)
			}
		}
	}()

	answer := func(cmd message.Cmd, set interface{}) {
		data, err := sock.Recv()
		require.NoError(t, err)
		msg := &message.Message{}
		err = json.Unmarshal(data, msg)
		require.NoError(t, err)
		require.Equal(t, message.CmdRunPlugin, msg.Cmd)
		msg.Cmd = cmd
		if set != nil {
			msg.SetData(set)
		}
		transport.response(ctx, msg)
	}

	// receive message and send response
	answer(message.CmdResponse, nil)

	err := <-ch
	require.NoError(t, err)
	require.Equal(t, "message", msgData)

	// ==== check error situations

	answer(message.CmdError, "ERROR_MESSAGE")
	err = <-ch
	require.Error(t, err)
	require.Contains(t, err.Error(), "ERROR_MESSAGE")

	// unsupported msg CMD response, can be only CmdResponse or CmdError
	answer(message.CmdRunPlugin, nil)
	err = <-ch
	require.Error(t, err)
	require.Contains(t, err.Error(), "Unspopported cmd CmdRunPlugin")

	// Bad error response format
	answer(message.CmdError, 1234) // error msg should be a string
	err = <-ch
	require.Error(t, err)
	require.Contains(t, err.Error(), "unmarshal")

	ctx, cancel2 := context.WithTimeout(bg, 10*time.Millisecond)
	defer cancel2()
	require.Error(t, transport.Request(ctx, message.CmdRunPlugin, make(chan int), make(chan int)))

	cancel2()
	require.Error(t, transport.Request(ctx, message.CmdRunPlugin, "message", &msgData))

}

func TestTransportOnRequest(t *testing.T) {
	sock, transport := getSocketAndTransport(t, addr)
	defer sock.Close()
	defer transport.Stop()
	bg := context.Background()

	ctx, cancel := context.WithTimeout(bg, 10*time.Millisecond)
	defer cancel()

	go transport.Serve(ctx)

	// check onRequest behaviour
	onRequest := func(recv, send interface{}) {
		transport.OnRequest(func(ctx context.Context, msg *message.Message) (interface{}, error) {
			println("blalba")
			if err := msg.GetData(recv); err != nil {
				return nil, err
			}
			return send, nil
		})
	}

	{
		pluginName := ""
		onRequest(&pluginName, []string{"0.0.2"})
		msg, err := message.New(message.CmdGetPluginVersions, "barbudo/wappalyzer")
		require.NoError(t, err)
		data, err := json.Marshal(msg)
		require.NoError(t, err)
		sock.Send(data)

		data, err = sock.Recv()
		require.NoError(t, err)
		require.Equal(t, pluginName, "barbudo/wappalyzer")
		msg = &message.Message{}
		err = json.Unmarshal(data, msg)
		require.NoError(t, err)
		require.Equal(t, message.CmdResponse, msg.Cmd)
		recvd := []string{}
		msg.GetData(&recvd)
		require.Equal(t, recvd, []string{"0.0.2"})

	}
}

func TestTransportServe(t *testing.T) {
	sock, transport := getSocketAndTransport(t, addr)
	defer sock.Close()
	defer transport.Stop()
	bg := context.Background()

	ctx, cancel := context.WithTimeout(bg, 1*time.Millisecond)
	defer cancel()
	ch := make(chan error)
	go func() {
		ch <- transport.Serve(ctx)
	}()
	msgData := ""
	go func() {
		ch <- transport.Request(ctx, message.CmdRunPlugin, "message", &msgData)
	}()

	// receive request and send response
	data, err := sock.Recv()
	require.NoError(t, err)
	msg := &message.Message{}
	err = json.Unmarshal(data, msg)
	require.NoError(t, err)
	msg.Cmd = message.CmdResponse
	data, err = json.Marshal(msg)
	require.NoError(t, err)
	sock.Send(data)
	//	sock.Send([]byte("blabla"))

	require.NoError(t, <-ch)
	require.NoError(t, <-ch)

}

func getSocket(t *testing.T, addr string) mangos.Socket {
	sock, err := pair.NewSocket()
	if err != nil {
		t.Fatal(err)
	}
	sock.AddTransport(ipc.NewTransport())
	err = sock.Listen(addr)
	if err != nil {
		t.Fatal(err)
	}
	return sock
}

func getTransport(t *testing.T, addr string) *Transport {
	transport, err := newDialer(addr)
	if err != nil {
		t.Fatal(err)
	}
	return transport
}

func getSocketAndTransport(t *testing.T, addr string) (mangos.Socket, *Transport) {
	return getSocket(t, addr), getTransport(t, addr)
}

func TestMain(m *testing.M) {
	os.Remove(addr[6:])
	os.Exit(m.Run())
}
