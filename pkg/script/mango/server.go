package mango

import (
	"code.google.com/p/go.net/context"
	"fmt"
	"github.com/bearded-web/bearded/pkg/agent/message"
	"github.com/facebookgo/stackerr"
)

type MangoServer struct {
	*Transport
}

func NewServer(addr string) (*MangoServer, error) {
	transport, err := newListener(addr)
	if err != nil {
		return nil, err
	}
	server := &MangoServer{
		Transport: transport,
	}
	transport.OnRequest(server.onRequest)
	return server, nil
}

func (s *MangoServer) onRequest(ctx context.Context, msg *message.Message) (interface{}, error) {
	switch msg.Cmd {
	case message.CmdGetPluginVersions:
		name := ""
		if err := msg.GetData(&name); err != nil {
			return nil, stackerr.Wrap(err)
		}
		return s.GetPluginVersions(name)
	default:
		return nil, fmt.Errorf("Unknown message cmd %s", msg.Cmd)
	}
}

func (s *MangoServer) GetPluginVersions(name string) ([]string, error) {
	return []string{"0.0.2"}, nil
}
