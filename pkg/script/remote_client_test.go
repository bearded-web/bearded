package script

import (
	"testing"
	"time"

	"code.google.com/p/go.net/context"
	"github.com/bearded-web/bearded/models/scan"
	"github.com/bearded-web/bearded/pkg/agent/api"
	"github.com/bearded-web/bearded/pkg/transport"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockTransport struct {
	transport.Fake
	mock.Mock

	resp *api.ResponseV1
}

func (m *MockTransport) Request(ctx context.Context, send, recv interface{}) error {
	args := m.Called(ctx, send, recv)
	resp := recv.(*api.ResponseV1)
	respSet := args.Get(1).(*api.ResponseV1)
	*resp = *respSet
	return args.Error(0)
}

func TestRemoteClient(t *testing.T) {
	transp := &MockTransport{}

	client, err := NewRemoteClient(transp)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	actualCfg := &scan.ScanConf{Target: "target"}

	transp.Mock.On("Request", ctx,
		api.RequestV1{Method: api.GetConfig},
		&api.ResponseV1{}).Return(nil, &api.ResponseV1{GetConfig: actualCfg}).Once()

	conf, err := client.GetConfig(ctx)
	require.NoError(t, err)
	require.Equal(t, actualCfg, conf)

}
