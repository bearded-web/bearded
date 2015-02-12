//go:generate stringer -type Method
package api

import (
	"github.com/bearded-web/bearded/models/plan"
	"github.com/bearded-web/bearded/models/plugin"
	"github.com/bearded-web/bearded/models/report"
)

type Method int

const (
	Ping Method = iota
	Connect
	GetConfig
	GetPluginVersions
	RunPlugin
	SendReport
)

type RequestV1 struct {
	Method Method

	GetPluginVersions string
	RunPlugin         *plan.WorkflowStep
	SendReport        *report.Report
}

type ResponseV1 struct {
	GetConfig *plugin.Conf

	GetPluginVersions []string
	RunPlugin         *report.Report
}
