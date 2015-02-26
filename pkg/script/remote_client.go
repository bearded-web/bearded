package script

import (
	"fmt"

	"code.google.com/p/go.net/context"
	"github.com/davecgh/go-spew/spew"

	"github.com/bearded-web/bearded/models/plan"
	"github.com/bearded-web/bearded/models/report"
	"github.com/bearded-web/bearded/pkg/agent/api"
	"github.com/bearded-web/bearded/pkg/transport"
)

type RemoteClient struct {
	transp transport.Transport

	FakeClient // TODO (m0sth8): remove when implement all methods
	connected  chan struct{}
}

// Check compile time interface compatibilities
var _ ClientV1 = (*RemoteClient)(nil)

// Remote client helps communicate plugins with agent through transport
func NewRemoteClient(transport transport.Transport) (*RemoteClient, error) {
	return &RemoteClient{
		transp:    transport,
		connected: make(chan struct{}, 1),
	}, nil
}

func (s *RemoteClient) Handle(ctx context.Context, msg transport.Extractor) (interface{}, error) {
	fmt.Printf("Handle msg", spew.Sdump(msg))
	req := api.RequestV1{}
	err := msg.Extract(&req)
	if err != nil {
		return nil, err
	}
	switch req.Method {
	case api.Connect:
		s.connected <- struct{}{}
		return nil, nil
	default:
		return nil, fmt.Errorf("Unknown method requested %s", req.Method)
	}
}

func (s *RemoteClient) WaitForConnection(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.connected:
		return nil
	}
}

// Client Methods

func (c *RemoteClient) GetConfig(ctx context.Context) (*plan.Conf, error) {
	req := api.RequestV1{
		Method: api.GetConfig,
	}
	resp := api.ResponseV1{}
	if err := c.transp.Request(ctx, req, &resp); err != nil {
		return nil, err
	}
	return resp.GetConfig, nil
}

func (c *RemoteClient) GetPlugin(ctx context.Context, name string) (*Plugin, error) {
	req := api.RequestV1{
		Method:            api.GetPluginVersions,
		GetPluginVersions: name,
	}
	resp := api.ResponseV1{}
	if err := c.transp.Request(ctx, req, &resp); err != nil {
		return nil, err
	}
	return NewPlugin(name, c, resp.GetPluginVersions...), nil
}

func (c *RemoteClient) RunPlugin(ctx context.Context, step *plan.WorkflowStep) (*report.Report, error) {
	req := api.RequestV1{
		Method:    api.RunPlugin,
		RunPlugin: step,
	}
	resp := api.ResponseV1{}
	if err := c.transp.Request(ctx, req, &resp); err != nil {
		return nil, err
	}
	return resp.RunPlugin, nil
}

func (c *RemoteClient) SendReport(ctx context.Context, rep *report.Report) error {
	req := api.RequestV1{
		Method:     api.SendReport,
		SendReport: rep,
	}
	resp := api.ResponseV1{}
	if err := c.transp.Request(ctx, req, &resp); err != nil {
		return err
	}
	return nil
}

func (c *RemoteClient) DownloadFile(ctx context.Context, fileId string) ([]byte, error) {
	req := api.RequestV1{
		Method:       api.DownloadFile,
		DownloadFile: fileId,
	}
	resp := api.ResponseV1{}
	if err := c.transp.Request(ctx, req, &resp); err != nil {
		return nil, err
	}
	return resp.DownloadFile, nil
}
