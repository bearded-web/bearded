package script

import (
	"code.google.com/p/go.net/context"
	"github.com/bearded-web/bearded/models/plugin"
	"github.com/bearded-web/bearded/models/report"
)

// Client helps to communicate with agent
type ClientV1 interface {
	GetPlugin(ctx context.Context, name string) (*Plugin, error)
	RunPlugin(ctx context.Context, name, version string, conf *plugin.Conf) (*report.Report, error)
	SendReport(ctx context.Context, rep *report.Report) error
}
