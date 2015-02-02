package script

import (
	"code.google.com/p/go.net/context"

	"github.com/bearded-web/bearded/models/plugin"
	"github.com/bearded-web/bearded/models/report"
)

type Plugin struct {
	Name     string
	Versions []string
	client   ClientV1
}

func NewPlugin(name string, client ClientV1, versions ...string) *Plugin {
	return &Plugin{
		Name:     name,
		client:   client,
		Versions: versions,
	}
}

func (p *Plugin) HasVersion(version string) bool {
	for _, version := range p.Versions {
		if version == version {
			return true
		}
	}
	return false
}

func (p *Plugin) LatestSupportedVersion(versions []string) string {
	for i := (len(versions) - 1); i >= 0; i-- {
		if p.HasVersion(versions[i]) {
			return versions[i]
		}
	}
	return ""
}

func (p *Plugin) LatestVersion() string {
	if p.Versions == nil || len(p.Versions) == 0 {
		return ""
	}
	return p.Versions[len(p.Versions)-1]
}

func (p *Plugin) Run(ctx context.Context, version string, conf *plugin.Conf) (*report.Report, error) {
	return p.client.RunPlugin(ctx, p.Name, version, conf)
}
