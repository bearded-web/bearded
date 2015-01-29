package agent

import "encoding/json"

type Status string

const (
	StatusUndefined  Status = ""           // Agent has undefined state if it isn't registered yet in system
	StatusRegistered Status = "registered" // Agent registered in system, but doesn't approved
	StatusApproved   Status = "approved"
	StatusBlocked    Status = "blocked" // agent was blocked by some reasons
)

var statuses = []interface{}{
	StatusUndefined,
	StatusRegistered,
	StatusApproved,
	StatusBlocked,
}

type Type string

const (
	System Type = "system"
)

var types = []interface{}{
	System,
}

// It's a hack to show custom type as string in swagger
func (t Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t Status) Enum() []interface{} {
	return statuses
}

func (t Status) Convert(text string) (interface{}, error) {
	return Status(text), nil
}

// It's a hack to show custom type as string in swagger
func (t Type) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t Type) Enum() []interface{} {
	return types
}

func (t Type) Convert(text string) (interface{}, error) {
	return Type(text), nil
}
