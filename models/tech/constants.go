package tech

import "encoding/json"

type ActivityType string

const (
	ActivityReported  = ActivityType("reported")  // the issue was reported by plugin
	ActivityConfirmed = ActivityType("confirmed") // the issue was confirmed by someone
	ActivityFalse     = ActivityType("false")     // set to false
	ActivityTrue      = ActivityType("true")      // set to true
)

var activities = []interface{}{
	ActivityReported,
	ActivityConfirmed,
	ActivityFalse,
	ActivityTrue,
}

// It's a hack to show custom type as string in swagger
func (t ActivityType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t ActivityType) Enum() []interface{} {
	return activities
}

func (t ActivityType) Convert(text string) (interface{}, error) {
	return ActivityType(text), nil
}
