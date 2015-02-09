package plugin

import (
	"fmt"
	"time"

	"github.com/bearded-web/bearded/pkg/pagination"
	"gopkg.in/mgo.v2/bson"
)

type Container struct {
	Registry string `json:"registry"` // use public if empty
	Image    string `json:"image"`
}

type Desc struct {
	Title string `json:"title"` // human readable name
	Info  string `json:"info"`
	Url   string `json:"url"`
}

type Required struct {
	Plugin     string     `json:"plugin"`                                                      // plugin id, ex: "barbudo/wpscan"
	Versions   []string   `json:"versions"`                                                    // compatible versions
	Dependence Dependence `json:"dependence" description:"one of blocking|important|optional"` // with blocking dependency plugin will not run
}

type Conf struct {
	CommandArgs string `json:"commandArgs,omitempty" description:"passed to command line for plugins with type:util"`
	Target      string
}

type Plugin struct {
	Id        bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	Name      string        `json:"name" description:"unique plugin id, ex: barbudo/wpscan"` // TODO: do we need aliases or tags?
	Version   string        `json:"version"`
	Type      PluginType    `json:"type" description:"one of: util|script"`
	Weight    PluginWeight  `json:"weight" description:"one of: light|middle|heavy"`
	Desc      *Desc         `json:"desc" description:"human readable description"`
	Container *Container    `json:"container,omitempty" description:"information about container"`
	Created   time.Time     `json:"created,omitempty" description:"when plugin is created"`
	Updated   time.Time     `json:"updated,omitempty" description:"when plugin is updated"`

	//	Requirements []*Required   `json:"requirements,omitempty" description:"other plugins required for running"`
	Enabled bool `json:"enabled" description:"is plugin enabled for running"`
	// experimental

	//	Links []*Link `json:"links,omitempty"`
}

// Short description of plugin
func (p *Plugin) String() string {
	var str string
	if p.Id != "" {
		str = fmt.Sprintf("%x - %s v.%s", string(p.Id), p.Name, p.Version)
	} else {
		str = fmt.Sprintf("%s v.%s", p.Name, p.Version)
	}
	if p.Id != "" && !p.Enabled {
		str = fmt.Sprintf("%s DISABLED", str)
	}
	return str
}

//type Link struct {
//	Type     LinkType `json:"type"`
//	Info     string   `json:"info"`
//	Plugin   string   `json:"plugin"`
//	Versions []string `json:"versions"`
//}

type PluginList struct {
	pagination.Meta `json:",inline"`
	Results         []*Plugin `json:"results"`
}
