package plugin

import "encoding/json"

type PluginType string

const (
	Util   PluginType = "util"
	Script PluginType = "script"
)

var pluginTypes = []interface{}{
	Util,
	Script,
}

func (t PluginType) String() string {
	return string(t)
}

// It's a hack to show custom type as string in swagger
func (t PluginType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t PluginType) Enum() []interface{} {
	return pluginTypes
}

func (t PluginType) Convert(text string) (interface{}, error) {
	return PluginType(text), nil
}

// =========

type PluginWeight string

const (
	Light  PluginWeight = "light"
	Middle PluginWeight = "middle"
	Heavy  PluginWeight = "heavy"
)

var pluginWeights = []interface{}{
	Light,
	Middle,
	Heavy,
}

// It's a hack to show custom type as string in swagger
func (t PluginWeight) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t PluginWeight) Enum() []interface{} {
	return pluginWeights
}

func (t PluginWeight) Convert(text string) (interface{}, error) {
	return PluginWeight(text), nil
}

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
