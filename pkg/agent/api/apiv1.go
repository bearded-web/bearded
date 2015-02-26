//go:generate stringer -type Method
package api

import (
	"github.com/bearded-web/bearded/models/plan"
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
	DownloadFile
)

type RequestV1 struct {
	Method Method

	GetPluginVersions string
	RunPlugin         *plan.WorkflowStep
	SendReport        *report.Report
	DownloadFile      string
}

type ResponseV1 struct {
	GetConfig *plan.Conf

	GetPluginVersions []string
	RunPlugin         *report.Report
	DownloadFile      []byte
}
