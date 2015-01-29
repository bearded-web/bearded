package plugin

import "encoding/json"

type PluginType string

func (t PluginType) String() string {
	return string(t)
}

// It's a hack to show custom type as string in swagger
func (t PluginType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

const (
	Util   PluginType = "util"
	Script PluginType = "script"
)

// =========

type PluginWeight string

// It's a hack to show custom type as string in swagger
func (t PluginWeight) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

const (
	Light  PluginWeight = "light"
	Middle PluginWeight = "middle"
	Heavy  PluginWeight = "heavy"
)

// =========

type Dependence string

const (
	Blocking  Dependence = "blocking"  // plugin will not run if dependency doesn't exist
	Important Dependence = "important" // plugin will run with warnings
	Optional  Dependence = "optional"  // plugin will run with info messages
)

// =========

type LinkType string

const (
	Input    LinkType = "input"    // take output from previous plugin execution
	Output   LinkType = "output"   // send output from current plugin execution to the next plugin
	Parallel LinkType = "parallel" // plugins communicates with each other
)
