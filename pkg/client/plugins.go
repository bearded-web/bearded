package client

import (
	"fmt"
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
	Name    string
	Version string
	Type    string
}

// List plugins.
//
//
func (s *PluginsService) List(opt *PluginsListOpts) (*plugin.PluginList, error) {
	pluginList := &plugin.PluginList{}
	return pluginList, s.client.List(pluginsUrl, opt, pluginList)
}

func (s *PluginsService) Get(id string) (*plugin.Plugin, error) {
	plugin := &plugin.Plugin{}
	return plugin, s.client.Get(pluginsUrl, id, plugin)
}

func (s *PluginsService) Create(src *plugin.Plugin) (*plugin.Plugin, error) {
	pl := &plugin.Plugin{}
	err := s.client.Create(pluginsUrl, src, pl)
	if err != nil {
		return nil, err
	}
	return pl, nil
}

func (s *PluginsService) Update(src *plugin.Plugin) (*plugin.Plugin, error) {
	pl := &plugin.Plugin{}
	id := fmt.Sprintf("%x", string(src.Id))
	err := s.client.Update(pluginsUrl, id, src, pl)
	if err != nil {
		return nil, err
	}
	return pl, nil
}
