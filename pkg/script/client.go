package script

import (
	"golang.org/x/net/context"
	"github.com/bearded-web/bearded/models/plan"
	"github.com/bearded-web/bearded/models/report"
)

// Client helps to communicate with agent
type ClientV1 interface {
	GetConfig(ctx context.Context) (*plan.Conf, error)
	GetPlugin(ctx context.Context, name string) (*Plugin, error)
	RunPlugin(ctx context.Context, step *plan.WorkflowStep) (*report.Report, error)
	SendReport(ctx context.Context, rep *report.Report) error
	DownloadFile(ctx context.Context, fileId string) ([]byte, error)
}
