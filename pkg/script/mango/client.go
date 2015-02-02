package mango

import (
	"code.google.com/p/go.net/context"

	"github.com/bearded-web/bearded/pkg/agent/message"
	"github.com/bearded-web/bearded/pkg/script"
)

type MangoClient struct {
	*Transport

	script.FakeClient // TODO (m0sth8): remove when implement all methods
}

// Check compile time interface compatibilities
var _ script.ClientV1 = (*MangoClient)(nil)

func NewClient(addr string) (*MangoClient, error) {
	transport, err := newDialer(addr)
	if err != nil {
		return nil, err
	}
	return &MangoClient{
		Transport: transport,
	}, nil

}

func (c *MangoClient) GetPlugin(ctx context.Context, name string) (*script.Plugin, error) {
	versions := []string{}
	err := c.Request(ctx, message.CmdGetPluginVersions, name, &versions)
	if err != nil {
		return nil, err
	}
	return script.NewPlugin(name, c, versions...), nil

}
