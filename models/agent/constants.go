package agent

import "encoding/json"

type Status string

// It's a hack to show custom type as string in swagger
func (t Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

const (
	Undefined  Status = ""           // Agent has undefined state if it isn't registered yet in system
	Registered        = "registered" // Agent registered in system, but doesn't approved
	Approved          = "approved"
	//	Waiting            = "waiting"
	//	Unavailable        = "unavailable"
	//	Paused             = "paused"
	Blocked = "blocked" // agent was blocked by some reasons
)

type Type string

// It's a hack to show custom type as string in swagger
func (t Type) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

const (
	System Type = "system"
)
