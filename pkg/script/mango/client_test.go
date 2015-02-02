package mango

import (
	"testing"
	"time"

	"code.google.com/p/go.net/context"
	"github.com/bearded-web/bearded/pkg/agent/message"
	"github.com/bearded-web/bearded/pkg/script"
	"github.com/stretchr/testify/require"
)

func TestClientGetPluginVersions(t *testing.T) {
	addr := "ipc:///tmp/mango_tests.socket"
	bg := context.Background()

	ctx, cancel := context.WithTimeout(bg, 10*time.Millisecond)
	defer cancel()

	serv, err := newListener(addr)
	defer serv.Stop()
	require.NoError(t, err)

	client, err := NewClient(addr)
	require.NoError(t, err)
	defer client.Stop()

	ch := make(chan struct {
		Pl  *script.Plugin
		Err error
	})
	go client.Serve(ctx)
	go func() {
		pl, err := client.GetPlugin(ctx, "barbudo/wappalyzer")
		v := struct {
			Pl  *script.Plugin
			Err error
		}{
			Pl:  pl,
			Err: err,
		}
		ch <- v
	}()

	time.Sleep(1 * time.Millisecond)
	msg := &message.Message{}
	err = serv.recv(msg)
	require.NoError(t, err)
	m := message.Message{
		Id:  msg.Id,
		Cmd: message.CmdResponse,
	}
	m.SetData([]string{"0.0.2"})
	err = serv.send(m)
	require.NoError(t, err)

	data := <-ch
	require.NoError(t, data.Err)
	require.Len(t, data.Pl.Versions, 1)
	require.Equal(t, "0.0.2", data.Pl.Versions[0])

}
