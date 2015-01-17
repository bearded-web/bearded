package plugin

type PluginType string

func (a PluginType) String() string {
	return string(a)
}

const (
	Util   PluginType = "util"
	Script            = "script"
)

// =========

type PluginWeight string

const (
	Light  PluginWeight = "light"
	Middle              = "middle"
	Heavy               = "heavy"
)

// =========

type Dependence string

const (
	Blocking  Dependence = "blocking"  // plugin will not run if dependency doesn't exist
	Important            = "important" // plugin will run with warnings
	Optional             = "optional"  // plugin will run with info messages
)

// =========

type LinkType string

const (
	Input    LinkType = "input"    // take output from previous plugin execution
	Output            = "output"   // send output from current plugin execution to the next plugin
	Parallel          = "parallel" // plugins communicates with each other
)
