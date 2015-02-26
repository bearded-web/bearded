package script

import (
	"fmt"

	"code.google.com/p/go.net/context"
	"github.com/bearded-web/bearded/models/plan"
	"github.com/bearded-web/bearded/models/report"
)

// fake client helps to make a mock objects if you don't want implement all methods ClientV1 interface
type FakeClient struct {
}

// Check compile time interface compatibilities
var _ ClientV1 = NewFakeClient()

func NewFakeClient() *FakeClient {
	return &FakeClient{}
}

func (f *FakeClient) GetPlugin(ctx context.Context, name string) (*Plugin, error) {
	return &Plugin{client: f, Name: name}, nil
}

func (f *FakeClient) RunPlugin(ctx context.Context, step *plan.WorkflowStep) (*report.Report, error) {
	return nil, fmt.Errorf("No reports")
}

func (f *FakeClient) SendReport(ctx context.Context, rep *report.Report) error {
	return nil
}

func (f *FakeClient) GetConfig(ctx context.Context) (*plan.Conf, error) {
	return nil, nil
}

func (f *FakeClient) DownloadFile(ctx context.Context, fileId string) ([]byte, error) {
	return nil, nil
}
