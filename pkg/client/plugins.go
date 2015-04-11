package client

import (
	"golang.org/x/net/context"
	"github.com/bearded-web/bearded/models/plugin"
)

const pluginsUrl = "plugins"

type PluginsService struct {
	client *Client
}

func (s *PluginsService) String() string {
	return Stringify(s)
}

type PluginsListOpts struct {
	Name    string `url:"name"`
	Version string `url:"version"`
	Type    string `url:"type"`
}

// List plugins.
//
//
func (s *PluginsService) List(ctx context.Context, opt *PluginsListOpts) (*plugin.PluginList, error) {
	pluginList := &plugin.PluginList{}
	return pluginList, s.client.List(ctx, pluginsUrl, opt, pluginList)
}

func (s *PluginsService) Get(ctx context.Context, id string) (*plugin.Plugin, error) {
	plugin := &plugin.Plugin{}
	return plugin, s.client.Get(ctx, pluginsUrl, id, plugin)
}

func (s *PluginsService) Create(ctx context.Context, src *plugin.Plugin) (*plugin.Plugin, error) {
	pl := &plugin.Plugin{}
	err := s.client.Create(ctx, pluginsUrl, src, pl)
	if err != nil {
		return nil, err
	}
	return pl, nil
}

func (s *PluginsService) Update(ctx context.Context, src *plugin.Plugin) (*plugin.Plugin, error) {
	pl := &plugin.Plugin{}
	id := FromId(src.Id)
	err := s.client.Update(ctx, pluginsUrl, id, src, pl)
	if err != nil {
		return nil, err
	}
	return pl, nil
}
