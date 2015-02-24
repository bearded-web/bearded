package agent

import (
	"testing"
	"time"

	"code.google.com/p/go.net/context"
	"github.com/bearded-web/bearded/models/plan"
	"github.com/bearded-web/bearded/models/plugin"
	"github.com/bearded-web/bearded/models/scan"
	"github.com/bearded-web/bearded/pkg/agent/api"
	"github.com/bearded-web/bearded/pkg/transport"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockTransport struct {
	transport.Fake
	mock.Mock
}

func (m *MockTransport) Request(ctx context.Context, send, recv interface{}) error {
	args := m.Called(ctx, send, recv)
	resp := recv.(*api.ResponseV1)
	respSet := args.Get(1).(*api.ResponseV1)
	*resp = *respSet
	return args.Error(0)
}

func TestServerMethods(t *testing.T) {
	transp := &MockTransport{}

	sess := &scan.Session{
		Step: &plan.WorkflowStep{
			Conf: &plan.Conf{
				CommandArgs: "args",
			},
		},
	}

	serv, err := NewRemoteServer(transp, nil, sess)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	actualCfg := &plan.Conf{}

	transp.Mock.On("Request", ctx,
		api.RequestV1{Method: api.GetConfig},
		&api.ResponseV1{}).Return(nil, &api.ResponseV1{GetConfig: actualCfg}).Once()

	serv.GetConfig(ctx)

}
