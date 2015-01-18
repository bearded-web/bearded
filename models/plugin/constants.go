package plugin

type PluginType string

func (t PluginType) String() string {
	return string(t)
}

// It's a hack to show custom type as string in swagger
func (t PluginType) MarshalJSON() ([]byte, error) {
	return []byte(t), nil
}

const (
	Util   PluginType = "util"
	Script            = "script"
)

// =========

type PluginWeight string

// It's a hack to show custom type as string in swagger
func (t PluginWeight) MarshalJSON() ([]byte, error) {
	return []byte(t), nil
}

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
